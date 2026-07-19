<script setup>
const props = defineProps({
  machines: {
    type: Array,
    default: () => [],
  },
  selectedDate: {
    type: String,
    default: "",
  },
});

function pct(value) {
  return Number(value || 0).toFixed(2);
}

function barWidth(value) {
  return `${Math.min(Math.max(Number(value || 0), 0), 100)}%`;
}
</script>

<template>
  <section class="panel ranking-panel">
    <div class="panel-head">
      <div>
        <h3>Rank Produktivitas Mesin</h3>
      </div>

      <span>{{ selectedDate }}</span>
    </div>

    <div class="rank-list">
      <div
        v-for="(m, index) in machines.slice(0, 10)"
        :key="m.uuid || index"
        class="rank-row"
      >
        <div class="rank-no" :class="{ top: index < 3 }">
          {{ index + 1 }}
        </div>

        <div class="rank-info">
          <strong :title="m.machineName">
            {{ m.machineName || "-" }}
          </strong>

          <small>
            {{ m.ip || "-" }} · {{ m.mainSource || "process_time" }}
          </small>
        </div>

        <div class="rank-bar-wrap">
          <div class="rank-bar">
            <div
              class="rank-bar-fill"
              :style="{ width: barWidth(m.productivity) }"
            ></div>
          </div>
        </div>

        <div class="rank-value">
          {{ pct(m.productivity) }}%
        </div>
      </div>

      <div v-if="!machines.length" class="empty">
        Belum ada data ranking mesin.
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel {
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  border: 1px solid #dbeafe;
  border-radius: 24px;
  padding: 20px;
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.05);
}

.ranking-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 18px;
}

.panel-head h3 {
  margin: 0 0 4px;
  font-size: 20px;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.panel-head p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.panel-head span {
  color: #64748b;
  font-size: 12px;
  font-weight: 900;
  white-space: nowrap;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  padding: 7px 10px;
}

.rank-list {
  display: grid;
  gap: 10px;
  flex: 1;
  align-content: start;
}

.rank-row {
  display: grid;
  grid-template-columns: 34px minmax(210px, 1fr) minmax(220px, 1.35fr) 78px;
  gap: 12px;
  align-items: center;
  padding: 10px 12px;
  border-radius: 16px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background 0.18s ease;
}

.rank-row:hover {
  transform: translateY(-1px);
  background: #f8fbff;
  border-color: #bfdbfe;
  box-shadow: 0 10px 20px rgba(37, 99, 235, 0.08);
}

.rank-no {
  width: 28px;
  height: 28px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid #bfdbfe;
  font-size: 13px;
  font-weight: 1000;
}

.rank-no.top {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  border-color: #2563eb;
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.22);
}

.rank-info {
  min-width: 0;
}

.rank-info strong {
  display: block;
  color: #0f172a;
  font-size: 13px;
  font-weight: 1000;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rank-info small {
  display: block;
  margin-top: 4px;
  color: #64748b;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rank-bar-wrap {
  min-width: 0;
}

.rank-bar {
  height: 10px;
  border-radius: 999px;
  background: #e5edf8;
  overflow: hidden;
}

.rank-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #ef4444, #f97316, #f59e0b);
}

.rank-value {
  text-align: right;
  color: #0f172a;
  font-size: 13px;
  font-weight: 1000;
  white-space: nowrap;
}

.empty {
  padding: 26px;
  text-align: center;
  color: #64748b;
  font-weight: 800;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 16px;
}

@media (max-width: 1100px) {
  .rank-row {
    grid-template-columns: 34px minmax(0, 1fr) 72px;
  }

  .rank-bar-wrap {
    grid-column: 2 / -1;
  }
}

@media (max-width: 640px) {
  .panel-head {
    flex-direction: column;
  }

  .rank-row {
    grid-template-columns: 30px minmax(0, 1fr);
  }

  .rank-value {
    grid-column: 2 / -1;
    text-align: left;
  }

  .rank-bar-wrap {
    grid-column: 2 / -1;
  }
}
</style>