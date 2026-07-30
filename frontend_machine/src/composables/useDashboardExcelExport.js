import { ref, watch } from "vue";
import {
  getMachineOperatorReport,
  getProductivity,
} from "../api/machineApi";
import {
  buildBadPriorityRows,
  buildOperatorExportMap,
  buildRangeDetailRows,
  buildRangeSummaryBaseRows,
  buildRangeSummaryRows,
  createRangeWorkbook,
  extractRows,
  filterRangeItemsByLocation,
  formatExcelDate,
  getOrderedRange,
  getWeekdaysBetween,
  normalizeRangeItem,
  normalizeText,
  safeFileName,
  writeWorkbook,
} from "../utils/dashboardExportExcel";

export function useDashboardExcelExport({
  selectedDate,
  locationFilter,
  machineSettings,
  showNotice,
}) {
  const rangeExporting = ref(false);
  const startDate = ref(selectedDate.value || "");
  const endDate = ref(selectedDate.value || "");

  watch(
    selectedDate,
    (value) => {
      if (!value) return;

      startDate.value = value;
      endDate.value = value;
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
      showNotice("Range tanggal tidak memiliki hari kerja Senin-Jumat.", "error");
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
      const results = await Promise.all(
        dates.map(async (dateText) => {
          const [productivityResult, operatorResult] = await Promise.allSettled([
            getProductivity(dateText),
            getMachineOperatorReport(dateText),
          ]);

          if (productivityResult.status === "rejected") {
            throw productivityResult.reason;
          }

          const operatorMap =
            operatorResult.status === "fulfilled"
              ? buildOperatorExportMap(operatorResult.value)
              : new Map();

          const rows = extractRows(productivityResult.value);

          return rows.map((row) =>
            normalizeRangeItem(
              row,
              dateText,
              operatorMap,
              getManualSettingByUuid
            )
          );
        })
      );

      const items = filterRangeItemsByLocation(
        results.flat(),
        locationFilter.value
      );

      if (!items.length) {
        showNotice("Data export tidak ditemukan untuk filter area ini.", "error");
        return;
      }

      const summaryBaseRows = buildRangeSummaryBaseRows(
        items,
        range.start,
        range.end
      );

      const summaryRows = buildRangeSummaryRows(summaryBaseRows);
      const detailRows = buildRangeDetailRows(items);
      const badPriorityRows = buildBadPriorityRows(
        summaryBaseRows,
        locationFilter.value
      );

      const workbook = createRangeWorkbook(
        summaryRows,
        detailRows,
        badPriorityRows
      );

      const locationSuffix =
        locationFilter.value === "ALL"
          ? "all-gm"
          : safeFileName(locationFilter.value);

      writeWorkbook(
        workbook,
        `productivity-range-${locationSuffix}-${range.start}_${range.end}.xlsx`
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