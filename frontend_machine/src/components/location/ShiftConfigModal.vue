<script setup>
import { computed, ref, watch } from "vue";
import { DEFAULT_SHIFT_SCHEDULE } from "../../utils/gm3Shift";

const props = defineProps({
  show: { type: Boolean, default: false },
  factory: { type: String, default: "" },
  lines: { type: Array, default: () => [] },
  savedConfigs: { type: Array, default: () => [] },
  savedShifts: { type: Array, default: () => [] },
  defaults: { type: Array, default: () => [] },
  saving: { type: Boolean, default: false },
});

const emit = defineEmits(["close", "save"]);

const draftShifts = ref([]);
const draftLines = ref([]);
const localError = ref("");

function toHHMM(value) {
  const text = String(value || "").trim();
  if (!text) return "";
  return text.length >= 5 ? text.slice(0, 5) : text;
}

function cloneFallbackSchedule() {
  const source =
    Array.isArray(props.defaults) && props.defaults.length
      ? props.defaults
      : DEFAULT_SHIFT_SCHEDULE;
  return source.map((item, index) => ({
    shiftNo: index + 1,
    shiftName: String(item.code || `SHIFT_${index + 1}`).toUpperCase(),
    startTime: toHHMM(item.start),
    endTime: toHHMM(item.end),
    breakStart: toHHMM(item.breakStart),
    breakEnd: toHHMM(item.breakEnd),
  }));
}

function mapShiftRows(rows) {
  return (Array.isArray(rows) ? rows : []).map((item, index) => ({
    shiftNo: Number(item.shiftNo) > 0 ? Number(item.shiftNo) : index + 1,
    shiftName: String(item.shiftName || `SHIFT_${index + 1}`).toUpperCase(),
    startTime: toHHMM(item.startTime),
    endTime: toHHMM(item.endTime),
    breakStart: toHHMM(item.breakStart),
    breakEnd: toHHMM(item.breakEnd),
  }));
}

function findSavedLine(lineName) {
  const key = String(lineName || "").trim().toUpperCase();
  return (props.savedConfigs || []).find(
    (item) => String(item.lineName || "").trim().toUpperCase() === key
  );
}

function buildDraftShifts() {
  const saved = Array.isArray(props.savedShifts) ? props.savedShifts : [];
  if (saved.length) return mapShiftRows(saved);
  return [];
}

function buildDraftLines() {
  return (props.lines || []).map((lineName) => {
    const saved = findSavedLine(lineName);
    const enabled =
      saved != null
        ? Boolean(saved.enabled)
        : String(props.factory || "").toUpperCase() === "GM3";
    const custom = Boolean(saved?.custom);
    return {
      lineName,
      mode: enabled ? "shift" : "normal",
      custom,
      shifts: custom ? mapShiftRows(saved?.shifts) : [],
    };
  });
}

watch(
  () => [
    props.show,
    props.factory,
    props.lines,
    props.savedConfigs,
    props.savedShifts,
    props.defaults,
  ],
  () => {
    if (props.show) {
      localError.value = "";
      draftShifts.value = buildDraftShifts();
      draftLines.value = buildDraftLines();
    }
  },
  { immediate: true, deep: true }
);

const title = computed(() => `Atur Shift — ${props.factory || "-"}`);
const shiftCount = computed(() => draftShifts.value.length);
const lineShiftCount = computed(
  () => draftLines.value.filter((l) => l.mode === "shift").length
);
const lineNormalCount = computed(
  () => draftLines.value.filter((l) => l.mode === "normal").length
);

function newShiftRow(list) {
  const nextNo = list.length + 1;
  return {
    shiftNo: nextNo,
    shiftName: `SHIFT_${nextNo}`,
    startTime: "",
    endTime: "",
    breakStart: "",
    breakEnd: "",
  };
}

function renumber(list) {
  list.forEach((shift, i) => {
    shift.shiftNo = i + 1;
    if (!String(shift.shiftName || "").trim()) shift.shiftName = `SHIFT_${i + 1}`;
  });
}

