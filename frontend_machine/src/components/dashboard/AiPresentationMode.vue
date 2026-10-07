<script setup>
import { computed, onBeforeUnmount, watch } from "vue";
import VueApexCharts from "vue3-apexcharts";

const props = defineProps({
  open: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  executive: { type: Object, default: null },
  errorMessage: { type: String, default: "" },
});

const emit = defineEmits([
  "close",
  "refresh",
  "briefing",
  "change-since",
  "export-briefing",
  "copy-briefing",
]);

const health = computed(() => props.executive?.health || null);
const compare = computed(() => props.executive?.compare || null);
const topActions = computed(() =>
  Array.isArray(props.executive?.topActions) ? props.executive.topActions : []
);

const scoreValue = computed(() => Number(health.value?.score ?? 0) || 0);

const scoreChartSeries = computed(() => [scoreValue.value]);

const scoreChartOptions = computed(() => {
  const tone = health.value?.tone || "warn";
  const color =
    tone === "good" ? "#1d4ed8" : tone === "critical" ? "#dc2626" : "#ea580c";
  return {
    chart: {
      type: "radialBar",
      sparkline: { enabled: false },
      toolbar: { show: false },
      animations: { enabled: true },
    },
    plotOptions: {
      radialBar: {
        startAngle: -120,
        endAngle: 120,
        hollow: { size: "58%" },
        track: { background: "#e2e8f0", strokeWidth: "100%" },
        dataLabels: {
          name: {
            show: true,
            offsetY: 28,
            color: "#64748b",
            fontSize: "13px",
            fontWeight: 800,
          },
          value: {
            show: true,
            offsetY: -10,
            fontSize: "42px",
            fontWeight: 900,
            color: "#0f172a",
            formatter: (val) => `${Math.round(val)}`,
          },
        },
      },
    },
    fill: { colors: [color] },
    labels: [health.value?.label || "Skor"],
    stroke: { lineCap: "round" },
  };
});

const compareChartCategories = computed(() => {
  const metrics = Array.isArray(compare.value?.metrics) ? compare.value.metrics : [];
  return metrics
    .filter((m) => m.key !== "high")
    .map((m) => m.label || m.key);
});

const compareChartSeries = computed(() => {
  const metrics = (Array.isArray(compare.value?.metrics) ? compare.value.metrics : []).filter(
    (m) => m.key !== "high"
  );
  return [
    {
      name: "Hari ini",
      data: metrics.map((m) => Number(m.today || 0)),
    },
    {
      name: "Kemarin",
      data: metrics.map((m) => Number(m.yesterday || 0)),
    },
  ];
});

const compareChartOptions = computed(() => ({
  chart: {
    type: "bar",
    toolbar: { show: false },
    animations: { enabled: true },
  },
  plotOptions: {
    bar: {
      horizontal: false,
      columnWidth: "48%",
      borderRadius: 6,
    },
  },
  colors: ["#1d4ed8", "#94a3b8"],
  dataLabels: { enabled: false },
  stroke: { show: true, width: 2, colors: ["transparent"] },
  xaxis: {
    categories: compareChartCategories.value,
    labels: {
      style: { fontWeight: 700, colors: "#64748b" },
    },
  },
  yaxis: {
    labels: {
      style: { fontWeight: 700, colors: "#64748b" },
    },
  },
  legend: {
    position: "top",
    fontWeight: 800,
  },
  grid: {
    borderColor: "#e2e8f0",
    strokeDashArray: 4,
  },
  tooltip: {
    y: {
      formatter: (val, opts) => {
        const idx = opts?.dataPointIndex ?? 0;
        const metrics = (compare.value?.metrics || []).filter((m) => m.key !== "high");
        const unit = metrics[idx]?.unit || "";
        return `${val}${unit}`;
      },
    },
  },
}));

function onKeydown(event) {
  if (event.key === "Escape" && props.open) {
    emit("close");
  }
}

watch(
  () => props.open,
  (isOpen) => {
    document.body.style.overflow = isOpen ? "hidden" : "";
    if (isOpen) {
      window.addEventListener("keydown", onKeydown);
      emit("refresh");
    } else {
      window.removeEventListener("keydown", onKeydown);
    }
  }
);

onBeforeUnmount(() => {
  document.body.style.overflow = "";
  window.removeEventListener("keydown", onKeydown);
});

function deltaText(metric) {
  const d = Number(metric?.delta || 0);
  const unit = metric?.unit || "";
  const sign = d > 0 ? "+" : "";
  return `${sign}${d}${unit}`;
}
</script>

