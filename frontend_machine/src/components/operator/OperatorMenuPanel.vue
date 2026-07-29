<script setup>
const props = defineProps({
  activeSession: {
    type: Object,
    default: null,
  },
  machine: {
    type: Object,
    default: null,
  },
  activeNotes: {
    type: Array,
    default: () => [],
  },
  reasonMenus: {
    type: Array,
    default: () => [],
  },
  activeLossEvent: {
    type: Object,
    default: null,
  },
  activeLossDurationText: {
    type: String,
    default: "00:00:00",
  },
  lossEventError: {
    type: String,
    default: "",
  },
  noteSaving: Boolean,
  lossEventLoading: Boolean,
  formatDateTime: {
    type: Function,
    required: true,
  },
  formatTime: {
    type: Function,
    required: true,
  },
});

const otherNote = defineModel("otherNote", {
  type: String,
  default: "",
});

const emit = defineEmits([
  "submit-note",
  "finish-loss-event",
  "login-operator-baru",
]);

function getLossReasonLabel(event) {
  return (
    event?.reasonLabel ||
    event?.reasonName ||
    event?.reasonCode ||
    "Loss Time"
  );
}

function getLossOperatorName(event) {
  return event?.operatorName || props.activeSession?.operatorName || "-";
}

function getLossOperatorNik(event) {
  return event?.operatorNik || props.activeSession?.operatorNik || "-";
}

function isOtherReason(reason) {
  return String(reason?.reasonCode || "").toUpperCase() === "OTHER";
}

function submitReason(reason) {
  emit("submit-note", reason);
}
</script>

<template>
  <div v-if="activeSession" class="active-box">
    <div>
      <span>Operator Aktif</span>
      <strong>
        {{ activeSession.operatorNik }} - {{ activeSession.operatorName }}
      </strong>
    </div>

    <div>
      <span>Mesin</span>
      <strong>{{ machine?.machineName || activeSession.machineName || "-" }}</strong>
    </div>

    <div>
      <span>Proses</span>
      <strong>{{ activeSession.processName || "-" }}</strong>
    </div>

    <div>
      <span>Style</span>
      <strong>{{ activeSession.styleName || "-" }}</strong>
    </div>

    <div>
      <span>Login</span>
      <strong>{{ formatDateTime(activeSession.loginTime) }}</strong>
    </div>

    <div>
      <span>Status</span>
      <strong>{{ activeSession.status || "ACTIVE" }}</strong>
    </div>
  </div>

  <section class="menu-section">
    <div v-if="lossEventError" class="error-box">
      {{ lossEventError }}
    </div>

    <div v-if="activeLossEvent" class="loss-active-card">
      <div class="loss-header">
        <div>
          <span class="loss-label">Loss Time Aktif</span>
          <h3>{{ getLossReasonLabel(activeLossEvent) }}</h3>
        </div>

        <strong class="loss-duration">
          {{ activeLossDurationText || "00:00:00" }}
        </strong>
      </div>

      <div class="loss-info-grid">
        <div>
          <span>Operator</span>
          <strong>
            {{ getLossOperatorNik(activeLossEvent) }} -
            {{ getLossOperatorName(activeLossEvent) }}
          </strong>
        </div>

        <div>
          <span>Mulai</span>
          <strong>{{ formatDateTime(activeLossEvent.startTime) }}</strong>
        </div>

        <div>
          <span>Reason</span>
          <strong>{{ getLossReasonLabel(activeLossEvent) }}</strong>
        </div>

        <div>
          <span>Status</span>
          <strong>{{ activeLossEvent.status || "ACTIVE" }}</strong>
        </div>
      </div>

      <p v-if="activeLossEvent.note" class="loss-note">
        {{ activeLossEvent.note }}
      </p>

      <button
        type="button"
        class="btn-finish full"
        :disabled="noteSaving || lossEventLoading"
        @click="emit('finish-loss-event')"
      >
        {{ lossEventLoading ? "Menyelesaikan..." : "Selesai / Kembali Kerja" }}
      </button>

      <p class="loss-helper">
        Selesaikan loss time ini dulu sebelum memilih alasan lain atau login operator baru.
      </p>
    </div>

    <template v-else>
      <div class="note-field">
        <label>
          Catatan tambahan
          <textarea
            v-model="otherNote"
            placeholder="Opsional. Wajib diisi kalau pilih Other."
            :disabled="noteSaving || lossEventLoading"
          ></textarea>
        </label>
      </div>

      <div class="reason-grid">
        <button
          v-for="reason in reasonMenus"
          :key="reason.reasonCode"
          type="button"
          class="reason-btn"
          :class="{ other: isOtherReason(reason) }"
          :disabled="noteSaving || lossEventLoading"
          @click="submitReason(reason)"
        >
          {{ reason.reasonName }}
        </button>
      </div>

      <button
        type="button"
        class="btn-soft full"
        :disabled="noteSaving || lossEventLoading"
        @click="emit('login-operator-baru')"
      >
        Login Operator Baru
      </button>
    </template>

    <div v-if="activeNotes.length" class="last-notes">
      <h3>Keterangan Terakhir</h3>

      <div
        v-for="note in activeNotes"
        :key="`note-${note.id}-${note.createdAt}`"
        class="note-row"
      >
        <strong>{{ note.reasonName }}</strong>
        <span>{{ formatTime(note.createdAt) }}</span>
        <p v-if="note.note">{{ note.note }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.active-box {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  padding: 14px;
  border-radius: 18px;
  background: linear-gradient(135deg, #eff6ff, #ffffff);
  border: 1px solid #bfdbfe;
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.05);
}

.active-box div {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.active-box span {
  color: #64748b;
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.active-box strong {
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  overflow-wrap: anywhere;
}

.menu-section {
  display: grid;
  gap: 14px;
  margin-top: 14px;
}

.error-box {
  padding: 12px 14px;
  border-radius: 14px;
  background: #fee2e2;
  border: 1px solid #fecaca;
  color: #991b1b;
  font-size: 13px;
  font-weight: 900;
}

.note-field label {
  display: grid;
  gap: 8px;
  color: #334155;
  font-size: 13px;
  font-weight: 900;
}

.note-field textarea {
  width: 100%;
  min-height: 92px;
  resize: vertical;
  border: 1px solid #dbeafe;
  border-radius: 16px;
  padding: 12px 14px;
  outline: none;
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
  background: #ffffff;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.04);
}

.note-field textarea:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

.note-field textarea:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.reason-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.reason-btn {
  min-height: 54px;
  border: 0;
  border-radius: 16px;
  padding: 12px 14px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #ffffff;
  font-size: 14px;
  font-weight: 1000;
  cursor: pointer;
  box-shadow: 0 12px 22px rgba(37, 99, 235, 0.22);
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    filter 0.18s ease,
    opacity 0.18s ease;
}

.reason-btn.other {
  background: linear-gradient(135deg, #64748b, #475569);
  box-shadow: 0 12px 22px rgba(71, 85, 105, 0.22);
}

.reason-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.03);
  box-shadow: 0 16px 28px rgba(37, 99, 235, 0.28);
}

.reason-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
  filter: none;
}

.btn-soft,
.btn-finish {
  border: 0;
  border-radius: 16px;
  min-height: 52px;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 1000;
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    filter 0.18s ease,
    opacity 0.18s ease;
}

.btn-soft {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.btn-soft:hover:not(:disabled) {
  background: #f8fafc;
}

.btn-finish {
  background: linear-gradient(135deg, #16a34a, #15803d);
  color: #ffffff;
  box-shadow: 0 14px 24px rgba(22, 163, 74, 0.24);
}

.btn-finish:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.03);
  box-shadow: 0 18px 30px rgba(22, 163, 74, 0.3);
}

.btn-soft:disabled,
.btn-finish:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
  filter: none;
}