function addShift() {
  draftShifts.value.push(newShiftRow(draftShifts.value));
}
function removeShift(index) {
  draftShifts.value.splice(index, 1);
  renumber(draftShifts.value);
}
function fillDefaultShifts() {
  draftShifts.value = cloneFallbackSchedule();
}

function setLineMode(line, mode) {
  line.mode = mode === "shift" ? "shift" : "normal";
  if (line.mode === "normal") line.custom = false;
}
function setAllLinesMode(mode) {
  const next = mode === "shift" ? "shift" : "normal";
  draftLines.value.forEach((line) => {
    line.mode = next;
    if (next === "normal") line.custom = false;
  });
}
function toggleCustom(line) {
  line.custom = !line.custom;
  if (line.custom && (!line.shifts || !line.shifts.length)) {
    line.shifts = draftShifts.value.length
      ? draftShifts.value.map((s) => ({ ...s }))
      : cloneFallbackSchedule();
  }
}
function addLineShift(line) {
  if (!Array.isArray(line.shifts)) line.shifts = [];
  line.shifts.push(newShiftRow(line.shifts));
}
function removeLineShift(line, index) {
  line.shifts.splice(index, 1);
  renumber(line.shifts);
}

function resetToNormal() {
  setAllLinesMode("normal");
  draftShifts.value = [];
  localError.value = "";
}

function cleanShifts(list) {
  return (list || []).map((shift, index) => ({
    shiftNo: index + 1,
    shiftName: String(shift.shiftName || `SHIFT_${index + 1}`).trim().toUpperCase(),
    startTime: String(shift.startTime || "").trim(),
    endTime: String(shift.endTime || "").trim(),
    breakStart: String(shift.breakStart || "").trim(),
    breakEnd: String(shift.breakEnd || "").trim(),
  }));
}

function hasEmptyTime(list) {
  return (list || []).some(
    (s) => !String(s.startTime || "").trim() || !String(s.endTime || "").trim()
  );
}

