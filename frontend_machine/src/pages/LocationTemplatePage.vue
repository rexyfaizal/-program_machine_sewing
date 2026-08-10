<script setup>
import LocationToolbar from "../components/location/LocationToolbar.vue";
import LocationInfoCards from "../components/location/LocationInfoCards.vue";
import LocationLineGrid from "../components/location/LocationLineGrid.vue";
import LineModal from "../components/location/LineModal.vue";
import MachineLocationModal from "../components/location/MachineLocationModal.vue";
import ShiftConfigModal from "../components/location/ShiftConfigModal.vue";

import { useLocationTemplate } from "../composables/useLocationTemplate";

const {
  selectedFactory,
  selectedDate,
  isAdmin,
  loading,
  errorMessage,
  notice,
  machines,
  modalOpen,
  modalMode,
  lineModalOpen,
  lineModalMode,
  oldLineName,
  lineFormName,
  shiftModalOpen,
  shiftModalSaving,
  shiftConfigs,
  shiftSettings,
  saturdayShiftSettings,
  shiftDefaults,
  saturdayShiftDefaults,
  draggingLine,
  dragOverLine,
  form,
  factoryOptions,
  lineLayout,
  activeLines,
  assignedCount,
  makeLocation,
  loadData,
  machinesByLine,
  selectFactory,
  openAddLineModal,
  openRenameLineModal,
  closeLineModal,
  saveLine,
  deleteLine,
  openShiftConfigModal,
  closeShiftConfigModal,
  saveShiftConfig,
  onLineDragStart,
  onLineDragEnter,
  onLineDragOver,
  onLineDrop,
  onLineDragEnd,
  openAddModal,
  openEditModal,
  closeModal,
  fillNameFromSelectedMachine,
  saveLocation,
  removeMachineFromLine,
} = useLocationTemplate();

const emit = defineEmits(["openDetailMachine"]);

function openDetailMachine(machine) {
  emit("openDetailMachine", machine);
}
</script>

<template>
  <section class="location-page">
    <LocationToolbar
      v-model:selected-factory="selectedFactory"
      v-model:selected-date="selectedDate"
      :factory-options="factoryOptions"
      :is-admin="isAdmin"
      @factory-change="selectFactory"
      @refresh="loadData"
      @open-shift-config="openShiftConfigModal"
    />

    <div v-if="notice" class="notice">
      {{ notice }}
    </div>

    <div v-if="errorMessage" class="notice error">
      {{ errorMessage }}
    </div>

    <LocationInfoCards
      :selected-factory="selectedFactory"
      :line-count="activeLines.length"
      :assigned-count="assignedCount"
    />

    <div v-if="loading" class="notice">
      Mengambil data mesin...
    </div>

    <LocationLineGrid
      :lines="activeLines"
      :is-admin="isAdmin"
      :dragging-line="draggingLine"
      :drag-over-line="dragOverLine"
      :machines-by-line="machinesByLine"
      @add-line="openAddLineModal"
      @rename-line="openRenameLineModal"
      @delete-line="deleteLine"
      @add-machine="openAddModal"
      @edit-machine="openEditModal"
      @remove-machine="removeMachineFromLine"
      @drag-start="onLineDragStart"
      @drag-enter="onLineDragEnter"
      @drag-over="onLineDragOver"
      @drop-line="onLineDrop"
      @drag-end="onLineDragEnd"
      @open-detail-machine="openDetailMachine"
    />

    <LineModal
      :show="lineModalOpen"
      :mode="lineModalMode"
      :selected-factory="selectedFactory"
      :old-line-name="oldLineName"
      v-model:line-name="lineFormName"
      @close="closeLineModal"
      @save="saveLine"
    />

    <ShiftConfigModal
      :show="shiftModalOpen"
      :factory="selectedFactory"
      :lines="activeLines"
      :saved-configs="shiftConfigs"
      :saved-shifts="shiftSettings"
      :saved-saturday-shifts="saturdayShiftSettings"
      :defaults="shiftDefaults"
      :saturday-defaults="saturdayShiftDefaults"
      :saving="shiftModalSaving"
      @close="closeShiftConfigModal"
      @save="saveShiftConfig"
    />

    <MachineLocationModal
      :show="modalOpen"
      :mode="modalMode"
      :form="form"
      :machines="machines"
      :factory-options="factoryOptions"
      :line-layout="lineLayout"
      :make-location="makeLocation"
      @close="closeModal"
      @save="saveLocation"
      @select-machine="fillNameFromSelectedMachine"
    />
  </section>
</template>

<style scoped>
.location-page {
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 24px;
  padding: 24px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
}

.notice {
  background: #f8fafc;
  border: 1px solid #dbe4ef;
  border-radius: 14px;
  padding: 12px 14px;
  color: #0f172a;
  font-weight: 800;
  margin-bottom: 16px;
}

.notice.error {
  background: #fef2f2;
  border-color: #fecaca;
  color: #b91c1c;
}
</style>
