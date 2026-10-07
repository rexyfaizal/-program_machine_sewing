<script setup>
import { computed } from "vue";
import {
  collectMachineOperatorNotes,
  getOperatorNoteItems,
} from "../../utils/machineTableFormat";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  item: {
    type: Object,
    default: null,
  },
  machine: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["close", "filter", "open-chart", "ask-ai"]);

const THRESHOLD_LOW = 55;
const THRESHOLD_BAD = 80;

const machineName = computed(() => {
  const m = props.machine || {};
  return (
    String(props.item?.machineName || "").trim() ||
    m.displayMachineName ||
    m.machineName ||
    m.nickName ||
    m.customName ||
    props.item?.uuid ||
    "-"
  );
});

const location = computed(() =>
  String(props.item?.location || props.machine?.location || "").trim() || "-"
);

const operatorName = computed(() => {
  const fromItem = String(props.item?.operatorName || "").trim();
  if (fromItem) return fromItem;
  const rows = Array.isArray(props.machine?.operatorRows)
    ? props.machine.operatorRows
    : [];
  if (rows.length) {
    return String(rows[0]?.operatorName || rows[0]?.name || "").trim() || "-";
  }
  return String(props.machine?.operatorName || "").trim() || "-";
});

const utilPct = computed(() => {
  const fromItem = Number(props.item?.utilPct);
  if (Number.isFinite(fromItem) && fromItem > 0) return fromItem;
  const fromMachine = Number(
    props.machine?.productivityPct ?? props.machine?.productivity ?? 0
  );
  return Number.isFinite(fromMachine) ? fromMachine : 0;
});

const runtimeSec = computed(
  () => Number(props.machine?.runtime ?? props.machine?.runtimeSec ?? 0) || 0
);
const procSec = computed(
  () => Number(props.machine?.procTime ?? props.machine?.procSec ?? 0) || 0
);
const lossTimeSec = computed(
  () => Number(props.machine?.lossTime ?? props.machine?.lossTimeSec ?? 0) || 0
);

const status = computed(() =>
  String(props.machine?.status || props.item?.priority || "").trim().toUpperCase() ||
  "-"
);

const noteSummary = computed(() => {
  // Prioritas: note yang sama seperti kolom Operator Note di tabel dashboard.
  const dashNotes = collectMachineOperatorNotes(props.machine);
  if (dashNotes) return dashNotes;

  const note = String(props.item?.noteSummary || "").trim();
  if (note) return note;

  const lossLabel = String(props.item?.lossReasonLabel || "").trim();
  const lossCode = String(props.item?.lossReasonCode || "").trim();
  if (lossLabel || lossCode) return lossLabel || lossCode;

  return "";
});

const hasNote = computed(() => Boolean(noteSummary.value));
const noteItems = computed(() => getOperatorNoteItems(props.machine || {}));

const rows = computed(() => {
  const util = utilPct.value;
  const utilRead =
    util <= 0
      ? "Belum terukur"
      : util < THRESHOLD_LOW
        ? `Di bawah ambang Insight (< ${THRESHOLD_LOW}%)`
        : util < THRESHOLD_BAD
          ? "Di bawah standar GOOD (≥ 80%)"
          : "Masih di kisaran wajar";

  const lossRatio =
    runtimeSec.value > 0
      ? (lossTimeSec.value / runtimeSec.value) * 100
      : 0;

  return [
    {
      metric: "Utilisasi",
      value: util > 0 ? `${util.toFixed(1)}%` : "-",
      read: utilRead,
      tone: util > 0 && util < THRESHOLD_LOW ? "bad" : util < THRESHOLD_BAD ? "warn" : "ok",
    },
    {
      metric: "Power On",
      value: formatDuration(runtimeSec.value),
      read: runtimeSec.value > 0 ? "Mesin sempat/nyala terukur" : "Belum ada runtime",
      tone: runtimeSec.value > 0 ? "ok" : "mute",
    },
    {
      metric: "Process",
      value: formatDuration(procSec.value),
      read:
        runtimeSec.value > 0 && procSec.value < runtimeSec.value * 0.55
          ? "Process jauh di bawah Power On"
          : procSec.value > 0
            ? "Ada process time"
            : "Belum ada process",
      tone:
        runtimeSec.value > 0 && procSec.value < runtimeSec.value * 0.55
          ? "bad"
          : "ok",
    },
    {
      metric: "Loss sistem",
      value: formatDuration(lossTimeSec.value),
      read:
        lossRatio >= 50
          ? `~${lossRatio.toFixed(0)}% dari Power On`
          : lossTimeSec.value > 0
            ? "Ada selisih Power On − Process"
            : "Tidak ada loss sistem",
      tone: lossRatio >= 50 ? "bad" : lossTimeSec.value > 0 ? "warn" : "ok",
    },
    {
      metric: "Status dashboard",
      value: status.value,
      read:
        status.value === "BAD"
          ? "Produktivitas di bawah ambang BAD"
          : status.value === "NORMAL"
            ? "Perlu perhatian"
            : status.value === "GOOD"
              ? "Dalam standar"
              : "Mengikuti prioritas Insight",
      tone: status.value === "BAD" ? "bad" : status.value === "NORMAL" ? "warn" : "mute",
    },
    {
      metric: "Keterangan / loss",
      value: hasNote.value ? "Ada" : "Tidak ada",
      read: hasNote.value
        ? noteSummary.value
        : "Celah data — belum bisa pastikan penyebab",
      tone: hasNote.value ? "ok" : "bad",
    },
    ...(Number(props.item?.brokenCount7d || 0) >= 3
      ? [
          {
            metric: "Sering rusak",
            value: `${Number(props.item.brokenCount7d)}x / 7 hari`,
            read: "Riwayat MACHINE_BROKEN — pantau teknis",
            tone: "bad",
          },
        ]
      : []),
  ];
});

