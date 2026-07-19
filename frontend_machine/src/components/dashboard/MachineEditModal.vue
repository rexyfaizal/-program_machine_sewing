<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  machine: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["close", "save", "delete"]);

const customName = ref("");
const location = ref("");
const pic = ref("");
const spv = ref("");

watch(
  () => props.machine,
  (machine) => {
    customName.value = machine?.customName || "";
    location.value = machine?.location === "-" ? "" : machine?.location || "";
    pic.value = machine?.pic === "-" ? "" : machine?.pic || "";
    spv.value = machine?.spv === "-" ? "" : machine?.spv || "";
  },
  { immediate: true }
);

function closeModal() {
  emit("close");
}

function saveData() {
  if (!props.machine?.uuid) return;

  emit("save", {
    uuid: props.machine.uuid,
    customName: customName.value.trim(),
    location: location.value.trim(),
    pic: pic.value.trim(),
    spv: spv.value.trim(),
  });
}

function deleteData() {
  if (!props.machine?.uuid) return;

  const ok = confirm(
    "Hapus setting manual untuk mesin ini? Nama, location, PIC, dan SPV akan kembali ke data default."
  );

  if (!ok) return;

  emit("delete", props.machine.uuid);
}
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="closeModal">
    <div class="modal">
      <div class="modal-head">
        <h3>Edit Data Manual Mesin</h3>

        <button class="btn-secondary btn-small" type="button" @click="closeModal">
          Tutup
        </button>
      </div>

      <div class="modal-body">
        <div class="field">
          <label>UUID</label>
          <input :value="machine?.uuid || ''" readonly />
        </div>

        <div class="field">
          <label>Nama Asli dari MachineInfo</label>
          <input
            :value="machine?.originalMachineName || machine?.machineName || ''"
            readonly
          />
        </div>

        <div class="field">
          <label>Nama Mesin Manual</label>
          <input
            v-model="customName"
            placeholder="Contoh: Jack Quilting Sleeve KK"
          />
        </div>

        <div class="field">
          <label>Location</label>
          <input
            v-model="location"
            placeholder="Contoh: GM1 LINE 1"
          />
        </div>

        <div class="field-row">
          <div class="field">
            <label>PIC</label>
            <input
              v-model="pic"
              placeholder="Contoh: Asep"
            />
          </div>

          <div class="field">
            <label>SPV</label>
            <input
              v-model="spv"
              placeholder="Contoh: Budi"
            />
          </div>
        </div>

        <p class="small-note">
          Kosongkan nama manual jika ingin dashboard kembali memakai nama asli dari
          <b>machineinfo</b>. Kolom <b>PIC</b> dan <b>SPV</b> akan disimpan ke
          setting manual mesin.
        </p>
      </div>

      <div class="modal-actions">
        <div>
          <button class="btn-danger" type="button" @click="deleteData">
            Hapus Setting
          </button>
        </div>

        <div class="right-actions">
          <button class="btn-secondary" type="button" @click="closeModal">
            Batal
          </button>

          <button type="button" @click="saveData">
            Simpan
          </button>
        </div>
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
  overflow: hidden;
}

.modal-head {
  padding: 18px 20px;
  border-bottom: 1px solid #dbe4ef;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  background: linear-gradient(135deg, #ffffff, #f8fbff);
}

.modal-head h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.modal-body {
  padding: 20px;
  display: grid;
  gap: 14px;
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
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

.field input {
  width: 100%;
  border: 1px solid #d8e2ee;
  border-radius: 14px;
  background: white;
  color: #0f172a;
  padding: 13px 14px;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.field input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

.field input[readonly] {
  background: #f8fafc;
  color: #64748b;
}

.small-note {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.55;
}

.modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 16px 20px 20px;
  border-top: 1px solid #dbe4ef;
  background: #ffffff;
}

.right-actions {
  display: flex;
  gap: 10px;
}

button {
  border: 0;
  border-radius: 14px;
  padding: 13px 18px;
  background: #061b3d;
  color: white;
  cursor: pointer;
  font-weight: 900;
  white-space: nowrap;
}

button:hover {
  filter: brightness(1.05);
}

.btn-secondary {
  background: white;
  color: #0f172a;
  border: 1px solid #dbe4ef;
}

.btn-danger {
  background: #dc2626;
  color: white;
}

.btn-small {
  padding: 8px 10px;
  font-size: 12px;
  border-radius: 10px;
}

@media (max-width: 650px) {
  .field-row {
    grid-template-columns: 1fr;
  }

  .modal-actions,
  .right-actions {
    flex-direction: column;
  }
}
</style>