<script setup>
defineProps({
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
  noteSaving: Boolean,
  formatDateTime: {
    type: Function,
    required: true,
  },
  formatTime: {
    type: Function,
    required: true,
  },
});

const otherNote = defineModel("otherNote", { type: String, default: "" });

const emit = defineEmits(["submit-note", "login-operator-baru"]);
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
    <div class="note-field">
      <label>
        Catatan tambahan
        <textarea
          v-model="otherNote"
          placeholder="Opsional. Wajib diisi kalau pilih Other."
        ></textarea>
      </label>
    </div>

    <div class="reason-grid">
      <button
        v-for="reason in reasonMenus"
        :key="reason.reasonCode"
        type="button"
        class="reason-btn"
        :disabled="noteSaving"
        @click="emit('submit-note', reason)"
      >
        {{ reason.reasonName }}
      </button>
    </div>

    <button type="button" class="btn-soft full" @click="emit('login-operator-baru')">
      Login Operator Baru
    </button>

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