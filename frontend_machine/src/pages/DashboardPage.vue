<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import * as XLSX from "xlsx";

import ExecutiveSummary from "../components/dashboard/ExecutiveSummary.vue";
import KpiCards from "../components/dashboard/KpiCards.vue";
import RankingPanel from "../components/dashboard/RankingPanel.vue";
import StatusDistribution from "../components/dashboard/StatusDistribution.vue";
import AttentionPanel from "../components/dashboard/AttentionPanel.vue";
import MachineTable from "../components/dashboard/MachineTable.vue";
import MachineEditModal from "../components/dashboard/MachineEditModal.vue";
import MachineQrModal from "../components/dashboard/MachineQrModal.vue";

import {
  deleteMachineSetting,
  saveMachineSetting,
  getProductivity,
} from "../api/machineApi";
import { useDashboard } from "../composables/useDashboard";
import { useProductivitySocket } from "../composables/useProductivitySocket";
import { getInitialAdminMode } from "../utils/adminMode";

const props = defineProps({
  selectedDate: {
    type: String,
    required: true,
  },
});

const emit = defineEmits([
  "update:selectedDate",
  "socket-status-change",
  "admin-mode-change",
]);

const keyword = ref("");
const editModalOpen = ref(false);
const qrModalOpen = ref(false);
const selectedEditMachine = ref(null);
const notice = ref("");
const noticeType = ref("ok");
const isAdmin = ref(false);
const rangeExporting = ref(false);

const startDate = ref(props.selectedDate || "");
const endDate = ref(props.selectedDate || "");

let noticeTimer = null;

const {
  machines,
  loading,
  errorMessage,
  loadDashboard,
  normalizeRows,
  machineSettings,
} = useDashboard();

const { socketStatus, connect, close } = useProductivitySocket(normalizeRows);

const localDate = computed({
  get: () => props.selectedDate,
  set: (value) => emit("update:selectedDate", value),
});

watch(
  socketStatus,
  (value) => {
    emit("socket-status-change", value);
  },
  { immediate: true }
);

watch(
  isAdmin,
  (value) => {
    emit("admin-mode-change", value);
  },
  { immediate: true }
);

watch(
  () => props.selectedDate,
  (value) => {
    if (!value) return;

    startDate.value = value;
    endDate.value = value;
  },
  { immediate: true }
);

watch(
  startDate,
  (value) => {
    if (!endDate.value) {
      endDate.value = value;
    }
  }
);

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

