<script setup>
import { computed } from "vue";

const props = defineProps({
  socketStatus: {
    type: String,
    default: "offline",
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

const socketClass = computed(() => {
  const status = String(props.socketStatus || "").toLowerCase();

  if (
    status === "active" ||
    status === "connected" ||
    status === "open" ||
    status === "online" ||
    status === "aktif" ||
    status === "success" ||
    status === "ready" ||
    status === "1" ||
    status === "true"
  ) {
    return "active";
  }

  if (
    status === "connecting" ||
    status === "pending" ||
    status === "reconnecting" ||
    status === "loading"
  ) {
    return "connecting";
  }

  return "offline";
});

const socketTitle = computed(() => {
  if (socketClass.value === "active") return "WebSocket Aktif";
  if (socketClass.value === "connecting") return "WebSocket Connecting";
  return "WebSocket Offline";
});
</script>

<template>
  <div class="status-dots">
    <span
      class="dot socket-dot"
      :class="socketClass"
      :title="socketTitle"
      aria-label="WebSocket Status"
    ></span>

    <span
      v-if="isAdmin"
      class="dot admin-dot"
      title="Admin Mode"
      aria-label="Admin Mode"
    ></span>
  </div>
</template>

<style scoped>
.status-dots {
  position: fixed;
  top: 18px;
  left: 18px;
  z-index: 9999;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(219, 234, 254, 0.95);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(10px);
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 999px;
  display: inline-block;
}

.socket-dot.active {
  background: #22c55e;
  box-shadow:
    0 0 0 5px rgba(34, 197, 94, 0.18),
    0 0 14px rgba(34, 197, 94, 0.6);
}

.socket-dot.connecting {
  background: #f59e0b;
  box-shadow:
    0 0 0 5px rgba(245, 158, 11, 0.18),
    0 0 14px rgba(245, 158, 11, 0.6);
}

.socket-dot.offline {
  background: #ef4444;
  box-shadow:
    0 0 0 5px rgba(239, 68, 68, 0.18),
    0 0 14px rgba(239, 68, 68, 0.6);
}

.admin-dot {
  background: #facc15;
  box-shadow:
    0 0 0 5px rgba(250, 204, 21, 0.2),
    0 0 14px rgba(250, 204, 21, 0.65);
}

@media (max-width: 700px) {
  .status-dots {
    top: 12px;
    left: 12px;
    padding: 7px;
  }

  .dot {
    width: 10px;
    height: 10px;
  }
}
</style>