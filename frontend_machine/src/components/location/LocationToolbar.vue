<script setup>
const props = defineProps({
  selectedFactory: {
    type: String,
    required: true,
  },
  selectedDate: {
    type: String,
    required: true,
  },
  factoryOptions: {
    type: Array,
    default: () => [],
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits([
  "update:selectedFactory",
  "update:selectedDate",
  "factoryChange",
  "refresh",
  "openShiftConfig",
]);

function selectFactory(factory) {
  emit("update:selectedFactory", factory.key);
  emit("factoryChange", factory.key);
}

function updateDate(event) {
  emit("update:selectedDate", event.target.value);
}
</script>

<template>
  <section class="location-toolbar-wrap">
    <div class="location-toolbar">
      <div class="toolbar-left">
        <div class="toolbar-label">Factory View</div>

        <div class="factory-tabs">
          <button
            v-for="factory in factoryOptions"
            :key="factory.key"
            type="button"
            class="factory-tab"
            :class="{ active: selectedFactory === factory.key }"
            @click="selectFactory(factory)"
          >
            {{ factory.label }}
          </button>
        </div>
      </div>

      <div class="toolbar-right">
        <label class="date-box">
          <input
            type="date"
            :value="selectedDate"
            @input="updateDate"
          />
        </label>

        <button
          v-if="isAdmin"
          type="button"
          class="shift-btn"
          @click="emit('openShiftConfig')"
        >
          Atur Shift
        </button>

        <button type="button" class="refresh-btn" @click="emit('refresh')">
          Refresh
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.location-toolbar-wrap {
  margin-bottom: 14px;
}

.location-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.98), rgba(239, 246, 255, 0.92));
  border: 1px solid #dbeafe;
  box-shadow:
    0 10px 20px rgba(37, 99, 235, 0.07),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.toolbar-left {
  flex: 1;
}

.toolbar-label {
  padding: 8px 12px;
  border-radius: 12px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  border: 1px solid #bfdbfe;
  white-space: nowrap;
}

.factory-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.factory-tab {
  min-width: 76px;
  height: 38px;
  padding: 0 14px;
  border-radius: 12px;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  transition:
    all 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.factory-tab:hover {
  background: #eff6ff;
  border-color: #93c5fd;
  color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.12);
}

.factory-tab.active {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border-color: #2563eb;
  color: #ffffff;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.date-box {
  min-width: 160px;
  height: 42px;
  display: grid;
  align-content: center;
  padding: 0 12px;
  border-radius: 12px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.04);
}

.date-box input {
  border: 0;
  outline: none;
  background: transparent;
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  padding: 0;
}

.refresh-btn {
  height: 42px;
  min-width: 98px;
  padding: 0 16px;
  border: 1px solid #2563eb;
  border-radius: 12px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.shift-btn {
  height: 42px;
  min-width: 110px;
  padding: 0 16px;
  border: 1px solid #0f766e;
  border-radius: 12px;
  background: linear-gradient(135deg, #14b8a6, #0f766e);
  color: #ffffff;
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(15, 118, 110, 0.22);
}

.shift-btn:hover {
  background: linear-gradient(135deg, #0f766e, #115e59);
}

.refresh-btn:hover {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

@media (max-width: 980px) {
  .location-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-right {
    justify-content: flex-start;
  }
}

@media (max-width: 640px) {
  .factory-tab {
    min-width: 70px;
    height: 38px;
    font-size: 13px;
  }

  .date-box {
    min-width: 100%;
  }

  .refresh-btn {
    width: 100%;
  }
}
</style>