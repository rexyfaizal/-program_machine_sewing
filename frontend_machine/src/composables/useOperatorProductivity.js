import { computed, ref } from "vue";
import {
  getLineShiftConfig,
  getMachineOperatorReport,
  getMachineSettings,
  getProductivity,
} from "../api/machineApi";
import {
  buildExportLossBreakdown,
  buildUuidShiftMetricsMap,
  extractRows,
  getLocationGroup,
  getVal,
  isAutoLogoutText,
  resolveDashboardMetricsForOperator,
  formatLossPercent,
  resolveExportOperatorShiftTag,
  toNumber,
} from "../utils/dashboardExportExcel";
import { buildLineShiftConfigMap } from "../utils/gm3Shift";
import { formatDurationHHMMSS } from "../utils/format";

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function formatLongDate(dateText) {
  const raw = String(dateText || "").trim();
  if (!raw) return "";

  const d = new Date(`${raw}T00:00:00`);
  if (Number.isNaN(d.getTime())) return raw;

  return d.toLocaleDateString("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}

function lineSortParts(location) {
  const text = String(location || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
  const gm = text.match(/\bGM\s*(\d+)/);
  const line = text.match(/\bLINE\s*(\d+)/);

  return {
    gm: gm ? Number(gm[1]) : 9999,
    line: line ? Number(line[1]) : 9999,
    text,
  };
}

function compareOperatorRows(a, b) {
  const ka = lineSortParts(a.location || a.locationLabel);
  const kb = lineSortParts(b.location || b.locationLabel);

  if (ka.gm !== kb.gm) return ka.gm - kb.gm;
  if (ka.line !== kb.line) return ka.line - kb.line;

  const locCmp = String(a.locationLabel || "").localeCompare(
    String(b.locationLabel || "")
  );
  if (locCmp) return locCmp;

  if (Boolean(a.loggedIn) !== Boolean(b.loggedIn)) {
    return a.loggedIn ? -1 : 1;
  }

  return String(a.operatorName || "").localeCompare(String(b.operatorName || ""));
}

function applyDashboardMetrics(row, metrics) {
  if (!metrics) {
    return {
      ...row,
      output: 0,
      avgCycle: 0,
      runtimeSec: 0,
      procSec: 0,
      lossTimeSec: 0,
      productivity: 0,
      powerOnText: "00:00:00",
      processText: "00:00:00",
      lossText: "00:00:00",
    };
  }

  return {
    ...row,
    output: metrics.output,
    avgCycle: metrics.avgCT,
    runtimeSec: metrics.runtime,
    procSec: metrics.procTime,
    lossTimeSec: metrics.lossTime,
    productivity: metrics.productivity,
    powerOnText: formatDurationHHMMSS(metrics.runtime),
    processText: formatDurationHHMMSS(metrics.procTime),
    lossText: formatDurationHHMMSS(metrics.lossTime),
  };
}

function attachNotePercents(row) {
  const powerOn = Number(row.runtimeSec || 0);

  return {
    ...row,
    tungguBahanPct: formatLossPercent(row.tungguBahanSec, powerOn),
    mesinRusakPct: formatLossPercent(row.mesinRusakSec, powerOn),
    toiletPct: formatLossPercent(row.toiletSec, powerOn),
    solatPct: formatLossPercent(row.solatSec, powerOn),
    otherPct: formatLossPercent(row.otherSec, powerOn),
  };
}

export function useOperatorProductivity() {
  const rows = ref([]);
  const loading = ref(false);
  const errorMessage = ref("");
  const keyword = ref("");
  const locationFilter = ref("ALL");

  function buildSettingsMap(data) {
    const map = new Map();

    extractRows(data).forEach((row) => {
      const uuid = String(getVal(row, "uuid", "UUID") || "").trim();
      if (!uuid) return;

      map.set(normalizeText(uuid), {
        location: String(getVal(row, "location", "Location") || "").trim(),
        spv: String(getVal(row, "spv", "SPV") || "").trim(),
        customName: String(
          getVal(row, "customName", "CustomName", "custom_name") || ""
        ).trim(),
      });
    });

    return map;
  }

  function normalizeSession(row, settingsMap, shiftConfigMap, workDate) {
    if (!row) return null;

    const uuid = String(getVal(row, "uuid", "UUID") || "").trim();
    const operatorNik = String(
      getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
    ).trim();
    const operatorName = String(
      getVal(row, "operatorName", "OperatorName", "operator_name") || ""
    ).trim();

    if (!uuid || (!operatorNik && !operatorName)) return null;

    if (isAutoLogoutText(operatorNik) || isAutoLogoutText(operatorName)) {
      return null;
    }

    const setting = settingsMap.get(normalizeText(uuid)) || {};
    const locationFromApi = String(getVal(row, "location", "Location") || "").trim();
    const location = setting.location || locationFromApi || "-";
    const spv = String(setting.spv || "").trim();
    const locationLabel = spv ? `${location} - ${spv}` : location;
    const area = getLocationGroup(location);

    const processName = String(
      getVal(row, "processName", "ProcessName", "process_name") || ""
    ).trim();
    const machineName = String(
      getVal(row, "machineName", "MachineName") || ""
    ).trim();

    const notes = extractRows(getVal(row, "notes", "Notes") || []);
    const loss = buildExportLossBreakdown(notes);

    const shiftTag =
      resolveExportOperatorShiftTag({
        loginTime: String(getVal(row, "loginTime", "LoginTime") || "").trim(),
        logoutTime: String(getVal(row, "logoutTime", "LogoutTime") || "").trim(),
        status: String(getVal(row, "status", "Status") || "").trim(),
        location,
        workDate,
        shiftConfigMap,
      }) || "Normal";

    return {
      id: Number(getVal(row, "id", "ID") || 0),
      uuid,
      area,
      location,
      locationLabel,
      operatorNik,
      operatorName,
      shiftTag,
      mesin: processName || machineName || "-",
      tungguBahanSec: toNumber(loss.tungguHancaSec),
      mesinRusakSec: toNumber(loss.mesinRusakSec),
      toiletSec: toNumber(loss.toiletSec),
      solatSec: toNumber(loss.solatSec),
      otherSec: toNumber(loss.otherSec),
      tungguBahanText: formatDurationHHMMSS(loss.tungguHancaSec),
      mesinRusakText: formatDurationHHMMSS(loss.mesinRusakSec),
      toiletText: formatDurationHHMMSS(loss.toiletSec),
      solatText: formatDurationHHMMSS(loss.solatSec),
      otherText: formatDurationHHMMSS(loss.otherSec),
      remarks: String(loss.remarksOther || "").trim() || "-",
      loginTime: String(getVal(row, "loginTime", "LoginTime") || "").trim(),
      loggedIn: true,
    };
  }

  function emptyLossFields() {
    return {
      tungguBahanSec: 0,
      mesinRusakSec: 0,
      toiletSec: 0,
      solatSec: 0,
      otherSec: 0,
      tungguBahanText: "00:00:00",
      mesinRusakText: "00:00:00",
      toiletText: "00:00:00",
      solatText: "00:00:00",
      otherText: "00:00:00",
      remarks: "-",
    };
  }

  function normalizeUnloggedMachine(row, settingsMap) {
    const uuid = String(getVal(row, "uuid", "UUID") || "").trim();
    if (!uuid) return null;

    const setting = settingsMap.get(normalizeText(uuid)) || {};
    const locationFromApi = String(getVal(row, "location", "Location") || "").trim();
    const location = setting.location || locationFromApi || "-";
    const spv = String(setting.spv || "").trim();
    const locationLabel = spv ? `${location} - ${spv}` : location;
    const area = getLocationGroup(location);

    const backendName = String(
      getVal(
        row,
        "nickName",
        "NickName",
        "machineName",
        "MachineName",
        "name",
        "Name"
      ) || uuid
    ).trim();
    const mesin = setting.customName || backendName || uuid;

    return {
      id: `unlogged-${uuid}`,
      uuid,
      area,
      location,
      locationLabel,
      operatorNik: "",
      operatorName: "Not logged in",
      shiftTag: "-",
      mesin,
      ...emptyLossFields(),
      loginTime: "",
      loggedIn: false,
    };
  }

  async function loadOperatorProductivity(date) {
    const requestDate = String(date || "").trim();
    if (!requestDate) return;

    loading.value = true;
    errorMessage.value = "";

    try {
      const [
        reportData,
        settingsData,
        shiftConfigData,
        currentProd,
        shift1Prod,
        shift2Prod,
        shift3Prod,
      ] = await Promise.all([
        getMachineOperatorReport(requestDate),
        getMachineSettings().catch(() => []),
        getLineShiftConfig("").catch(() => ({ lines: [] })),
        getProductivity(requestDate, { shift: "CURRENT" }).catch(() => []),
        getProductivity(requestDate, { shift: "SHIFT_1" }).catch(() => []),
        getProductivity(requestDate, { shift: "SHIFT_2" }).catch(() => []),
        getProductivity(requestDate, { shift: "SHIFT_3" }).catch(() => []),
      ]);

      const settingsMap = buildSettingsMap(settingsData);
      const shiftConfigMap = buildLineShiftConfigMap(
        Array.isArray(shiftConfigData?.lines) ? shiftConfigData.lines : []
      );
      const shiftMetricsMap = buildUuidShiftMetricsMap({
        DEFAULT: currentProd,
        CURRENT: currentProd,
        SHIFT_1: shift1Prod,
        SHIFT_2: shift2Prod,
        SHIFT_3: shift3Prod,
      });

      const sessionRows = extractRows(reportData)
        .map((row) =>
          normalizeSession(row, settingsMap, shiftConfigMap, requestDate)
        )
        .filter(Boolean)
        .map((row) =>
          attachNotePercents(
            applyDashboardMetrics(
              row,
              resolveDashboardMetricsForOperator(
                row.uuid,
                row.shiftTag,
                shiftMetricsMap
              )
            )
          )
        );

      const loggedUuids = new Set(
        sessionRows.map((row) => normalizeText(row.uuid))
      );

      const unloggedRows = extractRows(currentProd)
        .map((row) => normalizeUnloggedMachine(row, settingsMap))
        .filter((row) => row && !loggedUuids.has(normalizeText(row.uuid)))
        .map((row) =>
          attachNotePercents(
            applyDashboardMetrics(
              row,
              resolveDashboardMetricsForOperator(row.uuid, "-", shiftMetricsMap)
            )
          )
        );

      const normalized = [...sessionRows, ...unloggedRows].sort(
        compareOperatorRows
      );

      rows.value = normalized;
    } catch (err) {
      errorMessage.value =
        err?.message || "Gagal mengambil data produktivitas operator.";
      rows.value = [];
    } finally {
      loading.value = false;
    }
  }

  const locationOptions = computed(() => {
    const set = new Set(
      rows.value.map((row) => row.area).filter((area) => area && area !== "-")
    );
    return ["ALL", ...[...set].sort()];
  });

  const filteredRows = computed(() => {
    const key = String(keyword.value || "")
      .trim()
      .toLowerCase();
    const area = String(locationFilter.value || "ALL").trim().toUpperCase();

    return rows.value.filter((row) => {
      if (area !== "ALL" && String(row.area || "").toUpperCase() !== area) {
        return false;
      }

      if (!key) return true;

      return [
        row.area,
        row.locationLabel,
        row.uuid,
        row.operatorName,
        row.operatorNik,
        row.shiftTag,
        row.mesin,
        row.remarks,
      ]
        .join(" ")
        .toLowerCase()
        .includes(key);
    });
  });

  const machineCount = computed(() => {
    const uuids = new Set(
      filteredRows.value
        .map((row) => String(row.uuid || "").trim().toUpperCase())
        .filter(Boolean)
    );
    return uuids.size;
  });
  const loggedInCount = computed(
    () => filteredRows.value.filter((row) => row.loggedIn).length
  );
  const unloggedCount = computed(
    () => filteredRows.value.filter((row) => !row.loggedIn).length
  );

  const averages = computed(() => {
    const list = filteredRows.value;
    const count = list.length;

    if (!count) {
      return {
        output: 0,
        avgCycle: 0,
        runtimeSec: 0,
        procSec: 0,
        lossTimeSec: 0,
        productivity: 0,
        tungguBahanSec: 0,
        mesinRusakSec: 0,
        toiletSec: 0,
        solatSec: 0,
        otherSec: 0,
        tungguBahanPct: "",
        mesinRusakPct: "",
        toiletPct: "",
        solatPct: "",
        otherPct: "",
      };
    }

    const sum = list.reduce(
      (acc, row) => {
        acc.output += Number(row.output || 0);
        acc.avgCycle += Number(row.avgCycle || 0);
        acc.runtimeSec += Number(row.runtimeSec || 0);
        acc.procSec += Number(row.procSec || 0);
        acc.lossTimeSec += Number(row.lossTimeSec || 0);
        acc.productivity += Number(row.productivity || 0);
        acc.tungguBahanSec += Number(row.tungguBahanSec || 0);
        acc.mesinRusakSec += Number(row.mesinRusakSec || 0);
        acc.toiletSec += Number(row.toiletSec || 0);
        acc.solatSec += Number(row.solatSec || 0);
        acc.otherSec += Number(row.otherSec || 0);
        return acc;
      },
      {
        output: 0,
        avgCycle: 0,
        runtimeSec: 0,
        procSec: 0,
        lossTimeSec: 0,
        productivity: 0,
        tungguBahanSec: 0,
        mesinRusakSec: 0,
        toiletSec: 0,
        solatSec: 0,
        otherSec: 0,
      }
    );

    const runtimeSec = Math.round(sum.runtimeSec / count);
    const tungguBahanSec = Math.round(sum.tungguBahanSec / count);
    const mesinRusakSec = Math.round(sum.mesinRusakSec / count);
    const toiletSec = Math.round(sum.toiletSec / count);
    const solatSec = Math.round(sum.solatSec / count);
    const otherSec = Math.round(sum.otherSec / count);

    return {
      output: Math.round(sum.output / count),
      avgCycle: Number((sum.avgCycle / count).toFixed(2)),
      runtimeSec,
      procSec: Math.round(sum.procSec / count),
      lossTimeSec: Math.round(sum.lossTimeSec / count),
      productivity: Number((sum.productivity / count).toFixed(2)),
      tungguBahanSec,
      mesinRusakSec,
      toiletSec,
      solatSec,
      otherSec,
      tungguBahanPct: formatLossPercent(tungguBahanSec, runtimeSec),
      mesinRusakPct: formatLossPercent(mesinRusakSec, runtimeSec),
      toiletPct: formatLossPercent(toiletSec, runtimeSec),
      solatPct: formatLossPercent(solatSec, runtimeSec),
      otherPct: formatLossPercent(otherSec, runtimeSec),
    };
  });

  return {
    rows,
    loading,
    errorMessage,
    keyword,
    locationFilter,
    locationOptions,
    filteredRows,
    machineCount,
    loggedInCount,
    unloggedCount,
    averages,
    formatLongDate,
    loadOperatorProductivity,
  };
}
