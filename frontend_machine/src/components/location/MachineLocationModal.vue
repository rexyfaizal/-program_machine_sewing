<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  mode: {
    type: String,
    default: "add",
  },
  form: {
    type: Object,
    required: true,
  },
  machines: {
    type: Array,
    default: () => [],
  },
  factoryOptions: {
    type: Array,
    default: () => [],
  },
  lineLayout: {
    type: Object,
    default: () => ({}),
  },
  makeLocation: {
    type: Function,
    required: true,
  },
});

const emit = defineEmits(["close", "save", "selectMachine"]);

const machineSearch = ref("");
const dropdownOpen = ref(false);

const selectedMachine = computed(() => {
  return props.machines.find((m) => m.uuid === props.form.uuid) || null;
});

const filteredMachines = computed(() => {
  const keyword = machineSearch.value.trim().toLowerCase();

  if (!keyword) {
    return props.machines.slice(0, 80);
  }

  return props.machines
    .filter((machine) => {
      const text = [
        machine.machineName,
        machine.originalMachineName,
        machine.ip,
        machine.uuid,
        machine.location,
      ]
        .join(" ")
        .toLowerCase();

      return text.includes(keyword);
    })
    .slice(0, 80);
});

function machineSearchLabel(machine) {
  if (!machine) return "";

  return `${machine.machineName || "-"} - ${machine.ip || "-"} - ${machine.uuid || "-"}`;
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      machineSearch.value = selectedMachine.value
        ? machineSearchLabel(selectedMachine.value)
        : "";
      dropdownOpen.value = false;
    }
  }
);

watch(
  () => props.form.uuid,
  () => {
    if (selectedMachine.value) {
      machineSearch.value = machineSearchLabel(selectedMachine.value);
    }
  }
);

function openDropdown() {
  dropdownOpen.value = true;
}

function closeDropdownDelay() {
  setTimeout(() => {
    dropdownOpen.value = false;
  }, 150);
}

function selectMachine(machine) {
  props.form.uuid = machine.uuid;
  machineSearch.value = machineSearchLabel(machine);
  dropdownOpen.value = false;
  emit("selectMachine");
}

function clearMachine() {
  props.form.uuid = "";
  machineSearch.value = "";
  dropdownOpen.value = true;
  emit("selectMachine");
}
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-head">
        <h3>
          {{ mode === "add" ? "Tambah Mesin ke Line" : "Edit Location Mesin" }}
        </h3>

        <button type="button" class="close-btn" @click="emit('close')">
          Tutup
        </button>
      </div>

      <div class="modal-body">
        <div class="field">
          <label>Factory</label>

          <select v-model="form.factory">
            <option
              v-for="factory in factoryOptions"
              :key="factory.key"
              :value="factory.key"
            >
              {{ factory.label }}
            </option>
          </select>
        </div>

        <div class="field">
          <label>Line</label>

          <select v-model="form.line">
            <option
              v-for="line in lineLayout[form.factory]"
              :key="line"
              :value="line"
            >
              {{ line }}
            </option>
          </select>
        </div>

        <div class="field">
          <label>Pilih Mesin</label>

          <div class="search-select">
            <div class="search-input-wrap">
              <input
                v-model="machineSearch"
                type="text"
                placeholder="Cari nama mesin, IP, UUID, atau location..."
                @focus="openDropdown"
                @input="openDropdown"
                @blur="closeDropdownDelay"
              />

              <button
                v-if="form.uuid"
                type="button"
                class="clear-btn"
                @mousedown.prevent="clearMachine"
              >
                Bersihkan
              </button>
            </div>

            <div v-if="dropdownOpen" class="machine-dropdown">
              <button
                v-for="machine in filteredMachines"
                :key="machine.uuid"
                type="button"
                class="machine-option"
                :class="{ active: machine.uuid === form.uuid }"
                @mousedown.prevent="selectMachine(machine)"
              >
                <strong>{{ machine.machineName || "-" }}</strong>

                <small>
                  IP: {{ machine.ip || "-" }} · Location: {{ machine.location || "-" }}
                </small>

                <small class="uuid-line">
                  UUID: {{ machine.uuid || "-" }}
                </small>
              </button>

              <div v-if="!filteredMachines.length" class="no-result">
                Mesin tidak ditemukan.
              </div>
            </div>
          </div>

          <div v-if="selectedMachine" class="selected-info-box">
            <div>
              <span>Dipilih</span>
              <strong>{{ selectedMachine.machineName || "-" }}</strong>
            </div>

            <div>
              <span>IP</span>
              <strong>{{ selectedMachine.ip || "-" }}</strong>
            </div>

            <div>
              <span>UUID</span>
              <strong class="uuid-value">{{ selectedMachine.uuid || "-" }}</strong>
            </div>

            <div>
              <span>Location Saat Ini</span>
              <strong>{{ selectedMachine.location || "-" }}</strong>
            </div>
          </div>
        </div>

        <div class="field">
          <label>Nama Mesin Manual</label>

          <input
            v-model="form.customName"
            placeholder="Kosongkan jika memakai nama asli"
          />
        </div>

        <div class="preview-box">
          Location akan disimpan sebagai:
          <b>{{ makeLocation(form.factory, form.line) }}</b>
        </div>
      </div>

      <div class="modal-actions">
        <button type="button" class="cancel-btn" @click="emit('close')">
          Batal
        </button>

        <button type="button" class="save-btn" @click="emit('save')">
          Simpan
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
  width: min(620px, 100%);
  border-radius: 22px;
  border: 1px solid #dbe4ef;
  box-shadow: 0 25px 70px rgba(15, 23, 42, 0.26);
  overflow: visible;
}

