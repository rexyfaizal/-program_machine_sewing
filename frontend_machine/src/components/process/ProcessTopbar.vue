<script setup>
import { computed } from "vue";
import {
  getMachineStatusText,
  getMachineStatusClass,
} from "../../utils/machineStatus";

const props = defineProps({
  machine: {
    type: Object,
    default: () => ({}),
  },
  selectedDate: {
    type: String,
    required: true,
  },
});

const emit = defineEmits([
  "update:selectedDate",
  "last-day",
  "next-day",
  "today",
  "date-change",
]);

const displayName = computed(() => {
  return (
    props.machine?.machineName ||
    props.machine?.nickName ||
    props.machine?.originalMachineName ||
    props.machine?.uuid ||
    "Pilih Mesin"
  );
});

const displayLocation = computed(() => {
  return props.machine?.location || "-";
});

const displayIp = computed(() => {
  return props.machine?.ip || "-";
});

const displayUuid = computed(() => {
  return props.machine?.uuid || "-";
});

const displayStatusText = computed(() => {
  return getMachineStatusText(props.machine);
});

const displayStatusClass = computed(() => {
  return getMachineStatusClass(props.machine);
});

function updateDate(event) {
  emit("update:selectedDate", event.target.value);
}

function changeDate() {
  emit("date-change");
}
</script>

<template>
  <section class="process-topbar-card">
    <div class="topbar-main">
      <div class="machine-title-area">
        <span class="section-label">Detail Mesin</span>
        <h1>{{ displayName }}</h1>
      </div>

      <div class="date-actions">
        <button type="button" @click="emit('last-day')">
          Last Day
        </button>

        <input
          type="date"
          :value="selectedDate"
          @input="updateDate"
          @change="changeDate"
        />

        <button type="button" @click="emit('next-day')">
          Next Day
        </button>

        <button type="button" @click="emit('today')">
          Today
        </button>
      </div>
    </div>

    <div class="machine-info-table">
      <div class="info-cell">
        <span>Location</span>
        <strong>{{ displayLocation }}</strong>
      </div>

      <div class="info-cell">
        <span>IP Address</span>
        <strong>{{ displayIp }}</strong>
      </div>

      <div class="info-cell uuid-cell">
        <span>UUID</span>
        <strong>{{ displayUuid }}</strong>
      </div>

      <div class="info-cell status-cell">
        <span>Status Mesin</span>

        <strong class="status-badge" :class="displayStatusClass">
          <i></i>
          {{ displayStatusText }}
        </strong>
      </div>
    </div>
  </section>
</template>

<style scoped>
.process-topbar-card {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(248, 251, 255, 0.96)),
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.12), transparent 34%);
  border: 1px solid #dbe4ef;
  border-radius: 22px;
  padding: 22px;
  margin-bottom: 18px;
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.07);
}

.topbar-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 18px;
}

.machine-title-area {
  min-width: 0;
}

.section-label {
  display: inline-flex;
  width: fit-content;
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 10px;
}

.machine-title-area h1 {
  margin: 0;
  color: #0f172a;
  font-size: clamp(28px, 3vw, 42px);
  line-height: 1.05;
  font-weight: 1000;
  letter-spacing: -0.045em;
}

.date-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
  flex: 0 0 auto;
}

.date-actions button {
  border: 0;
  background: #2563eb;
  color: #ffffff;
  border-radius: 12px;
  padding: 11px 14px;
  font-size: 12px;
  font-weight: 1000;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.22);
}

.date-actions button:hover {
  background: #1d4ed8;
}

.date-actions button:hover {
  background: #2563eb;
}

.date-actions input {
  height: 42px;
  border: 1px solid #d8e2ee;
  border-radius: 12px;
  padding: 0 14px;
  background: #ffffff;
  color: #0f172a;
  font-size: 13px;
  font-weight: 900;
  outline: none;
}

.machine-info-table {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1.5fr 1fr;
  gap: 10px;
}

.info-cell {
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
  padding: 13px 14px;
  min-height: 74px;
  display: grid;
  align-content: center;
  gap: 6px;
}

.info-cell span {
  color: #64748b;
  font-size: 11px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.info-cell strong {
  color: #0f172a;
  font-size: 14px;
  font-weight: 1000;
  word-break: break-word;
}

.uuid-cell strong {
  font-family: Consolas, Menlo, monospace;
  font-size: 13px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: fit-content;
  gap: 8px;
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 13px !important;
  font-weight: 1000;
  border: 1px solid transparent;
}

.status-badge i {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  display: inline-block;
}

/* MacState 2 = Working = Hijau */
.status-badge.working {
  background: #dcfce7;
  color: #166534;
  border-color: #bbf7d0;
}

.status-badge.working i {
  background: #22c55e;
}

/* MacState 1 = Online = Biru */
.status-badge.online {
  background: #dbeafe;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

.status-badge.online i {
  background: #2563eb;
}

/* MacState 0 = Offline = Abu-abu */
.status-badge.offline {
  background: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}

.status-badge.offline i {
  background: #94a3b8;
}

@media (max-width: 1200px) {
  .topbar-main {
    flex-direction: column;
  }

  .date-actions {
    justify-content: flex-start;
  }

  .machine-info-table {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 650px) {
  .process-topbar-card {
    padding: 16px;
  }

  .machine-info-table {
    grid-template-columns: 1fr;
  }

  .date-actions {
    width: 100%;
  }

  .date-actions input,
  .date-actions button {
    flex: 1;
  }
}
</style>