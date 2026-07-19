<script setup>
const props = defineProps({
  summary: {
    type: Object,
    required: true,
  },
  donutStyle: {
    type: Object,
    required: true,
  },
});

function hoursFromSec(sec) {
  return `${(Number(sec || 0) / 3600).toFixed(2)}h`;
}
</script>

<template>
  <section class="panel">
    <div class="panel-head">
      <div>
        <h3>Status</h3>
      </div>
    </div>

    <div class="distribution">
      <div class="donut-box">
        <div class="donut" :style="donutStyle">
          <div class="donut-center">
            <strong>{{ summary.totalMachine }}</strong>
          </div>
        </div>
      </div>

      <div class="stats">
        <div class="stat-row">
          <span class="badge good">Good</span>
          <strong>{{ summary.good }}</strong>
        </div>

        <div class="stat-row">
          <span class="badge normal">Normal</span>
          <strong>{{ summary.normal }}</strong>
        </div>

        <div class="stat-row">
          <span class="badge bad">Bad</span>
          <strong>{{ summary.bad }}</strong>
        </div>

        <div class="divider"></div>

        <div class="meta-row">
          <span>Total Alarm</span>
          <strong>{{ summary.totalAlarm }}</strong>
        </div>

        <div class="meta-row">
          <span>Total Sewing Time</span>
          <strong>{{ hoursFromSec(summary.totalSewingTime) }}</strong>
        </div>
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

.panel-head {
  margin-bottom: 16px;
}

.panel-head h3 {
  margin: 0 0 4px;
  font-size: 20px;
  color: #0f172a;
}

.panel-head p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.distribution {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 18px;
  align-items: center;
}

.donut-box {
  display: flex;
  justify-content: center;
}

.donut {
  width: 150px;
  height: 150px;
  border-radius: 999px;
  display: grid;
  place-items: center;
}

.donut-center {
  width: 92px;
  height: 92px;
  border-radius: 999px;
  background: #ffffff;
  display: grid;
  place-items: center;
  text-align: center;
  box-shadow: inset 0 0 0 1px #e2e8f0;
}

.donut-center strong {
  font-size: 34px;
  line-height: 1;
  color: #0f172a;
}

.donut-center span {
  font-size: 12px;
  color: #64748b;
  font-weight: 800;
  text-transform: uppercase;
}

.stats {
  display: grid;
  gap: 12px;
}

.stat-row,
.meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.badge {
  display: inline-flex;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 900;
  text-transform: uppercase;
}

.badge.good {
  background: #dcfce7;
  color: #166534;
}

.badge.normal {
  background: #fef3c7;
  color: #92400e;
}

.badge.bad {
  background: #fee2e2;
  color: #991b1b;
}

.stat-row strong,
.meta-row strong {
  color: #0f172a;
  font-size: 15px;
}

.meta-row span {
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.divider {
  height: 1px;
  background: #e2e8f0;
}

@media (max-width: 640px) {
  .distribution {
    grid-template-columns: 1fr;
  }
}
</style>