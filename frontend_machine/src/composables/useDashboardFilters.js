import { computed, ref, watch } from "vue";
import { getLocationGroup } from "../utils/dashboardExportExcel";
import {
  GM3_SHIFT_OPTIONS,
  buildShiftOptionsFromConfigMap,
  resolveLineShiftForLocation,
} from "../utils/gm3Shift";

const SPECIFIC_SHIFT_CODES = ["SHIFT_1", "SHIFT_2", "SHIFT_3"];

export const STATUS_FILTER_OPTIONS = [
  { value: "ALL", label: "All Status" },
  { value: "GOOD", label: "GOOD" },
  { value: "NORMAL", label: "NORMAL" },
  { value: "BAD", label: "BAD" },
];

export function useDashboardFilters(machines, shiftConfigMap = null) {
  const keyword = ref("");
  const locationFilter = ref("ALL");
  const shiftFilter = ref("CURRENT");
  const statusFilter = ref("ALL");

  const locationOptions = computed(() => {
    return [
      ...new Set(
        machines.value
          .map((m) => getLocationGroup(m.location))
          .filter((location) => location && location !== "-")
      ),
    ].sort((a, b) => {
      const numA = Number(String(a).replace(/\D/g, ""));
      const numB = Number(String(b).replace(/\D/g, ""));

      if (numA && numB) return numA - numB;

      return String(a).localeCompare(String(b));
    });
  });

  const configMap = computed(() => {
    if (shiftConfigMap?.value instanceof Map) {
      return shiftConfigMap.value;
    }
    return new Map();
  });

  // Shift filter selalu tampil, termasuk Area = All GM,
  // supaya GM3 tetap dihitung per shift (default Current).
  const showShiftFilter = computed(() => true);

  const shiftOptions = computed(() => {
    const area = String(locationFilter.value || "").trim().toUpperCase();
    if (!area || area === "ALL") return GM3_SHIFT_OPTIONS;

    return buildShiftOptionsFromConfigMap(area, configMap.value);
  });

  const productivityShift = computed(() => {
    const selected = String(shiftFilter.value || "").trim().toUpperCase();
    const validOption = shiftOptions.value.some(
      (option) => String(option.value || "").toUpperCase() === selected
    );
    const resolved = validOption
      ? selected
      : String(shiftOptions.value[0]?.value || "CURRENT").toUpperCase();

    return resolved;
  });

  watch(
    locationOptions,
    (options) => {
      if (locationFilter.value === "ALL") return;

      if (!options.includes(locationFilter.value)) {
        locationFilter.value = "ALL";
      }
    },
    { immediate: true }
  );

  watch(
    shiftOptions,
    (options) => {
      const selected = String(shiftFilter.value || "").trim().toUpperCase();
      const isValid = options.some(
        (option) => String(option.value || "").toUpperCase() === selected
      );

      if (!isValid) {
        shiftFilter.value = String(options[0]?.value || "CURRENT");
      }
    },
    { immediate: true }
  );

  // Shift 1/2/3 spesifik tidak berlaku untuk line "Hari Penuh".
  // Current Shift tetap menampilkan line Normal (Hari Penuh).
  const hideFullDayLines = computed(() =>
    SPECIFIC_SHIFT_CODES.includes(
      String(productivityShift.value || "").trim().toUpperCase()
    )
  );

  const filteredMachines = computed(() => {
    const key = keyword.value.toLowerCase().trim();
    const selectedLocation = String(locationFilter.value || "ALL").trim();
    const selectedStatus = String(statusFilter.value || "ALL")
      .trim()
      .toUpperCase();

    return machines.value.filter((m) => {
      const locationGroup = getLocationGroup(m.location);

      if (selectedLocation !== "ALL" && locationGroup !== selectedLocation) {
        return false;
      }

      if (
        selectedStatus !== "ALL" &&
        String(m.status || "").trim().toUpperCase() !== selectedStatus
      ) {
        return false;
      }

      if (hideFullDayLines.value) {
        const lineShift = resolveLineShiftForLocation(m.location, configMap.value);
        if (!lineShift.useShift) return false;
      }

      if (!key) return true;

      return [
        m.machineName,
        m.displayMachineName,
        m.originalMachineName,
        m.location,

        m.pic,
        m.operatorNik,
        m.operatorName,
        m.operatorSubText,
        m.operatorLoginText,
        m.operatorActiveText,

        m.operatorNote,
        m.operatorNotes,
        m.spv,

        m.operatorProcessName,
        m.operatorStyleName,

        m.ip,
        m.uuid,
        m.tableName,
        m.program,
        m.status,
      ]
        .join(" ")
        .toLowerCase()
        .includes(key);
    });
  });

  function makeSummary(list) {
    const totalMachine = list.length;
    const totalProductivity = list.reduce(
      (sum, m) => sum + Number(m.productivity || 0),
      0
    );

    return {
      totalMachine,
      avgProductivity: totalMachine ? totalProductivity / totalMachine : 0,
      good: list.filter((m) => m.status === "GOOD").length,
      normal: list.filter((m) => m.status === "NORMAL").length,
      bad: list.filter((m) => m.status === "BAD").length,
      totalOutput: list.reduce((sum, m) => sum + Number(m.output || 0), 0),
      totalAlarm: list.reduce((sum, m) => sum + Number(m.alarm || 0), 0),
      totalSewingTime: list.reduce((sum, m) => sum + Number(m.procTime || 0), 0),
    };
  }

  const rankingMachines = computed(() => {
    return [...filteredMachines.value].sort(
      (a, b) => Number(b.productivity || 0) - Number(a.productivity || 0)
    );
  });

  const attentionMachines = computed(() => {
    return [...filteredMachines.value]
      .sort((a, b) => Number(a.productivity || 0) - Number(b.productivity || 0))
      .slice(0, 10);
  });

  const bestMachine = computed(() => rankingMachines.value[0] || null);
  const worstMachine = computed(() => attentionMachines.value[0] || null);
  const summary = computed(() => makeSummary(filteredMachines.value));

  const executiveMessage = computed(() => {
    const s = summary.value;

    if (!s.totalMachine) {
      return "Belum ada data mesin pada tanggal ini.";
    }

    return `${s.good} mesin GOOD, ${s.normal} NORMAL, dan ${s.bad} BAD dari total ${
      s.totalMachine
    } mesin. Rata-rata produktivitas ${s.avgProductivity.toFixed(2)}%.`;
  });

  const donutStyle = computed(() => {
    const total = summary.value.totalMachine || 1;
    const good = (summary.value.good / total) * 100;
    const normal = (summary.value.normal / total) * 100;
    const badStart = good + normal;

    return {
      background: `conic-gradient(#22c55e 0 ${good}%, #f59e0b ${good}% ${badStart}%, #ef4444 ${badStart}% 100%)`,
    };
  });

  return {
    keyword,
    locationFilter,
    shiftFilter,
    statusFilter,
    statusOptions: STATUS_FILTER_OPTIONS,
    shiftOptions,
    showShiftFilter,
    productivityShift,
    locationOptions,
    filteredMachines,
    rankingMachines,
    attentionMachines,
    bestMachine,
    worstMachine,
    summary,
    executiveMessage,
    donutStyle,
  };
}