.modal-head {
  padding: 18px 20px;
  border-bottom: 1px solid #dbe4ef;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.modal-head h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.close-btn,
.cancel-btn {
  background: white;
  color: #0f172a;
  border: 1px solid #dbe4ef;
  border-radius: 12px;
  padding: 10px 12px;
  font-weight: 900;
  cursor: pointer;
}

.close-btn:hover,
.cancel-btn:hover {
  background: #f8fafc;
}

.modal-body {
  padding: 20px;
  display: grid;
  gap: 14px;
}

.field label {
  display: block;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: 900;
  color: #64748b;
  margin-bottom: 7px;
}

.field input,
.field select {
  width: 100%;
  border: 1px solid #d8e2ee;
  border-radius: 14px;
  background: white;
  color: #0f172a;
  padding: 13px 14px;
  font-size: 14px;
  outline: none;
}

.field input:focus,
.field select:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}

.search-select {
  position: relative;
}

.search-input-wrap {
  position: relative;
}

.search-input-wrap input {
  padding-right: 96px;
}

.clear-btn {
  position: absolute;
  top: 50%;
  right: 8px;
  transform: translateY(-50%);
  border: 0;
  background: #eff6ff;
  color: #1d4ed8;
  border-radius: 10px;
  padding: 7px 9px;
  font-size: 11px;
  font-weight: 900;
  cursor: pointer;
}

.clear-btn:hover {
  background: #dbeafe;
}

.machine-dropdown {
  position: absolute;
  left: 0;
  right: 0;
  top: calc(100% + 6px);
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 14px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.18);
  max-height: 320px;
  overflow-y: auto;
  padding: 8px;
  z-index: 1005;
}

.machine-option {
  width: 100%;
  border: 0;
  background: transparent;
  text-align: left;
  border-radius: 12px;
  padding: 11px 12px;
  cursor: pointer;
  border: 1px solid transparent;
}

.machine-option:hover,
.machine-option.active {
  background: #2563eb;
  color: white;
  border-color: #2563eb;
}

.machine-option strong {
  display: block;
  font-size: 13px;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.machine-option small {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.machine-option .uuid-line {
  color: #2563eb;
  font-family: Consolas, Menlo, monospace;
  font-weight: 900;
  word-break: break-all;
  white-space: normal;
}

.machine-option:hover small,
.machine-option.active small,
.machine-option:hover .uuid-line,
.machine-option.active .uuid-line {
  color: rgba(255, 255, 255, 0.88);
}

.no-result {
  padding: 14px;
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
  text-align: center;
}

.selected-info-box {
  margin-top: 10px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 12px;
  border-radius: 14px;
  background: #f8fbff;
  border: 1px solid #dbeafe;
}

.selected-info-box div {
  min-width: 0;
}

.selected-info-box span {
  display: block;
  color: #64748b;
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 4px;
}

.selected-info-box strong {
  display: block;
  color: #0f172a;
  font-size: 12px;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-info-box .uuid-value {
  font-family: Consolas, Menlo, monospace;
  color: #2563eb;
}

.preview-box {
  background: #f8fafc;
  border: 1px solid #dbe4ef;
  border-radius: 14px;
  padding: 12px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.5;
}

.preview-box b {
  color: #0f172a;
}

.modal-actions {
  padding: 16px 20px 20px;
  border-top: 1px solid #dbe4ef;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.save-btn {
  background: #2563eb;
  color: white;
  border: 0;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 900;
  cursor: pointer;
}

.save-btn:hover {
  background: #1d4ed8;
}

@media (max-width: 560px) {
  .modal-body {
    padding: 16px;
  }

  .selected-info-box {
    grid-template-columns: 1fr;
  }

  .modal-actions {
    flex-direction: column-reverse;
  }

  .save-btn,
  .cancel-btn {
    width: 100%;
  }
}
</style>