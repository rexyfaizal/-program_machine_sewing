<script setup>
const props = defineProps({
  socketStatus: {
    type: String,
    default: "offline",
  },
  lastUpdate: {
    type: String,
    default: "-",
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["refresh"]);
</script>

<template>
  <section class="connection-bar">
    <div class="connection-spacer"></div>

    <div class="connection-right">
      <span class="last-update">
        Last update:
        <strong>{{ lastUpdate }}</strong>
      </span>

      <button type="button" class="refresh-btn" @click="emit('refresh')">
        Refresh
      </button>
    </div>
  </section>
</template>

<style scoped>
.connection-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  border-radius: 20px;
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  border: 1px solid #dbeafe;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.05);
}

.connection-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.last-update {
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.last-update strong {
  color: #0f172a;
  font-weight: 950;
}

.refresh-btn {
  border: 0;
  border-radius: 14px;
  padding: 11px 16px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.2);
}

.refresh-btn:hover {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

@media (max-width: 700px) {
  .connection-bar {
    justify-content: flex-start;
  }

  .connection-right {
    width: 100%;
    justify-content: space-between;
  }
}
</style>