<script setup>
import { computed } from "vue";

const props = defineProps({
  message: {
    type: String,
    default: "",
  },
  bestMachine: {
    type: Object,
    default: null,
  },
  worstMachine: {
    type: Object,
    default: null,
  },
});

function pct(v) {
  return `${Number(v || 0).toFixed(2)}%`;
}

const bestName = computed(() => props.bestMachine?.machineName || "-");
const worstName = computed(() => props.worstMachine?.machineName || "-");
</script>

<template>
  <section class="summary-grid">
    <div class="summary-main">
      <div class="summary-tag">Kesimpulan Hari Ini</div>
      <h3>Ringkasan performa produksi mesin</h3>
      <p>{{ message }}</p>
    </div>

    <div class="summary-mini">
      <div class="mini-tag">Mesin Terbaik</div>
      <strong>{{ bestName }}</strong>
      <span>{{ pct(bestMachine?.productivity) }}</span>
    </div>

    <div class="summary-mini danger">
      <div class="mini-tag">Terendah</div>
      <strong>{{ worstName }}</strong>
      <span>{{ pct(worstMachine?.productivity) }}</span>
    </div>
  </section>
</template>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: 1.3fr 0.55fr 0.55fr;
  gap: 18px;
}

.summary-main,
.summary-mini {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid #dbeafe;
  border-radius: 24px;
  padding: 20px 22px;
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.05);
}

.summary-main {
  position: relative;
  overflow: hidden;
}

.summary-main::after {
  content: "";
  position: absolute;
  right: -40px;
  top: -40px;
  width: 140px;
  height: 140px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.08);
}

.summary-tag,
.mini-tag {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  padding: 7px 12px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.summary-main h3 {
  margin: 0 0 10px;
  font-size: 22px;
  line-height: 1.2;
  color: #0f172a;
}

.summary-main p {
  margin: 0;
  font-size: 16px;
  line-height: 1.65;
  color: #334155;
  font-weight: 700;
  max-width: 95%;
}

.summary-mini {
  display: grid;
  align-content: start;
  gap: 8px;
}

.summary-mini strong {
  font-size: 19px;
  line-height: 1.35;
  color: #0f172a;
}

.summary-mini span {
  font-size: 26px;
  font-weight: 900;
  color: #2563eb;
}

.summary-mini.danger span {
  color: #ef4444;
}

@media (max-width: 1200px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>