function onSave() {
  localError.value = "";

  if (hasEmptyTime(draftShifts.value)) {
    localError.value = "Isi jam mulai & selesai untuk setiap shift jadwal area.";
    return;
  }

  for (const line of draftLines.value) {
    if (line.mode === "shift" && line.custom) {
      if (!line.shifts || !line.shifts.length) {
        localError.value = `Line ${line.lineName}: mode Custom tapi belum ada shift.`;
        return;
      }
      if (hasEmptyTime(line.shifts)) {
        localError.value = `Line ${line.lineName}: isi jam mulai & selesai tiap shift custom.`;
        return;
      }
    }
    if (line.mode === "shift" && !line.custom && !draftShifts.value.length) {
      localError.value = `Line ${line.lineName}: mode Shift butuh jadwal default area. Isi jadwal area atau pakai Custom.`;
      return;
    }
  }

  const shifts = cleanShifts(draftShifts.value);
  const lines = draftLines.value.map((line) => ({
    lineName: line.lineName,
    enabled: line.mode === "shift",
    custom: line.mode === "shift" && line.custom,
    shifts: line.mode === "shift" && line.custom ? cleanShifts(line.shifts) : [],
  }));

  emit("save", { area: props.factory, factory: props.factory, shifts, lines });
}
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-head">
        <div>
          <h3>{{ title }}</h3>
          <p class="head-sub">
            Jadwal default area dipakai bersama. Tiap line bisa Normal, Shift
            (pakai default), atau Shift Custom (jadwal sendiri).
          </p>
        </div>
        <button type="button" class="close-btn" @click="emit('close')">Tutup</button>
      </div>

      <div class="modal-body">
        <div v-if="localError" class="local-error">{{ localError }}</div>

        <section class="panel">
          <div class="panel-head">
            <div>
              <h4>1. Jadwal Default Area</h4>
              <p class="panel-hint">{{ shiftCount }} shift · {{ factory || "-" }}</p>
            </div>
            <div class="bulk-actions">
              <button type="button" class="bulk-btn" @click="resetToNormal">
                Kembali ke Normal
              </button>
              <button type="button" class="add-btn" @click="addShift">
                + Tambah Shift
              </button>
            </div>
          </div>

          <div v-if="!draftShifts.length" class="empty">
            Tidak ada jadwal default. Line mode Shift (non-custom) akan invalid.
            <div class="empty-actions">
              <button type="button" class="bulk-btn primary" @click="fillDefaultShifts">
                Isi default 3 shift
              </button>
            </div>
          </div>

          <div v-else class="schedule-table-wrap">
            <table class="schedule-table">
              <thead>
                <tr>
                  <th>Nama</th>
                  <th>Mulai</th>
                  <th>Selesai</th>
                  <th>Istirahat mulai</th>
                  <th>Istirahat selesai</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(shift, index) in draftShifts" :key="`s-${index}`">
                  <td>
                    <input v-model="shift.shiftName" type="text" maxlength="32" class="name-input" />
                  </td>
                  <td><input v-model="shift.startTime" type="time" /></td>
                  <td><input v-model="shift.endTime" type="time" /></td>
                  <td><input v-model="shift.breakStart" type="time" /></td>
                  <td><input v-model="shift.breakEnd" type="time" /></td>
                  <td>
                    <button type="button" class="remove-btn" @click="removeShift(index)">
                      Hapus
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="panel">
          <div class="panel-head">
            <div>
              <h4>2. Mode per Line</h4>
              <p class="panel-hint">
                Shift {{ lineShiftCount }} · Normal {{ lineNormalCount }}
              </p>
            </div>
            <div class="bulk-actions">
              <button type="button" class="bulk-btn" @click="setAllLinesMode('normal')">
                Semua Normal
              </button>
              <button type="button" class="bulk-btn primary" @click="setAllLinesMode('shift')">
                Semua Shift
              </button>
            </div>
          </div>

          <div v-if="!draftLines.length" class="empty">
            Belum ada line di factory ini.
          </div>

          <div v-else class="line-list">
            <article v-for="line in draftLines" :key="line.lineName" class="line-card">
              <div class="line-row">
                <strong class="line-name">{{ line.lineName }}</strong>

                <div class="mode-switch">
                  <button
                    type="button"
                    class="mode-btn"
                    :class="{ active: line.mode === 'normal' }"
                    @click="setLineMode(line, 'normal')"
                  >
                    Normal
                  </button>
                  <button
                    type="button"
                    class="mode-btn"
                    :class="{ active: line.mode === 'shift' }"
                    @click="setLineMode(line, 'shift')"
                  >
                    Shift
                  </button>
                </div>

                <label v-if="line.mode === 'shift'" class="custom-toggle">
                  <input type="checkbox" :checked="line.custom" @change="toggleCustom(line)" />
                  <span>Custom</span>
                </label>

                <span class="tag" :class="line.mode === 'shift' ? (line.custom ? 'custom' : 'shift') : 'normal'">
                  <template v-if="line.mode !== 'shift'">Hari penuh</template>
                  <template v-else-if="line.custom">Jadwal sendiri</template>
                  <template v-else>Pakai jadwal area</template>
                </span>
              </div>

              <div v-if="line.mode === 'shift' && line.custom" class="line-custom">
                <div class="custom-head">
                  <span>Jadwal khusus {{ line.lineName }}</span>
                  <button type="button" class="add-btn small" @click="addLineShift(line)">
                    + Shift
                  </button>
                </div>
                <div class="schedule-table-wrap">
                  <table class="schedule-table">
                    <thead>
                      <tr>
                        <th>Nama</th>
                        <th>Mulai</th>
                        <th>Selesai</th>
                        <th>Istirahat mulai</th>
                        <th>Istirahat selesai</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(shift, index) in line.shifts" :key="`ls-${index}`">
                        <td>
                          <input v-model="shift.shiftName" type="text" maxlength="32" class="name-input" />
                        </td>
                        <td><input v-model="shift.startTime" type="time" /></td>
                        <td><input v-model="shift.endTime" type="time" /></td>
                        <td><input v-model="shift.breakStart" type="time" /></td>
                        <td><input v-model="shift.breakEnd" type="time" /></td>
                        <td>
                          <button type="button" class="remove-btn" @click="removeLineShift(line, index)">
                            Hapus
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>

      <div class="modal-actions">
        <button type="button" class="cancel-btn" @click="emit('close')">Batal</button>
        <button type="button" class="save-btn" :disabled="saving" @click="onSave">
          {{ saving ? "Menyimpan..." : "Simpan" }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px;
  z-index: 999;
}