const filteredMachines = computed(() => {
  const key = keyword.value.toLowerCase().trim();

  if (!key) return machines.value;

  return machines.value.filter((m) => {
    return [
      m.machineName,
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

function showNotice(message, type = "ok") {
  notice.value = message;
  noticeType.value = type;

  if (noticeTimer) {
    clearTimeout(noticeTimer);
  }

  noticeTimer = setTimeout(() => {
    notice.value = "";
  }, 3500);
}

async function refreshDashboardByDate() {
  if (startDate.value) {
    localDate.value = startDate.value;
  }

  await loadDashboard(startDate.value || localDate.value);
  connect(startDate.value || localDate.value);
}

function openEditModal(machine) {
  if (!isAdmin.value) {
    showNotice("Akses edit hanya untuk admin.", "error");
    return;
  }

  selectedEditMachine.value = machine;
  editModalOpen.value = true;
}

function closeEditModal() {
  editModalOpen.value = false;
  selectedEditMachine.value = null;
}

function openQrModal() {
  if (!isAdmin.value) {
    showNotice("Akses generate QR hanya untuk admin.", "error");
    return;
  }

  qrModalOpen.value = true;
}

function closeQrModal() {
  qrModalOpen.value = false;
}

async function handleSaveMachineSetting(machine) {
  if (!machine?.uuid) return;

  if (!isAdmin.value) {
    showNotice("Akses simpan hanya untuk admin.", "error");
    return;
  }

  try {
    await saveMachineSetting({
      uuid: machine.uuid,
      customName: machine.customName || "",
      location: machine.location || "",
      pic: machine.pic || "",
      spv: machine.spv || "",
    });

    showNotice("Setting manual berhasil disimpan.", "ok");
    closeEditModal();

    await loadDashboard(localDate.value);
  } catch (err) {
    showNotice(`Gagal menyimpan setting mesin: ${err.message}`, "error");
  }
}

async function handleDeleteMachineSetting(uuid) {
  if (!isAdmin.value) {
    showNotice("Akses hapus hanya untuk admin.", "error");
    return;
  }

  if (!uuid) return;

  try {
    await deleteMachineSetting(uuid);

    showNotice("Setting manual berhasil dihapus.", "ok");
    closeEditModal();

    await loadDashboard(localDate.value);
  } catch (err) {
    showNotice(`Gagal hapus: ${err.message}`, "error");
  }
}

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function getVal(obj, ...keys) {
  for (const key of keys) {
    if (obj && obj[key] !== undefined && obj[key] !== null) {
      return obj[key];
    }
  }

  return undefined;
}

function extractRows(data) {
  if (Array.isArray(data)) return data;

  return (
    data?.rows ||
    data?.Rows ||
    data?.data ||
    data?.Data ||
    data?.items ||
    data?.Items ||
    data?.machines ||
    data?.Machines ||
    []
  );
}

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function formatExcelDate(dateText) {
  const text = String(dateText || "").trim();

  if (!text) return "";

  const parts = text.split("-");

  if (parts.length === 3) {
    const year = parts[0];
    const month = Number(parts[1]);
    const day = Number(parts[2]);

    return `${day}/${month}/${year}`;
  }

  return text;
}

function formatSeconds(seconds) {
  const totalSeconds = Math.max(0, Math.floor(Number(seconds || 0)));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  return `${hours}h ${minutes}m`;
}

function statusFromProductivity(productivity) {
  const value = Number(productivity || 0);

  if (value >= 90) return "GOOD";
  if (value >= 80) return "NORMAL";
  return "BAD";
}

function toIsoDate(date) {
  const offset = date.getTimezoneOffset();
  const local = new Date(date.getTime() - offset * 60000);

  return local.toISOString().slice(0, 10);
}

function getOrderedRange() {
  const start = String(startDate.value || localDate.value || "").trim();
  const end = String(endDate.value || start || "").trim();

  if (!start && !end) {
    return {
      start: "",
      end: "",
    };
  }

  if (start && end && end < start) {
    return {
      start: end,
      end: start,
    };
  }

  return {
    start,
    end: end || start,
  };
}

function getWeekdaysBetween(startDateText, endDateText) {
  const start = new Date(`${startDateText}T00:00:00`);
  const end = new Date(`${endDateText}T00:00:00`);
  const dates = [];

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return dates;
  }

  const cursor = new Date(start);

  while (cursor <= end) {
    const day = cursor.getDay();

    if (day >= 1 && day <= 5) {
      dates.push(toIsoDate(cursor));
    }

    cursor.setDate(cursor.getDate() + 1);
  }

  return dates;
}

function getDayName(dateText) {
  const d = new Date(`${dateText}T00:00:00`);

  return d.toLocaleDateString("id-ID", {
    weekday: "long",
  });
}

function setWorksheetWidth(worksheet, widths) {
  worksheet["!cols"] = widths.map((wch) => ({ wch }));
}

function setAutoFilter(worksheet, rowCount, colCount) {
  if (!rowCount || !colCount) return;

  const lastCol = XLSX.utils.encode_col(colCount - 1);
  const lastRow = rowCount + 1;

  worksheet["!autofilter"] = {
    ref: `A1:${lastCol}${lastRow}`,
  };
}

function getManualSettingByUuid(uuid) {
  const key = normalizeText(uuid);
  return machineSettings?.value?.get(key) || null;
}

function normalizeRangeItem(row, dateText) {
  const uuid = String(getVal(row, "uuid", "UUID") || "");
  const setting = getManualSettingByUuid(uuid);

  const procTime = toNumber(
    getVal(
      row,
      "procSec",
      "ProcSec",
      "procTimeSec",
      "ProcTimeSec",
      "procTime",
      "ProcTime",
      "process_time",
      "ProcessTime"
    ) || 0
  );

  const runtime = toNumber(
    getVal(
      row,
      "runtimeSec",
      "RuntimeSec",
      "runtime",
      "Runtime",
      "RunTime",
      "runTime"
    ) || 0
  );

  const lossTime = Math.max(0, runtime - procTime);
  const productivity =
    runtime > 0 ? Math.min((procTime / runtime) * 100, 100) : 0;

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
  );

  const machineName = setting?.customName || backendName;

  const locationFromApi = String(getVal(row, "location", "Location") || "");
  const location = setting?.location || locationFromApi || "-";

  return {
    dateText,
    tanggal: formatExcelDate(dateText),
    hari: getDayName(dateText),
    mesin: machineName,
    location,
    runtime,
    procTime,
    lossTime,
    productivity: Number(productivity.toFixed(2)),
    status: statusFromProductivity(productivity),
  };
}

function buildRangeDetailRows(items) {
  return items
    .map((item) => {
      return {
        Tanggal: item.tanggal,
        Hari: item.hari,
        Mesin: item.mesin,
        Location: item.location,
        "Power On Duration": formatSeconds(item.runtime),
        "Running Time": formatSeconds(item.procTime),
        "Loss Time": formatSeconds(item.lossTime),
        Produktivitas: item.productivity,
        Status: item.status,
      };
    })
    .sort((a, b) => {
      const dateA = String(a.Tanggal || "");
      const dateB = String(b.Tanggal || "");
      const locCompare = String(a.Location).localeCompare(String(b.Location));

      if (dateA !== dateB) return dateA.localeCompare(dateB);
      if (locCompare !== 0) return locCompare;

      return String(a.Mesin).localeCompare(String(b.Mesin));
    });
}

function buildRangeSummaryRows(items, rangeStart, rangeEnd) {
  const map = new Map();

  items.forEach((item) => {
    const key = `${item.location}||${item.mesin}`;

    if (!map.has(key)) {
      map.set(key, {
        periode: `${formatExcelDate(rangeStart)} - ${formatExcelDate(rangeEnd)}`,
        location: item.location,
        mesin: item.mesin,
        hariData: 0,
        totalPowerOn: 0,
        totalRunning: 0,
        totalLoss: 0,
      });
    }

    const current = map.get(key);

    current.hariData += 1;
    current.totalPowerOn += Number(item.runtime || 0);
    current.totalRunning += Number(item.procTime || 0);
    current.totalLoss += Number(item.lossTime || 0);
  });

  return [...map.values()]
    .map((item) => {
      const productivity =
        item.totalPowerOn > 0
          ? Math.min((item.totalRunning / item.totalPowerOn) * 100, 100)
          : 0;

      return {
        Periode: item.periode,
        Location: item.location,
        Mesin: item.mesin,
        "Hari Data": item.hariData,
        "Total Power On Duration": formatSeconds(item.totalPowerOn),
        "Total Running Time": formatSeconds(item.totalRunning),
        "Total Loss Time": formatSeconds(item.totalLoss),
        Produktivitas: Number(productivity.toFixed(2)),
        Status: statusFromProductivity(productivity),
        __lossSec: item.totalLoss,
      };
    })
    .sort((a, b) => {
      const locCompare = String(a.Location).localeCompare(String(b.Location));
      if (locCompare !== 0) return locCompare;

      return Number(a.Produktivitas || 0) - Number(b.Produktivitas || 0);
    });
}

function buildBadPriorityRows(summaryRows) {
  const badRows = summaryRows
    .filter((row) => row.Status === "BAD")
    .sort((a, b) => Number(a.Produktivitas || 0) - Number(b.Produktivitas || 0))
    .map((row, index) => {
      let rekomendasi = "Cek penyebab running time rendah.";

      if (Number(row.__lossSec || 0) > 14400) {
        rekomendasi = "Loss time tinggi. Cek idle, operator, dan kondisi mesin.";
      }

      return {
        Rank: index + 1,
        Location: row.Location,
        Mesin: row.Mesin,
        Produktivitas: row.Produktivitas,
        "Total Loss Time": row["Total Loss Time"],
        Rekomendasi: rekomendasi,
      };
    });

  if (badRows.length) {
    return badRows;
  }

  return [
    {
      Rank: "",
      Location: "",
      Mesin: "Tidak ada mesin BAD",
      Produktivitas: "",
      "Total Loss Time": "",
      Rekomendasi: "Semua mesin tidak berstatus BAD pada periode ini.",
    },
  ];
}

function cleanSummaryRows(summaryRows) {
  return summaryRows.map((row) => {
    const cleaned = { ...row };
    delete cleaned.__lossSec;
    return cleaned;
  });
}

async function downloadExcel() {
  if (rangeExporting.value) return;

  const range = getOrderedRange();

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
        const data = await getProductivity(dateText);
        const rows = extractRows(data);

        return rows.map((row) => normalizeRangeItem(row, dateText));
      })
    );

    const items = results.flat();

    if (!items.length) {
      showNotice("Data export tidak ditemukan.", "error");
      return;
    }

    const summaryRowsRaw = buildRangeSummaryRows(items, range.start, range.end);
    const summaryRows = cleanSummaryRows(summaryRowsRaw);
    const detailRows = buildRangeDetailRows(items);
    const badPriorityRows = buildBadPriorityRows(summaryRowsRaw);

    const workbook = XLSX.utils.book_new();

    const summarySheet = XLSX.utils.json_to_sheet(summaryRows);
    setWorksheetWidth(summarySheet, [
      24, // Periode
      18, // Location
      42, // Mesin
      12, // Hari Data
      24, // Total Power On Duration
      22, // Total Running Time
      18, // Total Loss Time
      16, // Produktivitas
      12, // Status
    ]);
    setAutoFilter(summarySheet, summaryRows.length, 9);

    const detailSheet = XLSX.utils.json_to_sheet(detailRows);
    setWorksheetWidth(detailSheet, [
      12, // Tanggal
      14, // Hari
      42, // Mesin
      18, // Location
      22, // Power On Duration
      18, // Running Time
      16, // Loss Time
      16, // Produktivitas
      12, // Status
    ]);
    setAutoFilter(detailSheet, detailRows.length, 9);

    const badSheet = XLSX.utils.json_to_sheet(badPriorityRows);
    setWorksheetWidth(badSheet, [
      8, // Rank
      18, // Location
      42, // Mesin
      16, // Produktivitas
      18, // Total Loss Time
      48, // Rekomendasi
    ]);
    setAutoFilter(badSheet, badPriorityRows.length, 6);

    XLSX.utils.book_append_sheet(workbook, summarySheet, "Range Summary");
    XLSX.utils.book_append_sheet(workbook, detailSheet, "Daily Detail");
    XLSX.utils.book_append_sheet(workbook, badSheet, "BAD Priority");

    XLSX.writeFile(
      workbook,
      `productivity-range-${range.start}_${range.end}.xlsx`
    );

    showNotice("Export Excel berhasil dibuat.", "ok");
  } catch (err) {
    showNotice(`Gagal export Excel: ${err.message}`, "error");
  } finally {
    rangeExporting.value = false;
  }
}

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());

  emit("admin-mode-change", isAdmin.value);
  emit("socket-status-change", socketStatus.value);

  await loadDashboard(localDate.value);
  connect(localDate.value);
});

