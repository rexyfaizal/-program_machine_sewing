<script setup>
defineProps({
  activeSession: {
    type: Object,
    default: null,
  },
  formatDateTime: {
    type: Function,
    required: true,
  },
});

const emit = defineEmits(["login-operator-baru", "batal"]);
</script>

<template>
  <div v-if="activeSession" class="conflict-box">
    <p>Mesin ini sedang login oleh:</p>

    <h2>{{ activeSession.operatorNik }} - {{ activeSession.operatorName }}</h2>

    <small>
      Login: {{ formatDateTime(activeSession.loginTime) }}
    </small>

    <div class="conflict-actions">
      <button type="button" class="btn-danger" @click="emit('login-operator-baru')">
        Login Operator Baru
      </button>

      <button type="button" class="btn-soft" @click="emit('batal')">
        Batal
      </button>
    </div>
  </div>
</template>