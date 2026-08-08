<script setup>
import { computed, ref, watch } from "vue";
import VueApexCharts from "vue3-apexcharts";

import { getProductivity } from "../../api/machineApi";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  machine: {
    type: Object,
    default: null,
  },
  startDate: {
    type: String,
    default: "",
  },
  endDate: {
    type: String,
    default: "",
  },
  shift: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["close"]);

const loading = ref(false);
const errorMessage = ref("");
const categories = ref([]);
const productivityData = ref([]);

const machineName = computed(() => {
  const m = props.machine || {};
  return (
    m.displayMachineName ||
    m.machineName ||
    m.nickName ||
    m.customName ||
    m.uuid ||
    "-"
  );
});

const rangeLabel = computed(() => {
  const start = props.startDate || props.endDate;
  const end = props.endDate || props.startDate;
  if (!start) return "-";
  if (!end || end === start) return formatDateLabel(start);
  return `${formatDateLabel(start)} - ${formatDateLabel(end)}`;
});

const validValues = computed(() =>
  productivityData.value.filter((v) => v !== null && !Number.isNaN(v))
);

const averageProductivity = computed(() => {
  if (!validValues.value.length) return 0;
  const total = validValues.value.reduce((sum, v) => sum + v, 0);
  return Number((total / validValues.value.length).toFixed(2));
});

const bestDay = computed(() => pickExtreme("max"));
const worstDay = computed(() => pickExtreme("min"));

const hasData = computed(() => validValues.value.length > 0);

const chartSeries = computed(() => [
  {
    name: "Produktivitas (%)",
    data: productivityData.value,
  },
]);

const chartOptions = computed(() => ({
  chart: {
    type: "line",
    height: 360,
    toolbar: { show: true, tools: { download: true } },
    fontFamily: "inherit",
    animations: { enabled: true },
  },
  colors: ["#2563eb"],
  stroke: { curve: "smooth", width: 3 },
  markers: { size: 5, hover: { size: 7 } },
  dataLabels: {
    enabled: true,
    formatter: (val) => (val === null ? "" : `${Number(val).toFixed(0)}%`),
    style: { fontSize: "11px", fontWeight: 700 },
    background: { enabled: true, foreColor: "#0f172a", borderWidth: 0 },
  },
  xaxis: {
    categories: categories.value,
    labels: { style: { fontWeight: 700 } },
    title: { text: "Tanggal" },
  },
  yaxis: {
    min: 0,
    max: 100,
    tickAmount: 5,
    labels: { formatter: (val) => `${Number(val).toFixed(0)}%` },
    title: { text: "Produktivitas (%)" },
  },
  annotations: {
    yaxis: averageProductivity.value
      ? [
          {
            y: averageProductivity.value,
            borderColor: "#f59e0b",
            strokeDashArray: 5,
            label: {
              text: `Rata-rata ${averageProductivity.value}%`,
              style: { background: "#f59e0b", color: "#ffffff" },
            },
          },
        ]
      : [],
  },
  tooltip: {
    y: {
      formatter: (val) => (val === null ? "Tidak ada data" : `${Number(val).toFixed(2)}%`),
    },
  },
  grid: { borderColor: "#e2e8f0" },
  noData: { text: "Tidak ada data pada rentang ini." },
}));

function pickExtreme(kind) {
  let idx = -1;
  let value = kind === "max" ? -Infinity : Infinity;
  productivityData.value.forEach((v, i) => {
    if (v === null || Number.isNaN(v)) return;
    if ((kind === "max" && v > value) || (kind === "min" && v < value)) {
      value = v;
      idx = i;
    }
  });
  if (idx === -1) return null;
  return { label: categories.value[idx], value: Number(value.toFixed(2)) };
}