const conclusion = computed(() => {
  if (hasNote.value) {
    return `Ada keterangan operator: ${noteSummary.value}. Utilisasi ${utilPct.value.toFixed(1)}% — tindaklanjuti sesuai isi note, bukan anggap celah data.`;
  }
  const decision = String(props.item?.decision || "").trim();
  if (decision && !/tanpa keterangan|celah data/i.test(decision)) {
    return decision;
  }
  const evidence = String(props.item?.evidence || "").trim();
  if (evidence && !/Note\/loss aktif: tidak ada|Keterangan\/loss aktif: tidak ada/i.test(evidence)) {
    return evidence;
  }
  if (!hasNote.value && utilPct.value > 0 && utilPct.value < THRESHOLD_LOW) {
    return `Utilisasi ${utilPct.value.toFixed(1)}% tanpa keterangan — lengkapi note loss dulu sebelum menafsirkan penyebab.`;
  }
  return "Ringkasan metrik di atas diambil dari data dashboard + Insight aktif.";
});

const step = computed(() => {
  const code = String(props.item?.code || "").toUpperCase();
  const broken = Number(props.item?.brokenCount7d || 0);
  if (code === "FREQUENT_BROKEN" || broken >= 3) {
    return (
      String(props.item?.improvement || "").trim() ||
      `Jadwalkan cek teknis preventive sebelum shift berikutnya (sering rusak ${broken}x/7 hari).`
    );
  }
  if (hasNote.value) {
    const op = operatorName.value !== "-" ? operatorName.value : "operator";
    return `Verifikasi ke ${op}: pastikan keterangan "${noteSummary.value}" sudah ditindaklanjuti (util ${utilPct.value.toFixed(1)}%).`;
  }
  return (
    String(props.item?.improvement || props.item?.decision || "").trim() ||
    "Monitor dan follow-up utilisasi mesin."
  );
});

function formatDuration(sec) {
  const n = Number(sec || 0);
  if (!Number.isFinite(n) || n <= 0) return "0 menit";
  const h = Math.floor(n / 3600);
  const m = Math.floor((n % 3600) / 60);
  if (h > 0) return `${h}j ${m}m`;
  if (m === 0 && n > 0) return `${Math.floor(n)} dtk`;
  return `${m} menit`;
}

function close() {
  emit("close");
}

function onFilter() {
  emit("filter", props.item);
}

function onOpenChart() {
  emit("open-chart", props.item);
}

function onAskAi() {
  emit("ask-ai", props.item);
}
</script>

