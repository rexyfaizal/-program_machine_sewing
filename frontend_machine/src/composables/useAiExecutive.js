import { ref } from "vue";

import { getProductivity, postAiExecutive } from "../api/machineApi";
import { collectMachineOperatorNotes } from "../utils/machineTableFormat";

const SNAPSHOT_KEY = "ai_exec_snapshot_v1";
const SNAPSHOT_MIN_GAP_MS = 5 * 60 * 1000;

function firstOperatorName(machine) {
  const rows = Array.isArray(machine?.operatorRows)
    ? machine.operatorRows
    : Array.isArray(machine?.operatorDisplayRows)
      ? machine.operatorDisplayRows
      : [];
  if (rows.length) {
    return String(
      rows[0]?.operatorName || rows[0]?.name || rows[0]?.label || ""
    )
      .trim()
      .replace(/^\d+\s*-\s*/, "");
  }
  return String(machine?.operatorName || machine?.pic || "").trim();
}

function pickNum(...values) {
  for (const value of values) {
    const n = Number(value);
    if (Number.isFinite(n)) return n;
  }
  return 0;
}

function toMachinePayload(machines = []) {
  return (Array.isArray(machines) ? machines : [])
    .map((m) => {
      const runtimeSec = pickNum(
        m?.runtime,
        m?.runtimeSec,
        m?.RuntimeSec,
        m?.powerOnSeconds,
        m?.PowerOnSeconds,
        m?.totalPowerOn
      );
      const procSec = pickNum(
        m?.procTime,
        m?.procSec,
        m?.ProcSec,
        m?.ProcTimeSec,
        m?.productiveSeconds
      );
      const productivityPct = pickNum(
        m?.productivityPct,
        m?.productivity,
        m?.ProductivityPct
      );
      const lossTimeSec = pickNum(m?.lossTime, m?.lossTimeSec, m?.LossTimeSec);

      return {
        uuid: String(m?.uuid || m?.UUID || "").trim(),
        machineName: String(
          m?.displayMachineName ||
            m?.machineName ||
            m?.MachineName ||
            m?.nickName ||
            m?.NickName ||
            ""
        ).trim(),
        location: String(m?.location || m?.Location || "").trim(),
        macState: String(m?.macState ?? m?.MacState ?? "").trim(),
        status: String(m?.status || m?.Status || "").trim(),
        productivityPct,
        runtimeSec,
        procSec,
        lossTimeSec:
          lossTimeSec > 0
            ? lossTimeSec
            : Math.max(0, runtimeSec - procSec),
        operatorName: firstOperatorName(m),
        operatorNotes: collectMachineOperatorNotes(m),
      };
    })
    .filter((m) => m.uuid);
}

