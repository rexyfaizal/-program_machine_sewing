<script setup>
import {
  getMachineStatusText,
  getMachineStatusClass,
} from "../../utils/machineStatus";

defineProps({
  line: {
    type: String,
    required: true,
  },
  machines: {
    type: Array,
    default: () => [],
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
  draggingLine: {
    type: String,
    default: "",
  },
  dragOverLine: {
    type: String,
    default: "",
  },
});

const emit = defineEmits([
  "renameLine",
  "deleteLine",
  "addMachine",
  "editMachine",
  "removeMachine",
  "dragStart",
  "dragEnter",
  "dragOver",
  "dropLine",
  "dragEnd",
  "openDetailMachine",
]);

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function formatHour(seconds) {
  const hour = toNumber(seconds) / 3600;
  return `${hour.toFixed(2)}h`;
}

function formatSecond(seconds) {
  return `${toNumber(seconds).toFixed(0)}s`;
}

function downTimeSec(machine) {
  const powerOnDuration = toNumber(machine?.runtime);
  const runningTime = toNumber(machine?.procTime);

  return Math.max(0, powerOnDuration - runningTime);
}

function productivityText(machine) {
  return `${toNumber(machine?.productivity).toFixed(2)}%`;
}

function operatorLoginTitle(machine) {
  const name = String(machine?.operatorName || "").trim();
  const nik = String(machine?.operatorNik || "").trim();

  if (nik && name) return `Operator login: ${nik} - ${name}`;
  if (name) return `Operator login: ${name}`;
  if (nik) return `Operator login: ${nik}`;

  return "Operator sedang login";
}
</script>

<template>
  <div
    class="line-card"
    :class="{
      dragging: draggingLine === line,
      'drag-over': dragOverLine === line,
      'admin-draggable': isAdmin
    }"
    :draggable="isAdmin"
    @dragstart="emit('dragStart', { event: $event, line })"
    @dragenter.prevent="emit('dragEnter', line)"
    @dragover="emit('dragOver', $event)"
    @drop.prevent="emit('dropLine', line)"
    @dragend="emit('dragEnd')"
  >
    <div class="line-card-head">
      <div class="line-title-row">
        <span>{{ line }}</span>
        <small v-if="isAdmin">Drag untuk pindah posisi</small>
      </div>

      <div v-if="isAdmin" class="line-actions">
        <button
          type="button"
          class="line-action-btn"
          @click.stop="emit('renameLine', line)"
        >
          Rename
        </button>

        <button
          type="button"
          class="line-action-btn danger"
          @click.stop="emit('deleteLine', line)"
        >
          Hapus
        </button>
      </div>

      <button
        v-if="isAdmin"
        type="button"
        class="add-btn"
        @click.stop="emit('addMachine', line)"
      >
        + Mesin
      </button>
    </div>

    <div class="line-card-body">
      <div class="line-placeholder">
        <div
          v-for="machine in machines"
          :key="machine.uuid"
          class="machine-chip clickable"
          @click.stop="emit('openDetailMachine', machine)"
        >
          <div class="machine-status-col">
            <span
              class="machine-dot"
              :class="getMachineStatusClass(machine)"
              :title="getMachineStatusText(machine)"
            ></span>

            <span
              v-if="machine.operatorLoggedIn"
              class="operator-check"
              :title="operatorLoginTitle(machine)"
            >
              ✓
            </span>
          </div>

          <div class="machine-info">
            <strong :title="machine.machineName">
              {{ machine.machineName }}
            </strong>

            <small
              v-if="machine.usingProcessName && machine.operatorStyleName"
              class="style-line"
            >
              Style: {{ machine.operatorStyleName }}
            </small>

            <small>{{ getMachineStatusText(machine) }}</small>

            <small v-if="isAdmin">
              IP: {{ machine.ip }}
            </small>

            <small v-if="isAdmin">
              Output: {{ machine.output }}
            </small>
          </div>

          <div class="machine-tooltip">
            <div class="tooltip-title">
              {{ machine.machineName }}
            </div>

            <div
              v-if="machine.operatorLoggedIn"
              class="tooltip-row"
            >
              <span>Operator</span>
              <strong>
                {{
                  machine.operatorNik && machine.operatorName
                    ? `${machine.operatorNik} - ${machine.operatorName}`
                    : machine.operatorName || machine.operatorNik || "Login"
                }}
              </strong>
            </div>

            <div
              v-if="machine.usingProcessName && machine.operatorStyleName"
              class="tooltip-row"
            >
              <span>Style</span>
              <strong>{{ machine.operatorStyleName }}</strong>
            </div>

            <div
              v-if="machine.usingProcessName && machine.customName"
              class="tooltip-row"
            >
              <span>Mesin</span>
              <strong>{{ machine.customName }}</strong>
            </div>

            <div class="tooltip-row">
              <span>Power On Duration</span>
              <strong>{{ formatHour(machine.runtime) }}</strong>
              <small>{{ formatSecond(machine.runtime) }}</small>
            </div>

            <div class="tooltip-row">
              <span>Running Time</span>
              <strong>{{ formatHour(machine.procTime) }}</strong>
              <small>{{ formatSecond(machine.procTime) }}</small>
            </div>

            <div class="tooltip-row">
              <span>Loss Time</span>
              <strong>{{ formatHour(downTimeSec(machine)) }}</strong>
              <small>{{ formatSecond(downTimeSec(machine)) }}</small>
            </div>

            <div class="tooltip-row">
              <span>Produktivitas</span>
              <strong>{{ productivityText(machine) }}</strong>
              <small>{{ machine.mainSource || "process_time" }}</small>
            </div>
          </div>

          <div v-if="isAdmin" class="chip-actions">
            <button type="button" @click.stop="emit('editMachine', machine)">
              Edit
            </button>

            <button
              type="button"
              class="remove"
              @click.stop="emit('removeMachine', machine)"
            >
              Hapus
            </button>
          </div>
        </div>

        <span v-if="!machines.length" class="empty-line">
          TEMPLATE AREA
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.line-card {
  background: #ffffff;
  border: 2px solid #2563eb;
  border-radius: 11px;
  box-shadow:
    0 0 0 1px rgba(37, 99, 235, 0.08),
    0 10px 22px rgba(37, 99, 235, 0.10);
  border-radius: 12px;
  min-height: 520px;
  display: flex;
  flex-direction: column;
  overflow: visible;
  cursor: default;
  position: relative;
  z-index: 1;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease,
    background 0.2s ease;
}