.full {
  width: 100%;
}

.loss-active-card {
  display: grid;
  gap: 14px;
  padding: 16px;
  border-radius: 20px;
  background: linear-gradient(135deg, #fff7ed, #ffffff);
  border: 1px solid #fed7aa;
  box-shadow: 0 14px 28px rgba(234, 88, 12, 0.08);
}

.loss-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 14px;
}

.loss-label {
  display: inline-flex;
  width: fit-content;
  border-radius: 999px;
  padding: 6px 10px;
  background: #ffedd5;
  color: #c2410c;
  font-size: 11px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.07em;
}

.loss-header h3 {
  margin: 8px 0 0;
  color: #0f172a;
  font-size: 22px;
  font-weight: 1000;
  letter-spacing: -0.03em;
}

.loss-duration {
  display: inline-flex;
  justify-content: center;
  min-width: 120px;
  border-radius: 16px;
  padding: 10px 12px;
  background: #0f172a;
  color: #ffffff;
  font-size: 18px;
  font-weight: 1000;
  font-family: Consolas, Menlo, monospace;
}

.loss-info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.loss-info-grid div {
  display: grid;
  gap: 5px;
  padding: 10px 12px;
  border-radius: 14px;
  background: #ffffff;
  border: 1px solid #fed7aa;
}

.loss-info-grid span {
  color: #9a3412;
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.loss-info-grid strong {
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  overflow-wrap: anywhere;
}

.loss-note {
  margin: 0;
  padding: 11px 12px;
  border-radius: 14px;
  background: #ffffff;
  border: 1px dashed #fdba74;
  color: #334155;
  font-size: 13px;
  font-weight: 800;
}

.loss-helper {
  margin: -4px 0 0;
  color: #9a3412;
  font-size: 12px;
  font-weight: 800;
  line-height: 1.45;
}

.last-notes {
  display: grid;
  gap: 9px;
  padding: 14px;
  border-radius: 18px;
  background: #ffffff;
  border: 1px solid #dbeafe;
}

.last-notes h3 {
  margin: 0;
  color: #0f172a;
  font-size: 15px;
  font-weight: 1000;
}

.note-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 4px 12px;
  padding: 10px 0;
  border-top: 1px solid #e2e8f0;
}

.note-row strong {
  color: #0f172a;
  font-size: 13px;
  font-weight: 1000;
}

.note-row span {
  color: #64748b;
  font-size: 12px;
  font-weight: 900;
}

.note-row p {
  grid-column: 1 / -1;
  margin: 0;
  color: #475569;
  font-size: 12px;
  font-weight: 800;
  line-height: 1.45;
}

@media (max-width: 760px) {
  .active-box {
    grid-template-columns: 1fr;
  }

  .reason-grid {
    grid-template-columns: 1fr;
  }

  .loss-header {
    display: grid;
  }

  .loss-duration {
    width: 100%;
  }

  .loss-info-grid {
    grid-template-columns: 1fr;
  }
}
</style>