function yesterdayOf(dateText) {
  const raw = String(dateText || "").trim();
  if (!/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString().slice(0, 10);
  }
  const d = new Date(`${raw}T12:00:00`);
  d.setDate(d.getDate() - 1);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function extractMachines(payload) {
  if (Array.isArray(payload)) return payload;
  return (
    payload?.rows ||
    payload?.Rows ||
    payload?.machines ||
    payload?.Machines ||
    payload?.data?.rows ||
    payload?.data?.Rows ||
    payload?.data?.machines ||
    payload?.data?.Machines ||
    payload?.items ||
    payload?.Items ||
    []
  );
}

function readStore() {
  try {
    const raw = localStorage.getItem(SNAPSHOT_KEY);
    if (!raw) return { previous: null, current: null };
    const parsed = JSON.parse(raw);
    return {
      previous: parsed?.previous || null,
      current: parsed?.current || null,
    };
  } catch {
    return { previous: null, current: null };
  }
}

function writeStore(store) {
  try {
    localStorage.setItem(SNAPSHOT_KEY, JSON.stringify(store));
  } catch {
    // ignore
  }
}

function toSnapFromExecutive(data) {
  const health = data?.health || {};
  return {
    ts: Date.now(),
    at: data?.generatedAt || new Date().toLocaleString("id-ID"),
    workDate: String(data?.workDate || "").trim(),
    score: Number(health.score ?? 0) || 0,
    label: String(health.label || "").trim(),
    avgUtilPct: Number(health.avgUtilPct ?? 0) || 0,
    activeCount: Number(health.activeCount ?? 0) || 0,
    badCount: Number(health.badCount ?? 0) || 0,
    highCount: Number(health.highCount ?? 0) || 0,
    mediumCount: Number(health.mediumCount ?? 0) || 0,
  };
}

function deltaLine(label, before, after, unit = "", invertBetter = false) {
  const a = Number(after) || 0;
  const b = Number(before) || 0;
  const d = a - b;
  if (Math.abs(d) < 0.05) {
    return `- ${label}: tetap ${a}${unit}`;
  }
  const sign = d > 0 ? "+" : "";
  const good = invertBetter ? d < 0 : d > 0;
  const tone = good ? "membaik" : "memburuk";
  return `- ${label}: ${b}${unit} → ${a}${unit} (${sign}${Number(d.toFixed(1))}${unit}, ${tone})`;
}

function buildChangeSinceText(current, previous) {
  if (!current) {
    return "Data skor utilitas belum siap. Buka Presentasi atau refresh dulu.";
  }
  if (!previous) {
    return [
      "APA YANG BERUBAH SEJAK TADI",
      "",
      "Ini acuan pertama. Belum ada data sebelumnya untuk dibandingkan.",
      `Kondisi sekarang (${current.at}):`,
      `- Skor utilitas: ${current.score}/100 (${current.label || "-"})`,
      `- Util rata-rata: ${Number(current.avgUtilPct).toFixed(1)}%`,
      `- Mesin aktif: ${current.activeCount}`,
      `- BAD: ${current.badCount}`,
      `- Prioritas segera: ${current.highCount}`,
      "",
      "Cek lagi nanti (minimal ~5 menit) untuk melihat perubahan.",
    ].join("\n");
  }

  return [
    "APA YANG BERUBAH SEJAK TADI",
    "",
    `Sebelumnya: ${previous.at}`,
    `Sekarang: ${current.at}`,
    "",
    deltaLine("Skor utilitas", previous.score, current.score, ""),
    deltaLine("Util rata-rata", previous.avgUtilPct, current.avgUtilPct, "%"),
    deltaLine("Mesin aktif", previous.activeCount, current.activeCount, ""),
    deltaLine("Status BAD", previous.badCount, current.badCount, "", true),
    deltaLine("Prioritas segera", previous.highCount, current.highCount, "", true),
    "",
    `Ringkas: skor ${previous.score} → ${current.score} (${current.label || "-"}).`,
  ].join("\n");
}

function applySnapshotRotation(currentSnap) {
  const store = readStore();
  let previous = store.previous;
  const oldCurrent = store.current;

  if (!oldCurrent) {
    writeStore({ previous: null, current: currentSnap });
    return { previous: null, current: currentSnap };
  }

  const age = Number(currentSnap.ts || Date.now()) - Number(oldCurrent.ts || 0);
  if (age >= SNAPSHOT_MIN_GAP_MS) {
    previous = oldCurrent;
  } else if (!previous) {
    previous = oldCurrent;
  }

  writeStore({ previous, current: currentSnap });
  return { previous, current: currentSnap };
}

export function useAiExecutive() {
  const loading = ref(false);
  const errorMessage = ref("");
  const executive = ref(null);
  const presentationOpen = ref(false);
  const changeSinceText = ref("");

  async function refreshExecutive({
    date,
    shift = "",
    location = "",
    machinesToday = [],
    insight = null,
    silent = false,
  } = {}) {
    if (!silent) loading.value = true;
    errorMessage.value = "";

    const workDate = String(date || "").trim();
    const compareDate = yesterdayOf(workDate);
    let machinesYesterday = [];

    try {
      if (workDate) {
        const yRaw = await getProductivity(compareDate, { shift });
        machinesYesterday = toMachinePayload(extractMachines(yRaw));
      }
    } catch {
      machinesYesterday = [];
    }

    try {
      const data = await postAiExecutive({
        date: workDate,
        compareDate,
        shift,
        location,
        machinesToday: toMachinePayload(machinesToday),
        machinesYesterday,
        insight,
      });
      executive.value = data;

      const currentSnap = toSnapFromExecutive(data);
      const rotated = applySnapshotRotation(currentSnap);
      changeSinceText.value = buildChangeSinceText(rotated.current, rotated.previous);

      return data;
    } catch (err) {
      errorMessage.value = err?.message || "Gagal memuat AI Executive.";
      throw err;
    } finally {
      if (!silent) loading.value = false;
    }
  }

  function getChangeSinceReport() {
    if (changeSinceText.value) return changeSinceText.value;
    const store = readStore();
    if (store.current) {
      return buildChangeSinceText(store.current, store.previous);
    }
    return "Belum ada data perubahan. Refresh Presentasi / dashboard dulu.";
  }

  function openPresentation() {
    presentationOpen.value = true;
  }

  function closePresentation() {
    presentationOpen.value = false;
  }

  function exportBriefing({ asMarkdown = false } = {}) {
    const text = String(executive.value?.briefing || "").trim();
    if (!text) {
      throw new Error("Briefing belum tersedia. Refresh Presentasi dulu.");
    }
    const stamp = String(executive.value?.generatedAt || "")
      .replace(/[^\d]/g, "")
      .slice(0, 14);
    const ext = asMarkdown ? "md" : "txt";
    const filename = `briefing-atasan-${stamp || Date.now()}.${ext}`;
    const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    return filename;
  }

  async function copyBriefing() {
    const text = String(executive.value?.briefing || "").trim();
    if (!text) {
      throw new Error("Briefing belum tersedia. Refresh Presentasi dulu.");
    }
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return;
    }
    const ta = document.createElement("textarea");
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand("copy");
    ta.remove();
  }

  return {
    loading,
    errorMessage,
    executive,
    presentationOpen,
    changeSinceText,
    refreshExecutive,
    getChangeSinceReport,
    openPresentation,
    closePresentation,
    exportBriefing,
    copyBriefing,
    yesterdayOf,
  };
}