.line-card.admin-draggable {
  cursor: grab;
}

.line-card.admin-draggable:active {
  cursor: grabbing;
}

.line-card:hover {
  background: linear-gradient(180deg, #2563eb 0%, #1d4ed8 100%);
  border-color: #1d4ed8;
  transform: translateY(-3px);
  box-shadow: 0 16px 34px rgba(37, 99, 235, 0.28);
  z-index: 9999;
}

.line-card.dragging {
  opacity: 0.45;
  transform: scale(0.98);
  border-color: #2563eb;
}

.line-card.drag-over {
  border-color: #2563eb;
  box-shadow:
    0 0 0 3px rgba(37, 99, 235, 0.18),
    0 16px 32px rgba(37, 99, 235, 0.16);
}

.line-card-head {
  padding: 9px 5px;
  border-bottom: 1px solid #1d4ed8;
  background: #2563eb;
  text-align: center;
  display: grid;
  gap: 6px;
  border-radius: 9px 9px 0 0;
}

.line-card:hover .line-card-head {
  background: #1d4ed8;
  border-bottom-color: rgba(255, 255, 255, 0.28);
}

.line-title-row {
  display: grid;
  gap: 3px;
}

.line-title-row span {
  font-size: 11px;
  font-weight: 900;
  color: #ffffff;
  letter-spacing: 0.04em;
}

.line-title-row small {
  font-size: 9px;
  color: rgba(255, 255, 255, 0.85);
  font-weight: 800;
}

.line-card:hover .line-title-row span,
.line-card:hover .line-title-row small {
  color: #ffffff;
}

.line-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 5px;
}

.line-action-btn,
.add-btn {
  border: 1px solid #dbe4ef;
  background: white;
  color: #0f172a;
  border-radius: 8px;
  padding: 6px 6px;
  font-size: 10px;
  font-weight: 900;
  cursor: pointer;
}

.line-action-btn:hover {
  background: #e0ecff;
  color: #1d4ed8;
  border-color: #93c5fd;
}

.add-btn {
  background: #2563eb;
  color: #ffffff;
  border-color: #2563eb;
}

.add-btn:hover {
  background: #1d4ed8;
  color: #ffffff;
  border-color: #1d4ed8;
}

.line-card:hover .add-btn {
  background: #ffffff;
  color: #1d4ed8;
  border-color: #ffffff;
}

.line-card:hover .line-action-btn {
  background: rgba(255, 255, 255, 0.95);
  border-color: rgba(255, 255, 255, 0.85);
  color: #0f172a;
}

.line-action-btn.danger {
  color: #b91c1c;
}

.line-action-btn.danger:hover {
  background: #b91c1c;
  color: white;
  border-color: #b91c1c;
}

.line-card:hover .line-action-btn.danger {
  color: #b91c1c;
}

.line-card-body {
  flex: 1;
  padding: 10px;
  display: flex;
}

.line-placeholder {
  width: 100%;
  min-height: 100%;
  border: 1.5px dashed rgba(37, 99, 235, 0.35);
  border-radius: 10px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background:
    linear-gradient(to right, rgba(148, 163, 184, 0.06) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(148, 163, 184, 0.06) 1px, transparent 1px);
  background-size: 14px 14px;
  overflow: visible;
}

.line-card:hover .line-placeholder {
  border-color: rgba(255, 255, 255, 0.5);
  background:
    linear-gradient(to right, rgba(255, 255, 255, 0.08) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255, 255, 255, 0.08) 1px, transparent 1px);
  background-size: 14px 14px;
}