<template>
  <div v-if="open" class="pres" role="dialog" aria-modal="true">
    <div class="pres__stage">
      <header class="pres__head">
        <div>
          <p class="pres__eyebrow">Slide 01 · Mode Presentasi</p>
          <h2>Skor Utilitas Mesin</h2>
          <p class="pres__sub">
            {{ executive?.workDate || "-" }} · update
            {{ executive?.generatedAt || "-" }}
          </p>
        </div>
        <div class="pres__actions">
          <button type="button" class="btn soft" :disabled="loading" @click="emit('refresh')">
            {{ loading ? "Memuat..." : "Refresh" }}
          </button>
          <button type="button" class="btn soft" @click="emit('change-since')">
            Berubah sejak tadi?
          </button>
          <button type="button" class="btn soft" @click="emit('copy-briefing')">
            Salin Briefing
          </button>
          <button type="button" class="btn soft" @click="emit('export-briefing')">
            Export Briefing
          </button>
          <button type="button" class="btn solid" @click="emit('close')">Tutup</button>
        </div>
      </header>

      <p v-if="errorMessage" class="pres__error">{{ errorMessage }}</p>

      <section class="pres__slide slide-hero">
        <div class="score" :class="health?.tone || 'warn'">
          <span class="score__label">Skor Utilitas Mesin</span>
          <strong>{{ health?.score ?? "-" }}</strong>
          <em>{{ health?.label || "—" }}</em>
          <p>{{ health?.summary || "Menunggu data..." }}</p>
        </div>

        <div class="stats">
          <article>
            <span>Util rata-rata</span>
            <strong>{{ Number(health?.avgUtilPct || 0).toFixed(1) }}%</strong>
          </article>
          <article>
            <span>Mesin aktif</span>
            <strong>{{ health?.activeCount ?? 0 }}</strong>
          </article>
          <article>
            <span>Status BAD</span>
            <strong>{{ health?.badCount ?? 0 }}</strong>
          </article>
          <article>
            <span>Prioritas segera</span>
            <strong>{{ health?.highCount ?? 0 }}</strong>
          </article>
        </div>
      </section>

      <section class="pres__slide">
        <p class="pres__slide-label">Slide 02 · Diagram</p>
        <div class="pres__charts">
          <article class="card chart-card">
            <h3>Diagram Skor Utilitas</h3>
            <p class="card__summary">{{ health?.summary || "Menunggu data skor..." }}</p>
            <div class="chart-wrap score-chart">
              <VueApexCharts
                type="radialBar"
                height="280"
                :options="scoreChartOptions"
                :series="scoreChartSeries"
              />
            </div>
          </article>

          <article class="card chart-card">
            <h3>Diagram Vs Kemarin</h3>
            <p class="card__summary">{{ compare?.summary || "Belum ada pembandingan." }}</p>
            <div class="chart-wrap">
              <VueApexCharts
                type="bar"
                height="280"
                :options="compareChartOptions"
                :series="compareChartSeries"
              />
            </div>
          </article>
        </div>
      </section>

      <section class="pres__slide">
        <p class="pres__slide-label">Slide 03 · Aksi & Ringkas</p>
        <div class="pres__grid">
          <article class="card">
            <h3>Vs Kemarin</h3>
            <p class="card__summary">{{ compare?.summary || "Belum ada pembandingan." }}</p>
            <div class="metrics">
              <div
                v-for="metric in compare?.metrics || []"
                :key="metric.key"
                class="metric"
                :class="metric.better"
              >
                <span>{{ metric.label }}</span>
                <strong>
                  {{ Number(metric.today || 0).toFixed(metric.unit === "%" ? 1 : 0)
                  }}{{ metric.unit || "" }}
                </strong>
                <em>{{ deltaText(metric) }} vs kemarin</em>
              </div>
            </div>
          </article>

          <article class="card">
            <h3>Prioritas 1 Jam ke Depan</h3>
            <ol v-if="topActions.length">
              <li v-for="(action, idx) in topActions" :key="`${idx}-${action}`">
                {{ action }}
              </li>
            </ol>
            <p v-else class="card__summary">Tidak ada prioritas mendesak saat ini.</p>
          </article>
        </div>
      </section>

      <section class="pres__slide card briefing">
        <div class="briefing__head">
          <div>
            <p class="pres__slide-label">Slide 04 · Briefing</p>
            <h3>Briefing Atasan</h3>
          </div>
          <div class="briefing__actions">
            <button type="button" class="btn soft" @click="emit('copy-briefing')">Salin</button>
            <button type="button" class="btn soft" @click="emit('export-briefing')">Export</button>
            <button type="button" class="btn soft" @click="emit('briefing')">Ke Chat AI</button>
          </div>
        </div>
        <pre>{{ executive?.briefing || "Klik Refresh untuk menyusun briefing." }}</pre>
      </section>
    </div>
  </div>
</template>