function toIsoDate(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function buildDateRange(startText, endText) {
  let start = new Date(`${startText}T00:00:00`);
  let end = new Date(`${endText}T00:00:00`);
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return [];
  if (end < start) {
    const tmp = start;
    start = end;
    end = tmp;
  }
  const dates = [];
  const cursor = new Date(start);
  while (cursor <= end) {
    dates.push(toIsoDate(cursor));
    cursor.setDate(cursor.getDate() + 1);
  }
  return dates;
}

function formatDateLabel(dateText) {
  const d = new Date(`${dateText}T00:00:00`);
  if (Number.isNaN(d.getTime())) return dateText;
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short" });
}

function extractRows(data) {
  if (Array.isArray(data)) return data;
  return (
    data?.rows ||
    data?.Rows ||
    data?.data ||
    data?.Data ||
    data?.items ||
    data?.Items ||
    data?.machines ||
    data?.Machines ||
    []
  );
}

function readProductivity(row) {
  const keys = [
    "productivityPct",
    "ProductivityPct",
    "productivity",
    "Productivity",
    "productivityRaw",
    "ProductivityRaw",
  ];
  for (const key of keys) {
    if (row[key] !== undefined && row[key] !== null) {
      return Number(row[key]);
    }
  }
  return null;
}

async function loadChart() {
  const uuid = String(props.machine?.uuid || "").trim().toLowerCase();
  if (!uuid) {
    errorMessage.value = "Mesin tidak valid.";
    return;
  }

  const start = props.startDate || props.endDate;
  const end = props.endDate || props.startDate;
  const dates = buildDateRange(start, end);
  if (!dates.length) {
    errorMessage.value = "Rentang tanggal tidak valid.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  const shift = String(props.shift || "").trim();

  try {
    const results = await Promise.all(
      dates.map(async (date) => {
        try {
          const payload = await getProductivity(date, { shift });
          const rows = extractRows(payload);
          const match = rows.find(
            (r) =>
              String(r.uuid || r.UUID || "").trim().toLowerCase() === uuid
          );
          return match ? readProductivity(match) : null;
        } catch (err) {
          return null;
        }
      })
    );

    categories.value = dates.map(formatDateLabel);
    productivityData.value = results.map((v) =>
      v === null || Number.isNaN(v) ? null : Number(Number(v).toFixed(2))
    );
  } catch (err) {
    errorMessage.value = `Gagal memuat grafik: ${err.message}`;
  } finally {
    loading.value = false;
  }
}

function close() {
  emit("close");
}

watch(
  () => [props.show, props.machine?.uuid, props.startDate, props.endDate, props.shift],
  () => {
    if (props.show && props.machine?.uuid) {
      loadChart();
    }
  },
  { immediate: true }
);
</script>

<template>
  <div v-if="props.show" class="chart-overlay" @click.self="close">
    <div class="chart-modal">
      <div class="chart-head">
        <div>
          <h3>Grafik Produktivitas</h3>
          <p>
            <strong>{{ machineName }}</strong>
            <span v-if="props.machine?.location"> · {{ props.machine.location }}</span>
            <span class="range"> · {{ rangeLabel }}</span>
          </p>
        </div>

        <button type="button" class="btn-close" @click="close">✕</button>
      </div>

      <div v-if="loading" class="chart-state">Mengambil data grafik...</div>

      <div v-else-if="errorMessage" class="chart-state error">
        {{ errorMessage }}
      </div>

      <template v-else>
        <div class="chart-stats">
          <div class="stat">
            <span>Rata-rata</span>
            <strong>{{ averageProductivity }}%</strong>
          </div>
          <div class="stat good">
            <span>Tertinggi</span>
            <strong v-if="bestDay">{{ bestDay.value }}% ({{ bestDay.label }})</strong>
            <strong v-else>-</strong>
          </div>
          <div class="stat bad">
            <span>Terendah</span>
            <strong v-if="worstDay">{{ worstDay.value }}% ({{ worstDay.label }})</strong>
            <strong v-else>-</strong>
          </div>
        </div>

        <div class="chart-body">
          <VueApexCharts
            v-if="hasData"
            type="line"
            height="360"
            :options="chartOptions"
            :series="chartSeries"
          />
          <div v-else class="chart-state">
            Tidak ada data produktivitas pada rentang ini.
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.chart-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.chart-modal {
  background: #ffffff;
  border-radius: 22px;
  width: min(920px, 100%);
  max-height: 90vh;
  overflow: auto;
  padding: 22px 24px;
  box-shadow: 0 30px 60px rgba(15, 23, 42, 0.3);
}

.chart-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.chart-head h3 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.chart-head p {
  margin: 6px 0 0;
  color: #475569;
  font-size: 13px;
  font-weight: 700;
}

.chart-head .range {
  color: #2563eb;
}

.btn-close {
  border: 0;
  background: #f1f5f9;
  color: #0f172a;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 900;
  cursor: pointer;
}

.btn-close:hover {
  background: #e2e8f0;
}

.chart-state {
  padding: 40px 10px;
  text-align: center;
  color: #64748b;
  font-weight: 700;
}

.chart-state.error {
  color: #b91c1c;
}

.chart-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.stat {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat span {
  font-size: 12px;
  color: #64748b;
  font-weight: 700;
}

.stat strong {
  font-size: 16px;
  color: #0f172a;
}

.stat.good strong {
  color: #15803d;
}

.stat.bad strong {
  color: #b91c1c;
}

.chart-body {
  min-height: 360px;
}

@media (max-width: 640px) {
  .chart-stats {
    grid-template-columns: 1fr;
  }
}
</style>
