<script setup>
const props = defineProps({
  summary: {
    type: Object,
    required: true,
    default: () => ({
      totalMachine: 0,
      avgProductivity: 0,
      good: 0,
      normal: 0,
      bad: 0,
      totalOutput: 0,
    }),
  },
});

const cards = [
  {
    key: "totalMachine",
    title: "Total Mesin",
    caption: "Mesin dari machineinfo",
    suffix: "",
  },
  {
    key: "avgProductivity",
    title: "Avg Produktivitas",
    caption: "Rata-rata seluruh mesin",
    suffix: "%",
  },
  {
    key: "good",
    title: "Good",
    caption: "≥ 90%",
    suffix: "",
  },
  {
    key: "normal",
    title: "Normal",
    caption: "80% sampai < 90%",
    suffix: "",
  },
  {
    key: "bad",
    title: "Bad",
    caption: "< 80%",
    suffix: "",
  },
  {
    key: "totalOutput",
    title: "Total Output",
    caption: "Sum ProcCounts",
    suffix: "",
  },
];

function formatValue(key, value, suffix = "") {
  if (key === "avgProductivity") {
    return `${Number(value || 0).toFixed(2)}${suffix}`;
  }

  return `${Number(value || 0)}${suffix}`;
}
</script>

<template>
  <section class="kpi-grid">
    <article
      v-for="card in cards"
      :key="card.key"
      class="kpi-card"
    >
      <div class="kpi-dot"></div>

      <h3>{{ card.title }}</h3>
      <strong>
        {{ formatValue(card.key, props.summary?.[card.key], card.suffix) }}
      </strong>
      <p>{{ card.caption }}</p>
    </article>
  </section>
</template>

<style scoped>
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 18px;
}

.kpi-card {
  position: relative;
  overflow: hidden;
  border-radius: 22px;
  padding: 16px 18px;
  min-height: 126px;
  border: 1px solid #2f6df6;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 55%, #1d4ed8 100%);
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.20);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.kpi-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 18px 32px rgba(37, 99, 235, 0.28);
}

.kpi-card::after {
  content: "";
  position: absolute;
  top: -14px;
  right: -14px;
  width: 60px;
  height: 60px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.18);
}

.kpi-dot {
  position: absolute;
  top: 18px;
  right: 18px;
  width: 9px;
  height: 9px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.85);
}

.kpi-card h3 {
  margin: 0 0 18px;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.9);
}

.kpi-card strong {
  display: block;
  margin-bottom: 8px;
  font-size: 26px;
  line-height: 1;
  font-weight: 900;
  color: #ffffff;
}

.kpi-card p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.88);
}

@media (max-width: 1280px) {
  .kpi-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 520px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>