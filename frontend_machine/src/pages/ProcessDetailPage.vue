<script setup>
import { computed, onMounted, ref } from "vue";

import MachineSidebar from "../components/process/MachineSidebar.vue";
import ProcessTopbar from "../components/process/ProcessTopbar.vue";
import ProcessKpi from "../components/process/ProcessKpi.vue";
import HourChart from "../components/process/HourChart.vue";
import ProgramBars from "../components/process/ProgramBars.vue";
import ProcessOutputChart from "../components/process/ProcessOutputChart.vue";
import AlarmSummary from "../components/process/AlarmSummary.vue";
import ProcessTable from "../components/process/ProcessTable.vue";
import MachineEditModal from "../components/dashboard/MachineEditModal.vue";

import { useDashboard } from "../composables/useDashboard";
import { useProcessDetail } from "../composables/useProcessDetail";
import { deleteMachineSetting, saveMachineSetting } from "../api/machineApi";
import { getInitialAdminMode } from "../utils/adminMode";

const props = defineProps({
  selectedDate: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update:selectedDate"]);

const processKeyword = ref("");
const selectedUUID = ref("");
const editModalOpen = ref(false);
const selectedEditMachine = ref(null);
const notice = ref("");
const noticeType = ref("ok");
const isAdmin = ref(getInitialAdminMode());
const detailTab = ref("output"); // output | proses

const {
  machines,
  loading,
  errorMessage,
  loadDashboard,
} = useDashboard();

const {
  detailLoading,
  detailError,
  detailData,
  detailMachine,
  detailEvents,
  detailGroups,
  detailHours,
  detailAlarms,
  pagedDetailEvents,
  detailPage,
  detailPageSize,
  totalDetailPages,
  maxHourProc,
  maxGroupProc,
  avgNodeDistance,
  loadProcessDetail,
  prevDetailPage,
  nextDetailPage,
  goDetailPage,
  visibleDetailPages,
} = useProcessDetail();

const localDate = computed({
  get: () => props.selectedDate,
  set: (value) => emit("update:selectedDate", value),
});

const machineListForDetail = computed(() => {
  return [...machines.value].sort((a, b) => Number(b.output || 0) - Number(a.output || 0));
});

const filteredProcessMachines = computed(() => {
  const key = processKeyword.value.toLowerCase().trim();

  if (!key) return machineListForDetail.value;

  return machineListForDetail.value.filter((m) => {
    return [
      m.machineName,
      m.originalMachineName,
      m.location,
      m.ip,
      m.uuid,
      m.status,
      m.program,
    ]
      .join(" ")
      .toLowerCase()
      .includes(key);
  });
});

const selectedMachine = computed(() => {
  return machineListForDetail.value.find((m) => sameText(m.uuid, selectedUUID.value)) || null;
});

const displayMachine = computed(() => {
  const selected = selectedMachine.value || {};
  const detail = detailMachine.value || {};

  const manualName =
    selected.machineName ||
    selected.customName ||
    selected.nickName ||
    detail.machineName ||
    detail.nickName ||
    detail.originalNickName ||
    detail.uuid ||
    "Pilih Mesin";

  return {
    ...detail,
    ...selected,
    machineName: manualName,
    nickName: manualName,
    displayMachineName:
      selected.displayMachineName ||
      selected.operatorProcessName ||
      manualName,
    operatorProcessName: selected.operatorProcessName || "",
    operatorSessions: selected.operatorSessions || [],
    location: selected.location || detail.location || "-",
    ip: selected.ip || detail.ip || "-",
    uuid: selected.uuid || detail.uuid || "-",
  };
});

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function sameText(a, b) {
  return normalizeText(a) === normalizeText(b);
}

function getPendingDetailUuid() {
  return localStorage.getItem("machineDashboardDetailUuid") || "";
}

function clearPendingDetailUuid() {
  localStorage.removeItem("machineDashboardDetailUuid");
  localStorage.removeItem("machineDashboardDetailMachineName");
}

async function applyPendingMachineFromLocation() {
  const pendingUuid = getPendingDetailUuid();

  if (!pendingUuid) return false;

  const targetMachine = machineListForDetail.value.find((machine) => {
    return sameText(machine.uuid, pendingUuid);
  });

  if (!targetMachine) {
    return false;
  }

  selectedUUID.value = targetMachine.uuid;
  processKeyword.value = "";

  clearPendingDetailUuid();

  await loadProcessDetail(localDate.value, targetMachine.uuid);

  return true;
}

function showNotice(message, type = "ok") {
  notice.value = message;
  noticeType.value = type;

  setTimeout(() => {
    notice.value = "";
  }, 3000);
}

async function initializePage() {
  await loadDashboard(localDate.value);

  const appliedFromLocation = await applyPendingMachineFromLocation();
  if (appliedFromLocation) return;

  if (!selectedUUID.value && machineListForDetail.value.length) {
    const firstActive =
      machineListForDetail.value.find((m) => Number(m.output || 0) > 0) ||
      machineListForDetail.value[0];

    selectedUUID.value = firstActive.uuid;
  }

  if (selectedUUID.value) {
    await loadProcessDetail(localDate.value, selectedUUID.value);
  }
}

async function refreshPage() {
  const keepUUID = selectedUUID.value;

  await loadDashboard(localDate.value);

  const appliedFromLocation = await applyPendingMachineFromLocation();
  if (appliedFromLocation) return;

  if (keepUUID) {
    const stillExist = machineListForDetail.value.find((m) => sameText(m.uuid, keepUUID));

    if (stillExist) {
      selectedUUID.value = stillExist.uuid;
    }
  }

  if (!selectedUUID.value && machineListForDetail.value.length) {
    selectedUUID.value = machineListForDetail.value[0].uuid;
  }

  if (selectedUUID.value) {
    await loadProcessDetail(localDate.value, selectedUUID.value);
  }
}

async function selectProcessMachine(uuid) {
  selectedUUID.value = uuid;
  clearPendingDetailUuid();
  await loadProcessDetail(localDate.value, uuid);
}

async function refreshDetailByDate() {
  await refreshPage();
}

function moveDetailDay(delta) {
  const d = new Date(localDate.value + "T00:00:00");
  d.setDate(d.getDate() + delta);
  localDate.value = d.toISOString().slice(0, 10);
  refreshDetailByDate();
}

function setDetailToday() {
  const d = new Date();
  const offset = d.getTimezoneOffset();
  const local = new Date(d.getTime() - offset * 60000);
  localDate.value = local.toISOString().slice(0, 10);
  refreshDetailByDate();
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

async function handleSaveMachineSetting(payload) {
  if (!isAdmin.value) {
    showNotice("Akses simpan hanya untuk admin.", "error");
    return;
  }

  try {
    await saveMachineSetting(payload);
    showNotice("Nama/location mesin berhasil disimpan.", "ok");
    closeEditModal();

    await refreshPage();
  } catch (err) {
    showNotice(`Gagal simpan: ${err.message}`, "error");
  }
}

async function handleDeleteMachineSetting(uuid) {
  if (!isAdmin.value) {
    showNotice("Akses hapus hanya untuk admin.", "error");
    return;
  }

  try {
    await deleteMachineSetting(uuid);
    showNotice("Setting manual berhasil dihapus.", "ok");
    closeEditModal();

    await refreshPage();
  } catch (err) {
    showNotice(`Gagal hapus: ${err.message}`, "error");
  }
}

onMounted(() => {
  initializePage();
});
</script>

<template>
  <div class="process-page">
    <MachineSidebar
      v-model:selected-date="localDate"
      v-model:keyword="processKeyword"
      :machines="filteredProcessMachines"
      :selected-u-u-i-d="selectedUUID"
      :is-admin="isAdmin"
      @select="selectProcessMachine"
      @refresh="refreshPage"
      @date-change="refreshDetailByDate"
      @edit="openEditModal"
    />

    <main class="process-main">
      <ProcessTopbar
        v-model:selected-date="localDate"
        :machine="displayMachine"
        @last-day="moveDetailDay(-1)"
        @next-day="moveDetailDay(1)"
        @today="setDetailToday"
        @date-change="refreshDetailByDate"
      />


      <div v-if="loading" class="alert">Mengambil daftar mesin...</div>
      <div v-if="errorMessage" class="alert error">{{ errorMessage }}</div>
      <div v-if="detailLoading" class="alert">Mengambil detail proses...</div>
      <div v-if="detailError" class="alert error">{{ detailError }}</div>
      <div v-if="notice" class="alert" :class="{ error: noticeType === 'error' }">
        {{ notice }}
      </div>

      <div class="detail-tabs" role="tablist" aria-label="Tampilan detail mesin">
        <button
          type="button"
          class="detail-tab"
          :class="{ active: detailTab === 'output' }"
          role="tab"
          :aria-selected="detailTab === 'output'"
          @click="detailTab = 'output'"
        >
          Detail Mesin
        </button>
        <button
          type="button"
          class="detail-tab"
          :class="{ active: detailTab === 'proses' }"
          role="tab"
          :aria-selected="detailTab === 'proses'"
          @click="detailTab = 'proses'"
        >
          Detail Output
        </button>
      </div>

      <template v-if="detailTab === 'output'">
        <ProcessKpi
          :machine="displayMachine"
          :detail-machine="detailMachine"
          :selected-machine="selectedMachine"
          :events="detailEvents"
          :avg-node-distance="avgNodeDistance"
        />

        <section class="process-grid">
          <HourChart
            :hours="detailHours"
            :selected-date="localDate"
            :max-hour-proc="maxHourProc"
          />

          <ProgramBars
            :groups="detailGroups"
            :max-group-proc="maxGroupProc"
          />
        </section>

        <ProcessOutputChart
          :events="detailEvents"
          :groups="detailGroups"
        />

        <section class="process-grid single-panel">
          <AlarmSummary :alarms="detailAlarms" />
        </section>
      </template>

      <ProcessTable
        v-else
        :events="detailEvents"
        :paged-events="pagedDetailEvents"
        :page="detailPage"
        :total-pages="totalDetailPages"
        :page-size="detailPageSize"
        :visible-pages="visibleDetailPages"
        :machine-name="displayMachine.machineName || displayMachine.nickName || ''"
        :uuid="selectedUUID"
        :location="displayMachine.location || ''"
        :selected-date="localDate"
        @prev="prevDetailPage"
        @next="nextDetailPage"
        @go="goDetailPage"
        @notice="showNotice"
      />
    </main>

    <MachineEditModal
      v-if="isAdmin"
      :show="editModalOpen"
      :machine="selectedEditMachine"
      @close="closeEditModal"
      @save="handleSaveMachineSetting"
      @delete="handleDeleteMachineSetting"
    />
  </div>
</template>

<style scoped>
.process-page {
  display: grid;
  grid-template-columns: 380px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.process-main {
  min-width: 0;
}

.detail-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.detail-tab {
  border: 1px solid #d8e2ee;
  background: #fff;
  color: #334155;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 800;
  cursor: pointer;
}

.detail-tab:hover {
  background: #f8fafc;
}

.detail-tab.active {
  background: #2563eb;
  border-color: #2563eb;
  color: #fff;
}

.process-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.single-panel {
  grid-template-columns: minmax(320px, 520px);
}

.alert {
  background: white;
  border: 1px solid #d8e2ee;
  padding: 14px 16px;
  border-radius: 14px;
  margin-bottom: 16px;
  font-weight: 800;
}

.alert.error {
  color: #b00020;
}


@media (max-width: 1200px) {
  .process-page {
    grid-template-columns: 1fr;
  }

  .process-grid {
    grid-template-columns: 1fr;
  }

  .single-panel {
    grid-template-columns: 1fr;
  }
}
</style>