<style scoped>
.pres {
  position: fixed;
  inset: 0;
  z-index: 13000;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.16), transparent 40%),
    linear-gradient(180deg, #f8fbff 0%, #ffffff 45%, #f1f5f9 100%);
  padding: clamp(16px, 2.4vw, 28px);
  overflow: auto;
  box-sizing: border-box;
  color: #0f172a;
}

.pres__stage {
  width: min(1280px, 100%);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.pres__head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.pres__eyebrow,
.pres__slide-label {
  margin: 0 0 6px;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #1d4ed8;
}

.pres__head h2 {
  margin: 0;
  font-size: clamp(28px, 3.6vw, 44px);
  color: #0f172a;
  line-height: 1.12;
}

.pres__sub {
  margin: 6px 0 0;
  color: #64748b;
  font-weight: 700;
  font-size: 13px;
}

.pres__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.btn {
  border: 0;
  border-radius: 12px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
}

.btn.soft {
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}

.btn.solid {
  background: #1d4ed8;
  color: #fff;
}

.pres__error {
  margin: 0;
  padding: 12px 14px;
  border-radius: 12px;
  background: #fef2f2;
  color: #b91c1c;
  font-weight: 700;
}

.pres__slide {
  border: 1px solid #bfdbfe;
  border-radius: 24px;
  padding: clamp(16px, 2vw, 24px);
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 42%),
    linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 12px 28px rgba(37, 99, 235, 0.08);
}

.slide-hero {
  display: grid;
  grid-template-columns: minmax(240px, 320px) 1fr;
  gap: 16px;
  align-items: stretch;
}

.pres__charts {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 16px;
}

.chart-card .chart-wrap {
  min-height: 280px;
}

.score-chart {
  display: flex;
  align-items: center;
  justify-content: center;
}

.score {
  border-radius: 22px;
  padding: 22px;
  color: #1e3a8a;
  background: linear-gradient(160deg, #eff6ff, #dbeafe);
  border: 1px solid #93c5fd;
  box-shadow: 0 16px 40px rgba(37, 99, 235, 0.12);
}

.score.warn {
  color: #9a3412;
  background: linear-gradient(160deg, #fff7ed, #ffedd5);
  border-color: #fdba74;
}

.score.critical {
  color: #991b1b;
  background: linear-gradient(160deg, #fef2f2, #fee2e2);
  border-color: #fca5a5;
}

.score__label {
  display: block;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.score strong {
  display: block;
  margin-top: 8px;
  font-size: clamp(56px, 8vw, 84px);
  line-height: 0.95;
}

.score em {
  display: inline-block;
  margin-top: 8px;
  font-style: normal;
  font-weight: 900;
  font-size: 18px;
}

.score p {
  margin: 12px 0 0;
  font-size: 14px;
  line-height: 1.45;
  font-weight: 700;
}

.stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.stats article {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 18px;
  padding: 16px 18px;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.04);
}

.stats span {
  display: block;
  font-size: 12px;
  font-weight: 800;
  color: #64748b;
}

.stats strong {
  display: block;
  margin-top: 8px;
  font-size: clamp(24px, 3vw, 34px);
  color: #0f172a;
}

.pres__grid {
  display: grid;
  grid-template-columns: 1.1fr 1fr;
  gap: 16px;
}

.card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  padding: 18px 20px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.card h3 {
  margin: 0 0 10px;
  font-size: 18px;
  color: #0f172a;
}

.card__summary {
  margin: 0 0 14px;
  color: #334155;
  font-weight: 700;
  line-height: 1.5;
}

.metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.metric {
  border-radius: 14px;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.metric.up {
  background: #eff6ff;
  border-color: #93c5fd;
}

.metric.down {
  background: #fff7ed;
  border-color: #fed7aa;
}

.metric span {
  display: block;
  font-size: 11px;
  font-weight: 800;
  color: #64748b;
}

.metric strong {
  display: block;
  margin-top: 4px;
  font-size: 22px;
  color: #0f172a;
}

.metric em {
  display: block;
  margin-top: 4px;
  font-style: normal;
  font-size: 12px;
  font-weight: 800;
  color: #1d4ed8;
}

.card ol {
  margin: 0;
  padding-left: 18px;
  display: grid;
  gap: 10px;
}

.card li {
  color: #0f172a;
  font-weight: 700;
  line-height: 1.45;
  font-size: 14px;
}

.briefing {
  margin: 0;
}

.briefing pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  font-size: 14px;
  line-height: 1.55;
  color: #1e3a8a;
  font-weight: 650;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 14px;
  padding: 14px 16px;
}

.briefing__head {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: flex-start;
  margin-bottom: 10px;
}

.briefing__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 960px) {
  .slide-hero,
  .pres__grid,
  .pres__charts,
  .stats,
  .metrics {
    grid-template-columns: 1fr;
  }

  .pres__head {
    flex-direction: column;
  }
}
</style>
