<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";

import ExecutiveSummary from "../components/dashboard/ExecutiveSummary.vue";
import KpiCards from "../components/dashboard/KpiCards.vue";
import RankingPanel from "../components/dashboard/RankingPanel.vue";
import StatusDistribution from "../components/dashboard/StatusDistribution.vue";
import AttentionPanel from "../components/dashboard/AttentionPanel.vue";
import MachineTable from "../components/dashboard/MachineTable.vue";
import MachineEditModal from "../components/dashboard/MachineEditModal.vue";
import MachineQrModal from "../components/dashboard/MachineQrModal.vue";

import { deleteMachineSetting, saveMachineSetting } from "../api/machineApi";
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

let noticeTimer = null;

const {
  machines,
  loading,
  errorMessage,
  loadDashboard,
  normalizeRows,
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

      // Operator aktif
      m.pic,
      m.operatorNik,
      m.operatorName,
      m.operatorSubText,
      m.operatorLoginText,
      m.operatorActiveText,

      // Operator note
      m.operatorNote,
      m.operatorNotes,
      m.spv,

      // Proses/style operator
      m.operatorProcessName,
      m.operatorStyleName,

      // Data mesin
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
  }, 3000);
}

async function refreshDashboardByDate() {
  await loadDashboard(localDate.value);
  connect(localDate.value);
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

function csvCell(value) {
  const text = String(value ?? "");
  return `"${text.replace(/"/g, '""')}"`;
}

function downloadCSV() {
  const rows = [
    [
      "Mesin",
      "OriginalName",
      "Location",
      "Operator Aktif",
      "Operator Login Info",
      "Operator Note",
      "Operator NIK",
      "Operator Name",
      "Operator Process",
      "Operator Style",
      "IP",
      "Power On Duration",
      "Running Time",
      "Loss Time",
      "Produktivitas",
      "Status",
      "Output",
      "Abnormal",
      "AvgCT",
      "MaxCT",
      "Alarm",
      "Program",
      "UUID",
      "Table",
    ],
    ...filteredMachines.value.map((m) => {
      const runtime = Number(m.runtime || 0);
      const procTime = Number(m.procTime || 0);
      const lossTime = Math.max(0, runtime - procTime);

      return [
        m.machineName,
        m.originalMachineName,
        m.location,
        m.pic || "",
        m.operatorSubText || "",
        m.operatorNote || m.spv || "",
        m.operatorNik || "",
        m.operatorName || "",
        m.operatorProcessName || "",
        m.operatorStyleName || "",
        m.ip,
        runtime,
        procTime,
        lossTime,
        Number(m.productivity || 0).toFixed(2),
        m.status,
        m.output,
        m.abnormal,
        m.avgCT,
        m.maxCT,
        m.alarm,
        m.program,
        m.uuid,
        m.tableName,
      ];
    }),
  ];

  const csv = rows.map((r) => r.map(csvCell).join(",")).join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");

  a.href = url;
  a.download = `dashboard-mesin-${localDate.value}.csv`;
  a.click();

  URL.revokeObjectURL(url);
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
      v-model:selected-date="localDate"
      v-model:keyword="keyword"
      :machines="filteredMachines"
      :loading="loading"
      :page-size="10"
      :show-actions="isAdmin"
      @date-change="refreshDashboardByDate"
      @download="downloadCSV"
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