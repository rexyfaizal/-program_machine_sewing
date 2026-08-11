import { ref, watch } from "vue";
import {
  getLineShiftConfig,
  getMachineOperatorReport,
  getProductivity,
} from "../api/machineApi";
import {
  buildOperatorExportMap,
  buildRangeDetailRows,
  buildRangeSummaryBaseRows,
  buildRangeSummaryRows,
  buildUuidShiftMetricsMap,
  createRangeWorkbook,
  extractRows,
  filterRangeItemsByLocation,
  filterRangeItemsByStatus,
  formatExcelDate,
  getOrderedRange,
  getWeekdaysBetween,
  normalizeRangeItem,
  normalizeText,
  safeFileName,
  writeWorkbook,
} from "../utils/dashboardExportExcel";
import {
  buildLineShiftConfigMap,
  factoryHasEnabledShift,
} from "../utils/gm3Shift";

export function useDashboardExcelExport({
  selectedDate,
  locationFilter,
  statusFilter,
  machineSettings,
  showNotice,
  productivityShift,
  shiftConfigMap,
}) {
  const rangeExporting = ref(false);
  const startDate = ref(selectedDate.value || "");
  const endDate = ref(selectedDate.value || "");

  watch(
    selectedDate,
    (value) => {
      if (!value) return;
      if (!startDate.value) startDate.value = value;
      if (!endDate.value) endDate.value = value;
    },
    { immediate: true }
  );

  watch(startDate, (value) => {
    if (!endDate.value) {
      endDate.value = value;
    }
  });

  function getManualSettingByUuid(uuid) {
    const key = normalizeText(uuid);
    return machineSettings?.value?.get(key) || null;
  }

  function getLocationByUuid(uuid) {
    return getManualSettingByUuid(uuid)?.location || "";
  }

  function resolveExportShiftParam(configMap) {
    const area = String(locationFilter?.value || "").trim().toUpperCase();
    const shift = String(productivityShift?.value || "").trim();

    // All GM: tetap ikut filter shift (default CURRENT) agar GM3 tidak ALL-shifts.
    if (!area || area === "ALL") {
      return shift || "CURRENT";
    }

    if (!factoryHasEnabledShift(area, configMap)) {
      return "";
    }

    return shift;
  }

  async function downloadExcel() {
    if (rangeExporting.value) return;

    const range = getOrderedRange(
      startDate.value,
      endDate.value,
      selectedDate.value
    );

    if (!range.start || !range.end) {
      showNotice("Tanggal export belum lengkap.", "error");
      return;
    }

    const dates = getWeekdaysBetween(range.start, range.end);

    if (!dates.length) {
      showNotice("Range tanggal tidak memiliki hari kerja (Senin–Sabtu).", "error");
      return;
    }

    rangeExporting.value = true;
    showNotice(
      `Mengambil data ${formatExcelDate(range.start)} - ${formatExcelDate(
        range.end
      )}...`,
      "ok"
    );

    try {
      let configMap =
        shiftConfigMap?.value instanceof Map
          ? shiftConfigMap.value
          : new Map();

      if (!configMap.size) {
        try {
          const data = await getLineShiftConfig("");
          configMap = buildLineShiftConfigMap(
            Array.isArray(data?.lines) ? data.lines : []
          );
        } catch (err) {
          console.warn("Gagal load shift config untuk export:", err);
        }
      }

      const shiftParam = resolveExportShiftParam(configMap);

      const results = await Promise.all(
        dates.map(async (dateText) => {
          const selectedShift = String(shiftParam || "").trim().toUpperCase();
          const extraShiftCodes = ["SHIFT_1", "SHIFT_2", "SHIFT_3"].filter(
            (code) => code !== selectedShift
          );

          const [operatorResult, selectedProductivity, ...extraProductivity] =
            await Promise.allSettled([
              getMachineOperatorReport(dateText),
              getProductivity(dateText, { shift: shiftParam }),
              ...extraShiftCodes.map((code) =>
                getProductivity(dateText, { shift: code })
              ),
            ]);

          if (selectedProductivity.status === "rejected") {
            throw selectedProductivity.reason;
          }

          const productivityByShift = {
            DEFAULT: selectedProductivity.value,
          };

          if (["SHIFT_1", "SHIFT_2", "SHIFT_3"].includes(selectedShift)) {
            productivityByShift[selectedShift] = selectedProductivity.value;
          }

          extraShiftCodes.forEach((code, index) => {
            const result = extraProductivity[index];
            if (result?.status === "fulfilled") {
              productivityByShift[code] = result.value;
            }
          });

          const shiftMetricsMap = buildUuidShiftMetricsMap(productivityByShift);

          const operatorMap =
            operatorResult.status === "fulfilled"
              ? buildOperatorExportMap(operatorResult.value, {
                  workDate: dateText,
                  shiftCode: shiftParam || "ALL",
                  shiftConfigMap: configMap,
                  getLocationByUuid,
                })
              : new Map();

          const rows = extractRows(selectedProductivity.value);

          return rows.map((row) =>
            normalizeRangeItem(
              row,
              dateText,
              operatorMap,
              getManualSettingByUuid,
              shiftParam,
              configMap,
              shiftMetricsMap
            )
          );
        })
      );

      const items = filterRangeItemsByStatus(
        filterRangeItemsByLocation(results.flat(), locationFilter.value),
        statusFilter?.value
      );

      if (!items.length) {
        showNotice(
          "Data export tidak ditemukan untuk filter area/status ini.",
          "error"
        );
        return;
      }

      const summaryBaseRows = buildRangeSummaryBaseRows(
        items,
        range.start,
        range.end,
        shiftParam
      );

      const summaryRows = buildRangeSummaryRows(summaryBaseRows);
      const detailRows = buildRangeDetailRows(items, shiftParam);

      const workbook = createRangeWorkbook(summaryRows, detailRows);

      const locationSuffix =
        locationFilter.value === "ALL"
          ? "all-gm"
          : safeFileName(locationFilter.value);

      const statusSelected = String(statusFilter?.value || "ALL")
        .trim()
        .toUpperCase();
      const statusSuffix =
        statusSelected && statusSelected !== "ALL"
          ? `-${safeFileName(statusSelected)}`
          : "";

      const shiftSuffix = shiftParam
        ? `-${safeFileName(shiftParam)}`
        : "";

      writeWorkbook(
        workbook,
        `productivity-range-${locationSuffix}${statusSuffix}${shiftSuffix}-${range.start}_${range.end}.xlsx`
      );

      showNotice("Export Excel berhasil dibuat.", "ok");
    } catch (err) {
      showNotice(`Gagal export Excel: ${err.message}`, "error");
    } finally {
      rangeExporting.value = false;
    }
  }

  return {
    startDate,
    endDate,
    rangeExporting,
    downloadExcel,
  };
}