.modal {
  background: #fff;
  width: min(1000px, 100%);
  max-height: min(92vh, 980px);
  border-radius: 20px;
  border: 1px solid #dbe4ef;
  box-shadow: 0 24px 64px rgba(15, 23, 42, 0.24);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-head {
  padding: 18px 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.modal-head h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.head-sub {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
  max-width: 640px;
}

.close-btn,
.cancel-btn,
.save-btn,
.add-btn,
.remove-btn,
.bulk-btn {
  height: 38px;
  padding: 0 14px;
  border-radius: 11px;
  font-weight: 800;
  cursor: pointer;
}

.add-btn.small {
  height: 32px;
  padding: 0 10px;
}

.close-btn,
.cancel-btn,
.remove-btn,
.bulk-btn {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #0f172a;
}

.add-btn,
.bulk-btn.primary {
  border: 1px solid #93c5fd;
  background: #eff6ff;
  color: #1d4ed8;
}

.save-btn {
  border: 1px solid #2563eb;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff;
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.modal-body {
  padding: 16px 20px;
  overflow: auto;
  display: grid;
  gap: 16px;
}

.local-error {
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid #fecaca;
  background: #fef2f2;
  color: #b91c1c;
  font-size: 13px;
  font-weight: 700;
}

.panel {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f8fafc;
  padding: 14px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-head h4 {
  margin: 0;
  font-size: 15px;
  color: #0f172a;
}

.panel-hint {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 12px;
  font-weight: 700;
}

.bulk-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.schedule-table-wrap {
  overflow: auto;
  border-radius: 12px;
  border: 1px solid #dbe4ef;
  background: #fff;
}

.schedule-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 720px;
}

.schedule-table th,
.schedule-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #e2e8f0;
  text-align: left;
  vertical-align: middle;
}

.schedule-table th {
  font-size: 11px;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  color: #64748b;
  background: #f1f5f9;
}

.schedule-table input,
.name-input {
  width: 100%;
  height: 36px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 0 8px;
  font-weight: 700;
  color: #0f172a;
  background: #fff;
}

.line-list {
  display: grid;
  gap: 10px;
}

.line-card {
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #fff;
  padding: 12px;
}

.line-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.line-name {
  color: #0f172a;
  min-width: 90px;
}

.mode-switch {
  display: inline-flex;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  overflow: hidden;
  background: #f8fafc;
}

.mode-btn {
  border: 0;
  background: transparent;
  height: 34px;
  padding: 0 14px;
  font-weight: 800;
  color: #64748b;
  cursor: pointer;
}

.mode-btn.active {
  background: #2563eb;
  color: #fff;
}

.custom-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 800;
  color: #0f172a;
  cursor: pointer;
}

.tag {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
  margin-left: auto;
}

.tag.shift {
  background: #dbeafe;
  color: #1d4ed8;
}

.tag.custom {
  background: #fef3c7;
  color: #b45309;
}

.tag.normal {
  background: #e2e8f0;
  color: #334155;
}

.line-custom {
  margin-top: 12px;
  border-top: 1px dashed #cbd5e1;
  padding-top: 12px;
}

.custom-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 800;
  color: #334155;
  font-size: 13px;
}

.empty {
  padding: 18px;
  border: 1px dashed #cbd5e1;
  border-radius: 12px;
  color: #64748b;
  font-weight: 700;
  background: #fff;
}

.empty-actions {
  margin-top: 12px;
}

.modal-actions {
  padding: 14px 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 780px) {
  .panel-head {
    flex-direction: column;
    align-items: flex-start;
  }
  .schedule-table {
    min-width: 640px;
  }
}
</style>