.empty-line {
  margin: auto;
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  font-size: 11px;
  font-weight: 800;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.line-card:hover .empty-line {
  color: rgba(255, 255, 255, 0.75);
}

.machine-chip {
  background: linear-gradient(135deg, #ffffff 0%, #eff6ff 100%);
  border: 1.5px solid #93c5fd;
  border-left: 4px solid #60a5fa;
  border-radius: 12px;
  padding: 8px;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: 7px;
  box-shadow:
    0 8px 16px rgba(15, 23, 42, 0.08),
    inset 0 0 0 1px rgba(255, 255, 255, 0.6);
  position: relative;
  overflow: visible;
  z-index: 1;
}

.machine-chip.clickable {
  cursor: pointer;
}

.machine-chip.clickable:hover {
  transform: translateY(-3px) scale(1.04);
  background: linear-gradient(135deg, #fff7ed 0%, #fef3c7 100%);
  border-color: #facc15;
  border-left-color: #f97316;
  box-shadow:
    0 16px 32px rgba(15, 23, 42, 0.24),
    0 0 0 3px rgba(250, 204, 21, 0.35);
  z-index: 10000;
}

.machine-chip.clickable:hover .machine-info strong {
  color: #7c2d12;
}

.machine-chip.clickable:hover .machine-info small {
  color: #92400e;
  font-weight: 800;
}

.machine-chip.clickable:hover .machine-dot {
  box-shadow:
    0 0 0 3px rgba(255, 255, 255, 0.95),
    0 0 0 6px rgba(250, 204, 21, 0.45);
}

.machine-chip:hover {
  border-color: #2563eb;
}

.line-card:hover .machine-chip {
  background: linear-gradient(135deg, #ffffff 0%, #e0f2fe 100%);
  border-color: #bfdbfe;
  border-left-color: #ffffff;
}

.machine-status-col {
  display: grid;
  justify-items: center;
  align-content: start;
  gap: 5px;
  margin-top: 2px;
}

.machine-dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
}

.operator-check {
  width: 14px;
  height: 14px;
  border-radius: 999px;
  display: inline-grid;
  place-items: center;
  background: #16a34a;
  color: #ffffff;
  font-size: 9px;
  font-weight: 900;
  line-height: 1;
  box-shadow: 0 0 0 3px rgba(22, 163, 74, 0.18);
}

/* MacState 2 = Working = Hijau */
.machine-dot.working {
  background: #22c55e;
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.16);
}

/* MacState 1 = Online = Biru */
.machine-dot.online {
  background: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.16);
}

/* MacState 0 = Offline = Abu-abu */
.machine-dot.offline {
  background: #94a3b8;
  box-shadow: 0 0 0 4px rgba(148, 163, 184, 0.18);
}

.machine-info {
  min-width: 0;
}

.machine-info strong {
  display: block;
  font-size: 11px;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.machine-info small {
  display: block;
  font-size: 10px;
  color: #64748b;
  margin-top: 2px;
}

.machine-info small.style-line {
  color: #2563eb;
  font-weight: 700;
}

.machine-tooltip {
  position: absolute;
  left: calc(100% + 12px);
  top: 50%;
  transform: translateY(-50%) translateX(-6px);
  width: 245px;
  background: #0f172a;
  color: #ffffff;
  border-radius: 14px;
  padding: 12px;
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.35);
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition:
    opacity 0.18s ease,
    transform 0.18s ease,
    visibility 0.18s ease;
  z-index: 10001;
}

.machine-tooltip::before {
  content: "";
  position: absolute;
  left: -7px;
  top: 50%;
  transform: translateY(-50%);
  border-top: 7px solid transparent;
  border-bottom: 7px solid transparent;
  border-right: 7px solid #0f172a;
}

.machine-chip:hover .machine-tooltip {
  opacity: 1;
  visibility: visible;
  transform: translateY(-50%) translateX(0);
}

.tooltip-title {
  font-size: 12px;
  font-weight: 900;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.16);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tooltip-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 4px 8px;
  align-items: center;
  margin-top: 8px;
}

.tooltip-row span {
  color: #cbd5e1;
  font-size: 11px;
  font-weight: 800;
}

.tooltip-row strong {
  color: #ffffff;
  font-size: 12px;
  font-weight: 900;
}

.tooltip-row small {
  grid-column: 1 / -1;
  color: #94a3b8;
  font-size: 10px;
  margin: 0;
}

.chip-actions {
  grid-column: 1 / -1;
  display: flex;
  gap: 5px;
  margin-top: 4px;
}

.chip-actions button {
  flex: 1;
  border: 1px solid #dbe4ef;
  background: #ffffff;
  color: #0f172a;
  border-radius: 8px;
  padding: 5px;
  font-size: 10px;
  font-weight: 900;
  cursor: pointer;
}

.chip-actions button:hover {
  background: #f1f5f9;
}

.chip-actions .remove {
  color: #b91c1c;
}
</style>