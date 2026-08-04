<script setup>
import { computed, ref, watch } from "vue";
import { DEFAULT_SHIFT_SCHEDULE } from "../../utils/gm3Shift";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  factory: {
    type: String,
    default: "",
  },
  lines: {
    type: Array,
    default: () => [],
  },
  savedConfigs: {
    type: Array,
    default: () => [],
  },
  saving: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["close", "save"]);

const draftLines = ref([]);

const SHIFT_DEFS = [
  { code: "SHIFT_1", label: "Shift 1" },
  { code: "SHIFT_2", label: "Shift 2" },
  { code: "SHIFT_3", label: "Shift 3" },
];

function cloneDefaultSchedule() {
  return DEFAULT_SHIFT_SCHEDULE.map((item) => ({ ...item }));
}

function findSaved(lineName) {
  const key = String(lineName || "").trim().toUpperCase();
  return (props.savedConfigs || []).find(
    (item) => String(item.lineName || "").trim().toUpperCase() === key
  );
}

function buildDraft() {
  return (props.lines || []).map((lineName) => {
    const saved = findSaved(lineName);
    const scheduleSource =
      Array.isArray(saved?.schedule) && saved.schedule.length
        ? saved.schedule
        : cloneDefaultSchedule();

    const scheduleMap = {};
    scheduleSource.forEach((item) => {
      const code = String(item.code || "").toUpperCase();
      scheduleMap[code] = {
        code,
        start: item.start || "",
        end: item.end || "",
        breakStart: item.breakStart || "",
        breakEnd: item.breakEnd || "",
        active: true,
      };
    });

    SHIFT_DEFS.forEach((def) => {
      if (!scheduleMap[def.code]) {
        const fallback = DEFAULT_SHIFT_SCHEDULE.find((s) => s.code === def.code);
        scheduleMap[def.code] = {
          code: def.code,
          start: fallback?.start || "",
          end: fallback?.end || "",
          breakStart: fallback?.breakStart || "",
          breakEnd: fallback?.breakEnd || "",
          active: false,
        };
      }
    });

    // Legacy GM3 tanpa row tersimpan: anggap pakai shift default.
    const enabled =
      saved != null
        ? Boolean(saved.enabled)
        : String(props.factory || "").toUpperCase() === "GM3";

    if (saved == null && enabled) {
      SHIFT_DEFS.forEach((def) => {
        scheduleMap[def.code].active = true;
      });
    }

    return {
      lineName,
      enabled,
      shifts: SHIFT_DEFS.map((def) => ({ ...scheduleMap[def.code] })),
    };
  });
}

watch(
  () => [props.show, props.factory, props.lines, props.savedConfigs],
  () => {
    if (props.show) {
      draftLines.value = buildDraft();
    }
  },
  { immediate: true, deep: true }
);

const title = computed(() => `Atur Shift — ${props.factory || "-"}`);

function toggleEnabled(line) {
  line.enabled = !line.enabled;
}

function toggleShift(shift) {
  shift.active = !shift.active;
}

function onSave() {
  const lines = draftLines.value.map((line) => {
    const schedule = line.enabled
      ? line.shifts
          .filter((s) => s.active)
          .map((s) => ({
            code: s.code,
            start: String(s.start || "").trim(),
            end: String(s.end || "").trim(),
            breakStart: String(s.breakStart || "").trim(),
            breakEnd: String(s.breakEnd || "").trim(),
          }))
      : [];

    return {
      lineName: line.lineName,
      enabled: Boolean(line.enabled),
      schedule,
    };
  });

  emit("save", {
    factory: props.factory,
    lines,
  });
}
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-head">
        <h3>{{ title }}</h3>
        <button type="button" class="close-btn" @click="emit('close')">
          Tutup
        </button>
      </div>

      <div class="modal-body">
        <p class="hint">
          Aktifkan shift per line, lalu isi jam mulai–selesai. Istirahat opsional.
          Line nonaktif dihitung hari penuh.
        </p>

        <div v-if="!draftLines.length" class="empty">
          Belum ada line di factory ini.
        </div>

        <article
          v-for="line in draftLines"
          :key="line.lineName"
          class="line-block"
        >
          <div class="line-head">
            <strong>{{ line.lineName }}</strong>
            <label class="toggle">
              <input
                type="checkbox"
                :checked="line.enabled"
                @change="toggleEnabled(line)"
              />
              <span>{{ line.enabled ? "Pakai Shift" : "Hari penuh" }}</span>
            </label>
          </div>

          <div v-if="line.enabled" class="shift-list">
            <div
              v-for="shift in line.shifts"
              :key="shift.code"
              class="shift-row"
              :class="{ inactive: !shift.active }"
            >
              <label class="shift-check">
                <input
                  type="checkbox"
                  :checked="shift.active"
                  @change="toggleShift(shift)"
                />
                <span>{{ shift.code.replace("SHIFT_", "S") }}</span>
              </label>

              <div class="time-grid">
                <label>
                  Mulai
                  <input
                    v-model="shift.start"
                    type="time"
                    :disabled="!shift.active"
                  />
                </label>
                <label>
                  Selesai
                  <input
                    v-model="shift.end"
                    type="time"
                    :disabled="!shift.active"
                  />
                </label>
                <label>
                  Istirahat mulai
                  <input
                    v-model="shift.breakStart"
                    type="time"
                    :disabled="!shift.active"
                  />
                </label>
                <label>
                  Istirahat selesai
                  <input
                    v-model="shift.breakEnd"
                    type="time"
                    :disabled="!shift.active"
                  />
                </label>
              </div>
            </div>
          </div>
        </article>
      </div>

      <div class="modal-actions">
        <button type="button" class="cancel-btn" @click="emit('close')">
          Batal
        </button>
        <button
          type="button"
          class="save-btn"
          :disabled="saving"
          @click="onSave"
        >
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
  background: rgba(15, 23, 42, 0.48);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px;
  z-index: 999;
}

.modal {
  background: white;
  width: min(860px, 100%);
  max-height: min(90vh, 920px);
  border-radius: 22px;
  border: 1px solid #dbe4ef;
  box-shadow: 0 25px 70px rgba(15, 23, 42, 0.26);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-head {
  padding: 18px 20px;
  border-bottom: 1px solid #dbe4ef;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.modal-head h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.close-btn,
.cancel-btn,
.save-btn {
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  font-weight: 800;
  cursor: pointer;
}

.close-btn,
.cancel-btn {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #0f172a;
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
}

.hint {
  margin: 0 0 14px;
  color: #475569;
  font-size: 13px;
  line-height: 1.45;
}

.empty {
  padding: 18px;
  border: 1px dashed #cbd5e1;
  border-radius: 14px;
  color: #64748b;
  font-weight: 700;
}

.line-block {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 12px 14px;
  margin-bottom: 12px;
  background: #f8fafc;
}

.line-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.toggle,
.shift-check {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 800;
  color: #0f172a;
  cursor: pointer;
}

.shift-list {
  display: grid;
  gap: 10px;
}

.shift-row {
  background: #fff;
  border: 1px solid #dbeafe;
  border-radius: 12px;
  padding: 10px 12px;
}

.shift-row.inactive {
  opacity: 0.55;
}

.time-grid {
  margin-top: 8px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.time-grid label {
  display: grid;
  gap: 4px;
  font-size: 11px;
  font-weight: 800;
  color: #64748b;
}

.time-grid input {
  height: 36px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 0 8px;
  font-weight: 700;
  color: #0f172a;
}

.modal-actions {
  padding: 14px 20px;
  border-top: 1px solid #dbe4ef;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 720px) {
  .time-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .line-head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