<template>
  <div v-if="show && item" class="why-overlay" @click.self="close">
    <div class="why-modal" role="dialog" aria-modal="true" aria-labelledby="why-title">
      <header class="why-head">
        <div>
          <p class="why-kicker">Kenapa jelek</p>
          <h3 id="why-title">{{ machineName }}</h3>
          <p class="why-meta">
            {{ location }}
            <span v-if="operatorName !== '-'"> · {{ operatorName }}</span>
          </p>
        </div>
        <button type="button" class="why-close" aria-label="Tutup" @click="close">✕</button>
      </header>

      <div class="why-table-wrap">
        <table class="why-table">
          <thead>
            <tr>
              <th>Metrik</th>
              <th>Nilai</th>
              <th>Baca</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.metric">
              <td>{{ row.metric }}</td>
              <td>
                <strong :class="`tone-${row.tone}`">{{ row.value }}</strong>
              </td>
              <td>{{ row.read }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="noteItems.length" class="why-notes">
        <span class="why-label">Keterangan di dashboard</span>
        <ul>
          <li v-for="(note, idx) in noteItems" :key="note.id || idx">
            {{
              [
                [note.reasonName, note.body].filter(Boolean).join(" "),
                note.timeRange,
              ]
                .filter(Boolean)
                .join(" | ") || note.text
            }}
          </li>
        </ul>
      </div>

      <div class="why-conclusion">
        <span class="why-label">Kesimpulan</span>
        <p>{{ conclusion }}</p>
      </div>

      <div class="why-step">
        <span class="why-label why-label--mute">Langkah</span>
        <p>{{ step }}</p>
      </div>

      <footer class="why-actions">
        <button
          v-if="location && location !== '-'"
          type="button"
          class="why-btn"
          @click="onFilter"
        >
          Filter
        </button>
        <button
          v-if="item?.uuid"
          type="button"
          class="why-btn"
          @click="onOpenChart"
        >
          Chart
        </button>
        <button type="button" class="why-btn why-btn--main" @click="onAskAi">
          Tanya AI
        </button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.why-overlay {
  position: fixed;
  inset: 0;
  z-index: 13050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(15, 23, 42, 0.55);
}

.why-modal {
  width: min(640px, 100%);
  max-height: 90vh;
  overflow: auto;
  padding: 22px 24px;
  border-radius: 20px;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 42%);
  border: 1px solid #93c5fd;
  box-shadow: 0 28px 60px rgba(37, 99, 235, 0.22);
}

.why-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.why-kicker {
  margin: 0 0 4px;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #2563eb;
}

.why-head h3 {
  margin: 0;
  font-size: 18px;
  line-height: 1.3;
  color: #1e3a8a;
  font-weight: 800;
}

.why-meta {
  margin: 6px 0 0;
  font-size: 13px;
  color: #475569;
  font-weight: 500;
}

.why-close {
  width: 34px;
  height: 34px;
  border: 0;
  border-radius: 10px;
  background: #dbeafe;
  color: #1e3a8a;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  flex-shrink: 0;
}

.why-close:hover {
  background: #bfdbfe;
}

.why-table-wrap {
  border: 1px solid #bfdbfe;
  border-radius: 14px;
  overflow: hidden;
  background: #fff;
}

.why-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.why-table th,
.why-table td {
  padding: 10px 12px;
  text-align: left;
  vertical-align: top;
  border-top: 1px solid #e2e8f0;
}

.why-table thead th {
  border-top: 0;
  background: #f8fafc;
  color: #64748b;
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.why-table tbody td:first-child {
  color: #334155;
  font-weight: 600;
  width: 28%;
}

.why-table tbody td:nth-child(2) {
  width: 22%;
}

.why-table tbody td:last-child {
  color: #64748b;
  font-weight: 400;
  line-height: 1.45;
}

.tone-bad {
  color: #b91c1c;
  font-weight: 700;
}

.tone-warn {
  color: #c2410c;
  font-weight: 700;
}

.tone-ok {
  color: #1e40af;
  font-weight: 700;
}

.tone-mute {
  color: #64748b;
  font-weight: 600;
}

.why-notes {
  margin-top: 14px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #fff;
  border: 1px solid #bfdbfe;
}

.why-notes ul {
  margin: 0;
  padding-left: 18px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.why-notes li {
  font-size: 13px;
  line-height: 1.45;
  color: #334155;
  font-weight: 400;
}

.why-conclusion {
  margin-top: 14px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #eff6ff;
  border: 1px solid #93c5fd;
}

.why-step {
  margin-top: 12px;
}

.why-label {
  display: block;
  margin-bottom: 4px;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #2563eb;
}

.why-label--mute {
  color: #94a3b8;
  font-weight: 700;
}

.why-conclusion p {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #1e40af;
  font-weight: 600;
}

.why-step p {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.55;
  color: #475569;
  font-weight: 400;
}

.why-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.why-btn {
  border: 1px solid #93c5fd;
  background: #fff;
  color: #1d4ed8;
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.why-btn--main {
  background: #2563eb;
  border-color: #2563eb;
  color: #fff;
}

.why-btn:hover {
  background: #dbeafe;
}

.why-btn--main:hover {
  background: #1d4ed8;
  border-color: #1d4ed8;
  color: #fff;
}

@media (max-width: 560px) {
  .why-table th:last-child,
  .why-table td:last-child {
    display: none;
  }
}
</style>
