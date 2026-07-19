<script setup>
import { computed } from "vue";

const props = defineProps({
  events: {
    type: Array,
    required: true,
    default: () => [],
  },
  groups: {
    type: Array,
    default: () => [],
  },
  limit: {
    type: Number,
    default: 30,
  },
});

const itemWidth = 145;
const chartHeight = 330;
const lineTop = 36;
const lineAreaHeight = 110;
const barBaseY = 250;
const maxBarHeight = 72;
const minBarHeight = 50;

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function normalizeName(value) {
  return String(value || "").trim().toUpperCase();
}

function formatTimeOnly(value) {
  if (!value) return "-";

  const text = String(value);
  return text.length >= 16 ? text.slice(11, 16) : text;
}

function shortProgramName(value) {
  const text = String(value || "-");
  return text.length > 24 ? text.slice(0, 24) + "..." : text;
}

function secondsToHHMMSS(totalSec) {
  const sec = Math.max(0, toNumber(totalSec));
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = Math.floor(sec % 60);

  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(
    s
  ).padStart(2, "0")}`;
}

function getEventCount(e) {
  const count = toNumber(
    e.procCounts ??
      e.ProcCounts ??
      e.count ??
      e.Count ??
      0
  );

  if (count > 0) return count;

  if (String(e.status || "").toUpperCase() === "OK") return 1;

  return 0;
}

function getDirectEventOutput(e) {
  return toNumber(
    e.output ??
      e.Output ??
      e.outputNo ??
      e.output_no ??
      e.workCount ??
      e.WorkCount ??
      e.totalWorkCount ??
      e.totalWorkcount ??
      0
  );
}

/*
  Build map total output per program dari groups.
  Contoh:
  MTB725 140 SLEEVE KA => 20
  MTB725 140 SLEEVE KI => 19
*/
const groupOutputMap = computed(() => {
  const map = new Map();

  for (const g of props.groups || []) {
    const key = normalizeName(g.fileName ?? g.FileName);
    const output = toNumber(g.output ?? g.Output ?? 0);

    if (key) {
      map.set(key, output);
    }
  }

  return map;
});

/*
  events biasanya sudah urut terbaru -> terlama.

  Logika output:
  1. Kalau event punya field output langsung, pakai itu.
  2. Kalau belum ada output per event, pakai groups.output per program,
     lalu dihitung mundur berdasarkan procCounts.
*/
const latestEventsWithOutput = computed(() => {
  const latestEvents = [...props.events].slice(0, props.limit);

  const outputCursor = new Map(groupOutputMap.value);

  return latestEvents.map((e) => {
    const directOutput = getDirectEventOutput(e);

    if (directOutput > 0) {
      return {
        ...e,
        chartOutput: directOutput,
      };
    }

    const fileName = e.fileName ?? e.FileName ?? "";
    const key = normalizeName(fileName);
    const currentOutput = toNumber(outputCursor.get(key));
    const count = getEventCount(e);

    if (key && currentOutput > 0) {
      outputCursor.set(key, Math.max(0, currentOutput - count));

      return {
        ...e,
        chartOutput: currentOutput,
      };
    }

    return {
      ...e,
      chartOutput: 0,
    };
  });
});

/*
  Supaya grafik mirip aplikasi bawaan,
  kiri -> kanan dibuat kronologis lama ke baru.
*/
const chartEvents = computed(() => {
  return [...latestEventsWithOutput.value].reverse();
});

const maxProcSec = computed(() => {
  return Math.max(1, ...chartEvents.value.map((e) => toNumber(e.procSec)));
});

const outputValues = computed(() => {
  return chartEvents.value
    .map((e) => toNumber(e.chartOutput))
    .filter((n) => n > 0);
});

const maxOutput = computed(() => {
  return outputValues.value.length ? Math.max(...outputValues.value) : 1;
});

const minOutput = computed(() => {
  return outputValues.value.length ? Math.min(...outputValues.value) : 0;
});

const chartWidth = computed(() => {
  return Math.max(900, chartEvents.value.length * itemWidth + 120);
});

const points = computed(() => {
  return chartEvents.value.map((e, index) => {
    const x = 70 + index * itemWidth;

    const procSec = toNumber(e.procSec);
    const procRatio = procSec / maxProcSec.value;
    const y = lineTop + lineAreaHeight - procRatio * lineAreaHeight;

    const output = toNumber(e.chartOutput);

    const outputRange = Math.max(1, maxOutput.value - minOutput.value);
    const outputRatio =
      output > 0 ? (output - minOutput.value) / outputRange : 0;

    const barHeight =
      output <= 0
        ? minBarHeight
        : minBarHeight + outputRatio * (maxBarHeight - minBarHeight);

    return {
      x,
      y,
      procSec,
      procTime: e.procTime || secondsToHHMMSS(procSec),
      output,
      fileName: e.fileName || e.FileName || "-",
      startTime: e.startTime,
      endTime: e.endTime,
      status: String(e.status || "OK").toUpperCase(),
      barHeight,
    };
  });
});

const polylinePoints = computed(() => {
  return points.value.map((p) => `${p.x},${p.y}`).join(" ");
});
</script>

<template>
  <section class="process-panel output-chart-panel">
    <div class="process-panel-head">
      <h3>Process Output Diagram</h3>
      <span>mengikuti data timeline proses</span>
    </div>

    <div class="chart-scroll">
      <div
        class="chart-stage"
        :style="{
          width: chartWidth + 'px',
          height: chartHeight + 'px',
        }"
      >
        <svg
          class="chart-svg"
          :width="chartWidth"
          :height="chartHeight"
          :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
        >
          <line
            x1="20"
            :y1="barBaseY"
            :x2="chartWidth - 20"
            :y2="barBaseY"
            stroke="#e2e8f0"
            stroke-width="2"
          />

          <polyline
            v-if="points.length"
            :points="polylinePoints"
            fill="none"
            stroke="#f2b63d"
            stroke-width="4"
            stroke-linecap="round"
            stroke-linejoin="round"
          />

          <g v-for="(p, index) in points" :key="`point-${index}`">
            <circle
              :cx="p.x"
              :cy="p.y"
              r="5"
              fill="#f2b63d"
              stroke="#ffffff"
              stroke-width="2"
            />

            <text
              :x="p.x"
              :y="p.y - 12"
              text-anchor="middle"
              class="duration-text"
            >
              {{ secondsToHHMMSS(p.procSec) }}
            </text>
          </g>
        </svg>

        <div
          v-for="(p, index) in points"
          :key="`bar-${p.startTime}-${index}`"
          class="bar-item"
          :class="{ abnormal: p.status === 'ABNORMAL' }"
          :style="{
            left: p.x - 48 + 'px',
            top: barBaseY - p.barHeight + 'px',
            height: p.barHeight + 'px',
          }"
        >
          <div class="bar-output">
            {{ p.output > 0 ? p.output : "-" }}
          </div>

          <div class="bar-program" :title="p.fileName">
            {{ shortProgramName(p.fileName) }}
          </div>

          <div class="bar-time">
            {{ formatTimeOnly(p.startTime) }}
          </div>
        </div>

        <div
          v-for="(p, index) in points"
          :key="`axis-${p.startTime}-${index}`"
          class="axis-label"
          :style="{
            left: p.x - 62 + 'px',
            top: barBaseY + 12 + 'px',
          }"
        >
          {{ p.startTime || "-" }}
        </div>

        <div v-if="!points.length" class="empty-chart">
          Belum ada data process detail.
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.output-chart-panel {
  width: 100%;
}

.chart-scroll {
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 6px 6px 18px;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.chart-stage {
  position: relative;
  min-width: 900px;
}

.chart-svg {
  position: absolute;
  left: 0;
  top: 0;
  pointer-events: none;
}

.duration-text {
  fill: #3558d8;
  font-size: 12px;
  font-weight: 900;
}

.bar-item {
  position: absolute;
  width: 96px;
  background: #85cd69;
  border: 1px solid #69b54c;
  border-radius: 2px 2px 0 0;
  text-align: center;
  color: #111827;
  overflow: hidden;
  display: grid;
  align-content: start;
  padding: 4px 4px 3px;
}

.bar-item.abnormal {
  background: #fecaca;
  border-color: #ef4444;
}

.bar-output {
  font-size: 20px;
  line-height: 1;
  font-weight: 900;
  margin-bottom: 3px;
  font-family: "Times New Roman", serif;
}

.bar-program {
  font-size: 9px;
  line-height: 1.08;
  font-weight: 800;
  min-height: 22px;
  max-height: 24px;
  overflow: hidden;
  text-transform: uppercase;
}

.bar-time {
  font-size: 10px;
  font-weight: 800;
  margin-top: 3px;
}

.axis-label {
  position: absolute;
  width: 124px;
  font-size: 11px;
  line-height: 1.15;
  color: #111827;
  text-align: center;
  white-space: normal;
  word-break: break-word;
}

.empty-chart {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: #64748b;
  font-weight: 800;
}

.chart-scroll::-webkit-scrollbar {
  height: 12px;
}

.chart-scroll::-webkit-scrollbar-track {
  background: #e5e7eb;
  border-radius: 999px;
}

.chart-scroll::-webkit-scrollbar-thumb {
  background: #8b8f98;
  border-radius: 999px;
}

.chart-scroll::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>