onUnmounted(() => {
  close();

  if (noticeTimer) {
    clearTimeout(noticeTimer);
  }
});
</script>

<template>
  <section class="dashboard-page">
    <div v-if="loading" class="alert">Mengambil data dari backend...</div>
    <div v-if="rangeExporting" class="alert">
      Mengambil data range dan membuat Excel...
    </div>
    <div v-if="errorMessage" class="alert error">{{ errorMessage }}</div>
    <div v-if="notice" class="alert" :class="{ error: noticeType === 'error' }">
      {{ notice }}
    </div>

    <ExecutiveSummary
      :message="executiveMessage"
      :best-machine="bestMachine"
      :worst-machine="worstMachine"
    />

    <KpiCards :summary="summary" />

    <section class="content-grid" :class="{ 'user-view': !isAdmin }">
      <RankingPanel
        class="panel-large"
        :machines="rankingMachines"
        :selected-date="localDate"
      />

      <div class="right-stack">
        <StatusDistribution
          v-if="isAdmin"
          :summary="summary"
          :donut-style="donutStyle"
        />

        <AttentionPanel :machines="attentionMachines" />
      </div>
    </section>

    <section v-if="isAdmin" class="qr-action-panel">
      <div>
        <h3>QR Code Mesin</h3>
        <p>
          Generate QR semua mesin untuk login operator dan menu keterangan
          operator.
        </p>
      </div>

      <button type="button" class="btn-qr" @click="openQrModal">
        Generate QR Mesin
      </button>
    </section>

    <MachineTable
      v-model:selected-date="startDate"
      v-model:start-date="startDate"
      v-model:end-date="endDate"
      v-model:keyword="keyword"
      :machines="filteredMachines"
      :loading="loading || rangeExporting"
      :page-size="10"
      :show-actions="isAdmin"
      @date-change="refreshDashboardByDate"
      @download="downloadExcel"
      @edit="openEditModal"
    />

    <MachineEditModal
      v-if="isAdmin"
      :show="editModalOpen"
      :machine="selectedEditMachine"
      @close="closeEditModal"
      @save="handleSaveMachineSetting"
      @delete="handleDeleteMachineSetting"
    />

    <MachineQrModal
      v-if="isAdmin"
      :show="qrModalOpen"
      :machines="machines"
      @close="closeQrModal"
    />
  </section>
