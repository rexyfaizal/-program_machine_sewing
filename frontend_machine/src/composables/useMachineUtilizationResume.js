import { computed, onBeforeUnmount, ref, watch } from "vue";
import { getMachineOperatorReport, getProductivity } from "../api/machineApi";
import { getWeekdaysBetween } from "../utils/dashboardExportExcel";
import { getMonthWeekOptions, MONTH_LABELS_ID, todayLocal } from "../utils/format";
import {
  buildMachineUtilizationResume,
  filterMachineUtilizationRows,
  weeklyAverages,
} from "../utils/machineUtilizationResume";

export function useMachineUtilizationResume(initialDate = "") {
  const period = ref(String(initialDate || todayLocal()).slice(0, 7));
  const building = ref("ALL");
  const keyword = ref("");
  const selectedWeekNumbers = ref([1, 2]);
  const days = ref([]);
  const loading = ref(false);
  const errorMessage = ref("");
  const completedDays = ref(0);
  const totalDays = ref(0);
  const lastUpdate = ref("");
  const cache = new Map();
  let controller = null;
  let requestSeq = 0;

  const periodParts = computed(() => {
    if (!/^\d{4}-(0[1-9]|1[0-2])$/.test(period.value)) return null;
    const [year, month] = period.value.split("-").map(Number);
    return year >= 1900 && year <= 2100 ? { year, month } : null;
  });
  const monthLabel = computed(() => periodParts.value
    ? `${MONTH_LABELS_ID[periodParts.value.month - 1]} ${periodParts.value.year}` : "");
  const weekOptions = computed(() => periodParts.value
    ? getMonthWeekOptions(periodParts.value.year, periodParts.value.month - 1) : []);
  const weeks = computed(() => weekOptions.value.filter((week) => selectedWeekNumbers.value.includes(week.weekNo)));
  const rows = computed(() => buildMachineUtilizationResume(days.value, weeks.value));
  const filteredRows = computed(() => filterMachineUtilizationRows(rows.value, building.value, keyword.value));
  const averages = computed(() => weeklyAverages(filteredRows.value, weeks.value));

  async function load({ refresh = false } = {}) {
    const seq = ++requestSeq;
    controller?.abort();
    controller = new AbortController();
    const { signal } = controller;
    days.value = [];
    errorMessage.value = "";
    lastUpdate.value = "";
    completedDays.value = 0;
    const dates = weeks.value.flatMap((week) => getWeekdaysBetween(week.start, week.end))
      .filter((date) => date <= todayLocal());
    totalDays.value = dates.length;
    if (!periodParts.value || !weeks.value.length) {
      loading.value = false;
      errorMessage.value = "Pilih bulan dan minimal satu minggu untuk menampilkan laporan.";
      return;
    }
    if (refresh) dates.forEach((date) => cache.delete(date));
    loading.value = true;
    let cursor = 0;
    const loaded = [];

    async function worker() {
      while (cursor < dates.length && !signal.aborted) {
        const date = dates[cursor++];
        let day = cache.get(date);
        if (!day) {
          const results = await Promise.allSettled([
            getProductivity(date, { shift: "ALL", signal }),
            getMachineOperatorReport(date, { withStats: true, forExport: true, signal }),
          ]);
          if (signal.aborted) return;
          const failures = results.flatMap((result, index) => result.status === "rejected"
            ? [`${index === 0 ? "data mesin" : "data operator"}: ${result.reason?.message || result.reason}`] : []);
          if (failures.length) throw new Error(`${date} — ${failures.join("; ")}`);
          day = { date, productivity: results[0].value, operators: results[1].value };
          cache.set(date, day);
        }
        loaded.push(day);
        if (seq === requestSeq) completedDays.value += 1;
      }
    }

    try {
      await Promise.all([worker(), worker()]);
      if (seq !== requestSeq || signal.aborted) return;
      days.value = loaded;
      lastUpdate.value = new Date().toLocaleString("id-ID");
    } catch (error) {
      if (seq === requestSeq && !signal.aborted) {
        errorMessage.value = `Gagal memuat resume: ${error.message}. Klik Muat ulang untuk mencoba kembali.`;
        controller.abort();
      }
    } finally {
      if (seq === requestSeq) loading.value = false;
    }
  }

  watch(period, () => {
    if (!periodParts.value) return;
    const valid = selectedWeekNumbers.value.filter((number) => weekOptions.value.some((week) => week.weekNo === number));
    selectedWeekNumbers.value = valid.length ? valid : [weekOptions.value[0].weekNo];
  });
  watch(() => `${period.value}|${weeks.value.map((week) => week.weekNo).join(",")}`, () => load(), { immediate: true });
  onBeforeUnmount(() => {
    requestSeq += 1;
    controller?.abort();
  });

  return { period, building, keyword, selectedWeekNumbers, monthLabel, weekOptions, weeks,
    filteredRows, averages, loading, errorMessage, completedDays, totalDays, lastUpdate, load };
}
