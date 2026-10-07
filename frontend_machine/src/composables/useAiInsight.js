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

  let timer = null;
  let inFlight = false;

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
        return;
      }

      const data = await analyzeAiInsight({
        date: resolveDate(),
        shift: resolveShift(),
        location: resolveLocation(),
        machines: machineRows,
      });

      insight.value = data;
      lastUpdatedAt.value = data?.generatedAt || new Date().toLocaleString("id-ID");
    } catch (err) {
      errorMessage.value = err?.message || "Gagal menganalisa AI insight.";
    } finally {
      loading.value = false;
      inFlight = false;
    }
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
    refreshAiInsight,
    startAutoRefresh,
    stopAutoRefresh,
  };
}
