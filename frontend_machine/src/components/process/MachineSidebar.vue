<script setup>
import {
  getMachineStatusText,
  getMachineStatusClass,
} from "../../utils/machineStatus";

const props = defineProps({
  machines: {
    type: Array,
    required: true,
    default: () => [],
  },
  selectedUuid: {
    type: String,
    default: "",
  },
  selectedDate: {
    type: String,
    required: true,
  },
  keyword: {
    type: String,
    default: "",
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits([
  "select",
  "refresh",
  "update:selectedDate",
  "update:keyword",
  "dateChange",
  "edit",
]);

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function isSelectedMachine(machine) {
  return normalizeText(props.selectedUuid) === normalizeText(machine?.uuid);
}

function getMacStateValue(machine) {
  return (
    machine?.macState ??
    machine?.MacState ??
    machine?.mac_state ??
    machine?.Macstate ??
    machine?.macstate ??
    machine?.MACSTATE ??
    "0"
  );
}

function updateDate(event) {
  emit("update:selectedDate", event.target.value);
}

function updateKeyword(event) {
  emit("update:keyword", event.target.value);
}

function selectMachine(uuid) {
  emit("select", uuid);
}

function editMachine(event, machine) {
  event.stopPropagation();
  emit("edit", machine);
}
</script>

<template>
  <aside class="process-side">
    <div class="process-brand">
      <h2>Detail Proses</h2>

      <button type="button" @click="emit('refresh')">
        Refresh
      </button>
    </div>

    <div class="process-filter">
      <input
        type="date"
        :value="props.selectedDate"
        @input="updateDate"
        @change="emit('dateChange')"
      />

      <input
        :value="props.keyword"
        placeholder="Cari mesin / IP / UUID..."
        @input="updateKeyword"
      />
    </div>

    <div class="group-caption">
      <span>Daftar Mesin</span>
      <strong>{{ machines.length }}/{{ machines.length }}</strong>
    </div>

    <div class="machine-list">
      <button
        v-for="m in machines"
        :key="m.uuid"
        class="machine-item"
        :class="{ active: isSelectedMachine(m) }"
        type="button"
        @click="selectMachine(m.uuid)"
      >
        <span
          class="machine-dot"
          :class="getMachineStatusClass(m)"
        ></span>

        <span class="machine-text">
          <strong :title="m.machineName">
            {{ m.machineName || m.nickName || m.uuid || "-" }}
          </strong>

          <div class="machine-meta-box">
            <span class="meta-pill status-pill" :class="getMachineStatusClass(m)">
              {{ getMachineStatusText(m) }}
            </span>

            <span class="meta-pill">
              IP: {{ m.ip || "-" }}
            </span>

            <span
              v-if="m.location && m.location !== '-'"
              class="meta-pill location-pill"
            >
              Location: {{ m.location }}
            </span>

            <span v-if="isAdmin" class="meta-pill uuid-pill">
              UUID: {{ m.uuid || "-" }}
            </span>

            <span v-if="isAdmin" class="meta-pill macstate-pill">
              MacState: {{ getMacStateValue(m) }}
            </span>
          </div>
        </span>

        <span class="machine-actions">
          <button
            v-if="isAdmin"
            class="edit-mini"
            type="button"
            title="Edit nama/location"
            @click="editMachine($event, m)"
          >
            Edit
          </button>

          <span v-if="isSelectedMachine(m)" class="active-label">
            Aktif
          </span>
        </span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.process-side {
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 20px;
  padding: 16px;
  max-height: calc(100vh - 110px);
  overflow: auto;
  position: sticky;
  top: 18px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
}

.process-brand {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.process-brand h2 {
  margin: 0;
  font-size: 18px;
}

.process-brand button {
  border: 0;
  background: #2563eb;
  color: white;
  border-radius: 12px;
  padding: 10px 12px;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.process-brand button:hover {
  background: #1d4ed8;
}

.process-filter {
  display: grid;
  gap: 10px;
  margin-bottom: 14px;
}

.process-filter input {
  height: 42px;
  border: 1px solid #d8e2ee;
  border-radius: 13px;
  padding: 0 12px;
  outline: none;
}

.group-caption {
  display: flex;
  justify-content: space-between;
  color: #64748b;
  font-size: 12px;
  font-weight: 900;
  margin: 12px 4px 8px;
}

.machine-list {
  display: grid;
  gap: 8px;
}

.machine-item {
  width: 100%;
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 72px;
  gap: 10px;
  align-items: center;
  border: 1px solid #dbe4ef;
  background: white;
  border-radius: 15px;
  padding: 11px;
  cursor: pointer;
  text-align: left;
  transition:
    background 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.machine-item.active {
  background: #e0f2fe;
  border-color: #0f172a;
  box-shadow: inset 0 0 0 1px #0f172a;
}

.machine-dot {
  width: 28px;
  height: 28px;
  border-radius: 999px;
  background: #94a3b8;
}

/* MacState 2 = Working = Hijau */
.machine-dot.working {
  background: linear-gradient(135deg, #86efac, #16a34a);
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.14);
}

/* MacState 1 = Online = Biru */
.machine-dot.online {
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.14);
}

/* MacState 0 = Offline = Abu-abu */
.machine-dot.offline {
  background: linear-gradient(135deg, #cbd5e1, #64748b);
  box-shadow: 0 0 0 4px rgba(148, 163, 184, 0.16);
}

.machine-text {
  min-width: 0;
}

.machine-text strong {
  display: block;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.machine-text small {
  display: block;
  color: #64748b;
  margin-top: 3px;
  font-size: 12px;
}

.machine-meta-box {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-top: 6px;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  max-width: 100%;
  border-radius: 999px;
  padding: 4px 7px;
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.2;
}

.location-pill {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

.uuid-pill {
  background: #f8fafc;
  color: #334155;
  border-color: #cbd5e1;
  font-family: Consolas, Menlo, monospace;
}

.macstate-pill {
  background: #f5f3ff;
  color: #6d28d9;
  border-color: #ddd6fe;
}

/* Working = Hijau */
.status-pill.working {
  background: #dcfce7;
  color: #166534;
  border-color: #bbf7d0;
}

/* Online = Biru */
.status-pill.online {
  background: #dbeafe;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

/* Offline = Abu-abu */
.status-pill.offline {
  background: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}

.location-text {
  color: #334155 !important;
  font-weight: 800;
}

.macstate-text {
  color: #475569 !important;
  font-weight: 800;
}

.uuid-text {
  color: #475569 !important;
  font-weight: 800;
  font-family: Consolas, Menlo, monospace;
  font-size: 10px !important;
}

.machine-actions {
  display: grid;
  justify-items: end;
  gap: 5px;
}

.edit-mini {
  border: 1px solid #dbe4ef;
  background: white;
  color: #0f172a;
  border-radius: 9px;
  font-size: 11px;
  font-weight: 900;
  cursor: pointer;
  padding: 6px 8px;
}

.edit-mini:hover {
  background: #f1f5f9;
}

.active-label {
  color: #0f766e;
  background: #ccfbf1;
  border-radius: 999px;
  padding: 4px 7px;
  font-size: 10px;
  font-weight: 900;
}

.machine-item:hover {
  background: linear-gradient(135deg, #0878ec 0%, #0874df 60%, #0f63c7 100%);
  border-color: #0878ec;
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 14px 28px rgba(8, 120, 236, 0.28);
}

.machine-item:hover .machine-text strong {
  color: #ffffff !important;
}

.machine-item:hover .meta-pill {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff !important;
  border-color: rgba(255, 255, 255, 0.28);
}

/* Hover tetap mempertahankan warna status mesin */
.machine-item:hover .machine-dot.working {
  background: linear-gradient(135deg, #86efac, #16a34a);
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.22);
}

.machine-item:hover .machine-dot.online {
  background: linear-gradient(135deg, #60a5fa, #2563eb);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.22);
}

.machine-item:hover .machine-dot.offline {
  background: linear-gradient(135deg, #cbd5e1, #64748b);
  box-shadow: 0 0 0 4px rgba(148, 163, 184, 0.24);
}

.machine-item:hover .status-pill.working {
  background: #dcfce7;
  color: #166534 !important;
  border-color: #bbf7d0;
}

.machine-item:hover .status-pill.online {
  background: #dbeafe;
  color: #1d4ed8 !important;
  border-color: #bfdbfe;
}

.machine-item:hover .status-pill.offline {
  background: #f1f5f9;
  color: #475569 !important;
  border-color: #cbd5e1;
}

.machine-item:hover .edit-mini {
  background: #061b3d;
  border-color: #061b3d;
  color: #ffffff;
}

.machine-item:hover .active-label {
  background: rgba(255, 255, 255, 0.22);
  color: #ffffff;
}
</style>