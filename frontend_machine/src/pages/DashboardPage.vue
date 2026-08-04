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

import {
  deleteMachineSetting,
  saveMachineSetting,
} from "../api/machineApi";

import { useDashboard } from "../composables/useDashboard";
import { useDashboardExcelExport } from "../composables/useDashboardExcelExport";
import { useDashboardFilters } from "../composables/useDashboardFilters";
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
  machineSettings,
  shiftConfigMap,
} = useDashboard();

const { socketStatus, connect, close } = useProductivitySocket(normalizeRows);

const localDate = computed({
  get: () => props.selectedDate,
  set: (value) => emit("update:selectedDate", value),
});

const {
  keyword,
  locationFilter,
  shiftFilter,
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
} = useDashboardFilters(machines, shiftConfigMap);

const {
  startDate,
  endDate,
  rangeExporting,
  downloadExcel,
} = useDashboardExcelExport({
  selectedDate: localDate,
  locationFilter,
  machineSettings,
  showNotice,
  productivityShift,
  shiftConfigMap,
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

  const date = startDate.value || localDate.value;
  const shift = productivityShift.value;

  await loadDashboard(date, { shift });
  connect(date, shift);
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

    await loadDashboard(localDate.value, { shift: productivityShift.value });
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

    await loadDashboard(localDate.value, { shift: productivityShift.value });
  } catch (err) {
    showNotice(`Gagal hapus: ${err.message}`, "error");
  }
}

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());

  emit("admin-mode-change", isAdmin.value);
  emit("socket-status-change", socketStatus.value);

  await refreshDashboardByDate();
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
      v-model:location-value="locationFilter"
      v-model:shift-value="shiftFilter"
      :location-options="locationOptions"
      :shift-options="shiftOptions"
      :show-shift-filter="showShiftFilter"
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