</template>

<style scoped>
.dashboard-page {
  display: grid;
  gap: 18px;
}

.alert {
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  border: 1px solid #dbeafe;
  padding: 14px 16px;
  border-radius: 18px;
  font-weight: 800;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.alert.error {
  color: #b91c1c;
  background: linear-gradient(135deg, #fff7f7, #fee2e2);
  border-color: #fecaca;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.85fr);
  gap: 18px;
  align-items: stretch;
  min-width: 0;
}

.content-grid.user-view {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.right-stack {
  display: grid;
  gap: 18px;
  height: 100%;
  min-width: 0;
  align-content: stretch;
}

.right-stack > * {
  height: 100%;
}

.panel-large {
  height: 100%;
  min-width: 0;
}

.qr-action-panel {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 18px;
  padding: 18px 20px;
  border-radius: 22px;
  background: linear-gradient(135deg, #eff6ff, #ffffff);
  border: 1px solid #bfdbfe;
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.05);
}

.qr-action-panel h3 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
  letter-spacing: -0.02em;
}

.qr-action-panel p {
  margin: 5px 0 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.btn-qr {
  border: 0;
  border-radius: 15px;
  padding: 13px 18px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #ffffff;
  font-size: 14px;
  font-weight: 1000;
  cursor: pointer;
  white-space: nowrap;
  box-shadow: 0 14px 24px rgba(37, 99, 235, 0.24);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    filter 0.2s ease;
}

.btn-qr:hover {
  transform: translateY(-1px);
  filter: brightness(1.03);
  box-shadow: 0 18px 28px rgba(37, 99, 235, 0.3);
}

.btn-qr:active {
  transform: translateY(0);
}

@media (max-width: 1280px) {
  .content-grid,
  .content-grid.user-view {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .qr-action-panel {
    flex-direction: column;
    align-items: stretch;
  }

  .btn-qr {
    width: 100%;
  }
}
</style>