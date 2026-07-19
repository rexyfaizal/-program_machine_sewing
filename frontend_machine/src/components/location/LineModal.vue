<script setup>
const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  mode: {
    type: String,
    default: "add",
  },
  selectedFactory: {
    type: String,
    required: true,
  },
  oldLineName: {
    type: String,
    default: "",
  },
  lineName: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["update:lineName", "close", "save"]);

function updateLineName(event) {
  emit("update:lineName", event.target.value);
}
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal">
      <div class="modal-head">
        <h3>{{ mode === "add" ? "Tambah Line" : "Rename Line" }}</h3>

        <button type="button" class="close-btn" @click="emit('close')">
          Tutup
        </button>
      </div>

      <div class="modal-body">
        <div class="field">
          <label>Factory</label>
          <input :value="selectedFactory" readonly />
        </div>

        <div class="field">
          <label>Nama Line</label>
          <input
            :value="lineName"
            placeholder="Contoh: LINE 19 / LINE BACK"
            @input="updateLineName"
          />
        </div>

        <div v-if="mode === 'rename'" class="preview-box">
          Rename dari <b>{{ oldLineName }}</b> menjadi <b>{{ lineName }}</b>.
          Mesin di line lama akan ikut dipindahkan ke nama line baru.
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
  width: min(560px, 100%);
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
}

.modal-head h3 {
  margin: 0;
  font-size: 18px;
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

.field input {
  width: 100%;
  border: 1px solid #d8e2ee;
  border-radius: 14px;
  background: white;
  color: #0f172a;
  padding: 13px 14px;
  font-size: 14px;
  outline: none;
}

.field input[readonly] {
  background: #f8fafc;
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
</style>