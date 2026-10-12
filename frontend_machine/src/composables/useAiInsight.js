import { onUnmounted, ref } from "vue";

import { analyzeAiInsight } from "../api/machineApi";
import { collectMachineOperatorNotes } from "../utils/machineTableFormat";

const REFRESH_MS = 5 * 60 * 1000;

function firstOperatorName(machine) {
  const rows = Array.isArray(machine?.operatorRows)
    ? machine.operatorRows
    : Array.isArray(machine?.operatorDisplayRows)
      ? machine.operatorDisplayRows
      : [];
  if (rows.length) {
    return String(
      rows[0]?.operatorName ||
        rows[0]?.name ||
        rows[0]?.label ||
        ""
    )
      .trim()
      .replace(/^\d+\s*-\s*/, "");
  }
  return String(machine?.operatorName || machine?.pic || "").trim();
}

function toMachinePayload(machines = []) {
  return (Array.isArray(machines) ? machines : [])
    .map((m) => ({
      uuid: String(m?.uuid || "").trim(),
      machineName: String(m?.displayMachineName || m?.machineName || "").trim(),
      location: String(m?.location || "").trim(),
      macState: String(m?.macState ?? "").trim(),
      status: String(m?.status || "").trim(),
      productivityPct: Number(m?.productivityPct ?? m?.productivity ?? 0) || 0,
      runtimeSec: Number(m?.runtime ?? m?.runtimeSec ?? 0) || 0,
      procSec: Number(m?.procTime ?? m?.procSec ?? 0) || 0,
      lossTimeSec: Number(m?.lossTime ?? m?.lossTimeSec ?? 0) || 0,
      operatorName: firstOperatorName(m),
      operatorNotes: collectMachineOperatorNotes(m),
    }))
    .filter((m) => m.uuid);
}

function itemIdentity(item) {
  const uuid = String(item?.uuid || "").trim().toLowerCase();
  if (uuid) return `u:${uuid}`;
  return [
    "x",
    String(item?.scope || "").trim().toUpperCase(),
    String(item?.location || "").trim().toLowerCase(),
    String(item?.code || "").trim().toUpperCase(),
    String(item?.machineName || "").trim().toLowerCase(),
  ].join("|");
}

function priorityRank(priority) {
  const p = String(priority || "").toUpperCase();
  if (p === "HIGH") return 3;
  if (p === "MEDIUM") return 2;
  return 1;
}

function buildPriorityMap(items = []) {
  const map = {};
  for (const item of Array.isArray(items) ? items : []) {
    map[itemIdentity(item)] = priorityRank(item?.priority);
  }
  return map;
}

function computeRisenKeys(prevMap, items = []) {
  const risen = [];
  if (!prevMap || typeof prevMap !== "object") return risen;
  for (const item of Array.isArray(items) ? items : []) {
    const key = itemIdentity(item);
    const rank = priorityRank(item?.priority);
    const prev = prevMap[key];
    if (prev === undefined) {
      if (rank >= 2) risen.push(key);
    } else if (rank > prev) {
      risen.push(key);
    }
  }
  return risen;
}

export function useAiInsight({
  selectedDate,
  locationFilter,
  shiftFilter,
  machines,
}) {
  const loading = ref(false);
  const errorMessage = ref("");
  const insight = ref(null);
  const lastUpdatedAt = ref("");
  const risenKeys = ref([]);
  const changeSummary = ref("");

  let timer = null;
  let inFlight = false;
  let previousPriorityMap = null;

  function resolveDate() {
    if (typeof selectedDate === "function") return selectedDate() || "";
    return selectedDate?.value || selectedDate || "";
  }

  function resolveLocation() {
    if (typeof locationFilter === "function") return locationFilter() || "";
    return locationFilter?.value || locationFilter || "";
  }

  function resolveShift() {
    if (typeof shiftFilter === "function") return shiftFilter() || "";
    return shiftFilter?.value || shiftFilter || "";
  }

  function resolveMachines() {
    if (typeof machines === "function") return machines() || [];
    return machines?.value || machines || [];
  }

  async function refreshAiInsight({ silent = false } = {}) {
    if (inFlight) return;
    inFlight = true;

    if (!silent) {
      loading.value = true;
    }
    errorMessage.value = "";

    try {
      const machineRows = toMachinePayload(resolveMachines());
      if (!machineRows.length) {
        insight.value = {
          status: "empty",
          summary: "Menunggu data mesin dashboard...",
          items: [],
          highCount: 0,
          mediumCount: 0,
          lowCount: 0,
          machineCount: 0,
        };
        risenKeys.value = [];
        changeSummary.value = "";
        return;
      }

      const data = await analyzeAiInsight({
        date: resolveDate(),
        shift: resolveShift(),
        location: resolveLocation(),
        machines: machineRows,
      });

      const nextItems = Array.isArray(data?.items) ? data.items : [];
      if (previousPriorityMap) {
        const risen = computeRisenKeys(previousPriorityMap, nextItems);
        risenKeys.value = risen;
        changeSummary.value = risen.length
          ? `${risen.length} kartu naik prioritas / baru muncul sejak tadi.`
          : "Tidak ada kartu yang naik prioritas sejak tadi.";
      } else {
        risenKeys.value = [];
        changeSummary.value = "Acuan pertama tersimpan. Cek lagi nanti untuk lihat yang berubah.";
      }
      previousPriorityMap = buildPriorityMap(nextItems);

      insight.value = data;
      lastUpdatedAt.value = data?.generatedAt || new Date().toLocaleString("id-ID");
    } catch (err) {
      errorMessage.value = err?.message || "Gagal menganalisa AI insight.";
    } finally {
      loading.value = false;
      inFlight = false;
    }
  }

  async function checkChangesSince() {
    await refreshAiInsight({ silent: false });
    return {
      risenKeys: risenKeys.value,
      summary: changeSummary.value,
    };
  }

  function startAutoRefresh() {
    stopAutoRefresh();
    timer = setInterval(() => {
      refreshAiInsight({ silent: true });
    }, REFRESH_MS);
  }

  function stopAutoRefresh() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  onUnmounted(() => {
    stopAutoRefresh();
  });

  return {
    loading,
    errorMessage,
    insight,
    lastUpdatedAt,
    risenKeys,
    changeSummary,
    refreshAiInsight,
    checkChangesSince,
    startAutoRefresh,
    stopAutoRefresh,
  };
}
