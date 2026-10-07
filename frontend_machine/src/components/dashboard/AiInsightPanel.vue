<script setup>
import { computed, ref } from "vue";

const props = defineProps({
  insight: {
    type: Object,
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  errorMessage: {
    type: String,
    default: "",
  },
  lastUpdatedAt: {
    type: String,
    default: "",
  },
  embedded: {
    type: Boolean,
    default: false,
  },
  fullscreen: {
    type: Boolean,
    default: false,
  },
  hideToolbar: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits([
  "refresh",
  "filter-item",
  "open-chart",
  "ask-ai",
  "why-bad",
]);

const expandedKey = ref("");

const items = computed(() =>
  Array.isArray(props.insight?.items) ? props.insight.items : []
);

const highCount = computed(() => Number(props.insight?.highCount || 0));
const mediumCount = computed(() => Number(props.insight?.mediumCount || 0));
const machineCount = computed(() => Number(props.insight?.machineCount || 0));

function itemKey(item, index) {
  return `${item.code}-${item.uuid || item.location || index}`;
}

function toggleExpand(key) {
  expandedKey.value = expandedKey.value === key ? "" : key;
}

function titleOf(item) {
  if (String(item?.scope || "").toUpperCase() === "LINE") {
    return String(item?.location || "").trim() || "Line";
  }
  const machine = String(item?.machineName || "").trim();
  return machine || String(item?.location || "").trim() || "Mesin";
}

function subtitleOf(item) {
  if (String(item?.scope || "").toUpperCase() === "LINE") {
    return String(item?.signal || "").trim();
  }
  return String(item?.location || item?.signal || "").trim();
}

function nextStep(item) {
  return (
    String(item?.improvement || "").trim() ||
    String(item?.decision || "").trim() ||
    "Monitor dan follow-up utilisasi mesin."
  );
}

function stepLines(item) {
  const raw = nextStep(item);
  if (!raw) return [];

  const chunks = raw
    .split(/\n|(?:\s*→\s*)/)
    .map((line) => line.replace(/^\d+\.\s*/, "").trim())
    .filter(Boolean);

  return chunks.length ? chunks : [raw];
}

function priorityTone(priority) {
  const p = String(priority || "").toUpperCase();
  if (p === "HIGH") return "is-high";
  if (p === "MEDIUM") return "is-medium";
  return "is-low";
}

function findingTone(item) {
  const code = String(item?.code || "").toUpperCase();
  const loss = String(item?.lossReasonCode || item?.lossReasonLabel || "").toUpperCase();
  const note = String(item?.noteSummary || "").toUpperCase();
  const blob = `${loss} ${note}`;

  if (code === "FREQUENT_BROKEN" || isFrequentBroken(item)) return "code-broken";
  if (code === "MACHINE_OFFLINE") return "code-offline";
  if (code === "IDLE_NO_NOTE") return "code-note";
  if (code === "LINE_DECISION" || String(item?.scope || "").toUpperCase() === "LINE") {
    return "code-line";
  }
  if (
    blob.includes("HANCA") ||
    blob.includes("TUNGGU BAHAN") ||
    loss === "WAIT_HANCA"
  ) {
    return "code-hanca";
  }
  if (code === "ACTIVE_LOSS") {
    if (blob.includes("BROKEN") || blob.includes("RUSAK")) return "code-broken";
    return "code-loss";
  }
  if (code === "NOTE_CONTEXT_LOW_UTIL") return "code-context";
  return "";
}

function findingLabel(item) {
  const tone = findingTone(item);
  if (tone === "code-broken") {
    const n = Number(item?.brokenCount7d || 0);
    return n > 0 ? `Rusak ${n}x` : "Rusak";
  }
  if (tone === "code-offline") return "Offline";
  if (tone === "code-note") return "Tanpa note";
  if (tone === "code-context") return "Ada note";
  if (tone === "code-hanca") return "Hanca";
  if (tone === "code-line") return "Line";
  if (tone === "code-loss") return "Loss aktif";
  return "";
}

function scopeLabel(item) {
  return String(item?.scope || "MESIN").toUpperCase();
}

function isFrequentBroken(item) {
  const code = String(item?.code || "").toUpperCase();
  const count = Number(item?.brokenCount7d || 0);
  if (code === "FREQUENT_BROKEN") return true;
  if (count >= 3) return true;
  const loss = String(item?.lossReasonCode || item?.lossReasonLabel || "").toUpperCase();
  return count >= 3 || (code === "ACTIVE_LOSS" && loss.includes("BROKEN") && count >= 2);
}

function canFilter(item) {
  return Boolean(String(item?.location || "").trim());
}

function canOpenChart(item) {
  return Boolean(String(item?.uuid || "").trim());
}

function canShowWhyBad(item) {
  return (
    String(item?.scope || "").toUpperCase() !== "LINE" &&
    Boolean(String(item?.uuid || "").trim())
  );
}

function onFilter(item, event) {
  event.stopPropagation();
  if (!canFilter(item)) return;
  emit("filter-item", item);
}

function onOpenChart(item, event) {
  event.stopPropagation();
  if (!canOpenChart(item)) return;
  emit("open-chart", item);
}

function onAskAi(item, event) {
  event.stopPropagation();
  emit("ask-ai", item);
}

function onWhyBad(item, event) {
  event.stopPropagation();
  if (!canShowWhyBad(item)) return;
  emit("why-bad", item);
}
</script>

<template>
  <section
    class="ai-ins"
    :class="{ 'ai-ins--embedded': embedded, 'ai-ins--full': fullscreen }"
  >
    <div v-if="!hideToolbar" class="ai-ins__toolbar">
      <div class="ai-ins__stats">
        <span class="ai-ins__pill ai-ins__pill--high">{{ highCount }} segera</span>
        <span class="ai-ins__pill ai-ins__pill--medium">{{ mediumCount }} perhatian</span>
        <span class="ai-ins__pill ai-ins__pill--mute">{{ machineCount }} aktif</span>
      </div>
      <button
        type="button"
        class="ai-ins__refresh"
        :disabled="loading"
        @click="emit('refresh')"
      >
        {{ loading ? "Analisa..." : "Refresh" }}
      </button>
    </div>

    <div v-else class="ai-ins__stats ai-ins__stats--bar">
      <span class="ai-ins__pill ai-ins__pill--high">{{ highCount }} segera</span>
      <span class="ai-ins__pill ai-ins__pill--medium">{{ mediumCount }} perhatian</span>
      <span class="ai-ins__pill ai-ins__pill--mute">{{ machineCount }} aktif</span>
      <span v-if="lastUpdatedAt" class="ai-ins__meta">Update {{ lastUpdatedAt }}</span>
    </div>

    <p v-if="lastUpdatedAt && !hideToolbar" class="ai-ins__meta">Update {{ lastUpdatedAt }}</p>

    <p v-if="errorMessage" class="ai-ins__error">{{ errorMessage }}</p>

    <div v-else-if="insight?.summary" class="ai-ins__brief">
      <strong>Ringkas</strong>
      <p>{{ insight.summary }}</p>
    </div>

    <div class="ai-ins__grid">
      <article
        v-for="(item, index) in items"
        :key="itemKey(item, index)"
        class="ai-ins__item"
        :class="[
          priorityTone(item.priority),
          findingTone(item),
          { 'is-open': expandedKey === itemKey(item, index) },
        ]"
        @click="toggleExpand(itemKey(item, index))"
      >
        <div class="ai-ins__item-top">
          <span class="ai-ins__status" aria-hidden="true" />
          <h4 class="ai-ins__title">{{ titleOf(item) }}</h4>
          <div class="ai-ins__tags">
            <span class="ai-ins__tag ai-ins__tag--scope">{{ scopeLabel(item) }}</span>
            <span
              v-if="findingLabel(item)"
              class="ai-ins__tag"
              :class="`ai-ins__tag--${findingTone(item)}`"
            >
              {{ findingLabel(item) }}
            </span>
            <span class="ai-ins__tag ai-ins__tag--prio">{{ item.priority || "LOW" }}</span>
          </div>
        </div>

        <p v-if="subtitleOf(item)" class="ai-ins__subtitle">{{ subtitleOf(item) }}</p>

        <div v-if="item.decision" class="ai-ins__decision">
          <span class="ai-ins__label">Keputusan</span>
          <p>{{ item.decision }}</p>
        </div>

        <div class="ai-ins__step">
          <span class="ai-ins__step-label">Langkah</span>
          <ol v-if="stepLines(item).length > 1" class="ai-ins__steps">
            <li v-for="(line, stepIdx) in stepLines(item)" :key="`${itemKey(item, index)}-step-${stepIdx}`">
              {{ line }}
            </li>
          </ol>
          <p v-else class="ai-ins__step-one">{{ stepLines(item)[0] }}</p>
        </div>

        <div class="ai-ins__actions" @click.stop>
          <button
            v-if="canShowWhyBad(item)"
            type="button"
            class="ai-ins__act"
            @click="onWhyBad(item, $event)"
          >
            Kenapa
          </button>
          <button
            v-if="canFilter(item)"
            type="button"
            class="ai-ins__act"
            @click="onFilter(item, $event)"
          >
            Filter
          </button>
          <button
            v-if="canOpenChart(item)"
            type="button"
            class="ai-ins__act"
            @click="onOpenChart(item, $event)"
          >
            Chart
          </button>
          <button type="button" class="ai-ins__act ai-ins__act--main" @click="onAskAi(item, $event)">
            Tanya AI
          </button>
        </div>

        <div v-if="expandedKey === itemKey(item, index)" class="ai-ins__detail">
          <p v-if="item.prediction1h">
            <b>Proyeksi 1 jam:</b> {{ item.prediction1h }}
          </p>
          <p v-if="item.evidence"><b>Bukti:</b> {{ item.evidence }}</p>
          <p v-if="item.noteSummary">
            <b>Keterangan operator:</b> {{ item.noteSummary }}
          </p>
        </div>

        <span class="ai-ins__hint">
          {{
            expandedKey === itemKey(item, index)
              ? "Klik kartu untuk tutup detail"
              : "Klik kartu untuk bukti & note"
          }}
        </span>
      </article>

      <div v-if="!loading && !items.length && !errorMessage" class="ai-ins__empty">
        Situasi relatif aman. Tidak ada prioritas mendesak saat ini.
      </div>
    </div>
  </section>
</template>

<style scoped>
.ai-ins {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  flex: 1;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.ai-ins--full {
  gap: 14px;
}

.ai-ins__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
}

.ai-ins__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.ai-ins__stats--bar {
  flex-shrink: 0;
}

.ai-ins__pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 6px 11px;
  font-size: 12px;
  font-weight: 800;
  line-height: 1.2;
}

.ai-ins__pill--high {
  background: #fee2e2;
  color: #b91c1c;
}

.ai-ins__pill--medium {
  background: #ffedd5;
  color: #c2410c;
}

.ai-ins__pill--mute {
  background: #f1f5f9;
  color: #64748b;
}

.ai-ins__refresh {
  border: 0;
  border-radius: 10px;
  padding: 8px 12px;
  background: #1d4ed8;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  flex-shrink: 0;
}

.ai-ins__refresh:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.ai-ins__meta {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
  font-weight: 700;
  flex-shrink: 0;
}

.ai-ins__error {
  margin: 0;
  padding: 12px 14px;
  border-radius: 12px;
  background: #fef2f2;
  color: #b91c1c;
  font-size: 13px;
  font-weight: 700;
  flex-shrink: 0;
}

.ai-ins__brief {
  margin: 0;
  padding: 12px 14px;
  border-radius: 14px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  flex-shrink: 0;
}

.ai-ins__brief strong {
  display: block;
  font-size: 11px;
  color: #1d4ed8;
  margin-bottom: 6px;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.ai-ins__brief p {
  margin: 0;
  font-size: clamp(13px, 1.35vw, 15px);
  line-height: 1.55;
  color: #1e3a8a;
  font-weight: 650;
}

.ai-ins__grid {
  flex: 1;
  min-height: 0;
  overflow: auto;
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  align-content: start;
  padding: 2px 6px 12px 0;
  box-sizing: border-box;
}

/* Responsif otomatis ke lebar layar */
.ai-ins--full .ai-ins__grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
  gap: clamp(12px, 1.6vw, 18px);
}

.ai-ins__item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 0;
  padding: clamp(14px, 1.6vw, 18px);
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  box-sizing: border-box;
  min-width: 0;
  height: fit-content;
}

.ai-ins__item:hover {
  border-color: #cbd5e1;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
}

.ai-ins__item.is-high {
  border-color: #fecaca;
  background: #fffafa;
}

.ai-ins__item.is-medium {
  border-color: #fed7aa;
  background: #fffaf5;
}

.ai-ins__item.code-broken {
  border-color: #fca5a5;
  background: linear-gradient(180deg, #fff5f5 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #ef4444;
}

.ai-ins__item.code-offline {
  border-color: #cbd5e1;
  background: linear-gradient(180deg, #f8fafc 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #64748b;
}

.ai-ins__item.code-note {
  border-color: #fcd34d;
  background: linear-gradient(180deg, #fffbeb 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #f59e0b;
}

.ai-ins__item.code-context {
  border-color: #fdba74;
  background: linear-gradient(180deg, #fff7ed 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #ea580c;
}

.ai-ins__item.code-hanca {
  border-color: #67e8f9;
  background: linear-gradient(180deg, #ecfeff 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #0891b2;
}

.ai-ins__item.code-line {
  border-color: #93c5fd;
  background: linear-gradient(180deg, #eff6ff 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #2563eb;
}

.ai-ins__item.code-loss {
  border-color: #f9a8d4;
  background: linear-gradient(180deg, #fdf2f8 0%, #fff 72%);
  box-shadow: inset 4px 0 0 #db2777;
}

.ai-ins__item.code-broken .ai-ins__status { background: #ef4444; }
.ai-ins__item.code-offline .ai-ins__status { background: #64748b; }
.ai-ins__item.code-note .ai-ins__status { background: #f59e0b; }
.ai-ins__item.code-context .ai-ins__status { background: #ea580c; }
.ai-ins__item.code-hanca .ai-ins__status { background: #0891b2; }
.ai-ins__item.code-line .ai-ins__status { background: #2563eb; }
.ai-ins__item.code-loss .ai-ins__status { background: #db2777; }

.ai-ins__item.is-open {
  border-color: #60a5fa;
  background: linear-gradient(180deg, #eff6ff 0%, #dbeafe 100%);
  box-shadow: 0 12px 28px rgba(37, 99, 235, 0.16);
}

.ai-ins__item.is-open.is-high,
.ai-ins__item.is-open.is-medium {
  border-color: #60a5fa;
  background: linear-gradient(180deg, #eff6ff 0%, #dbeafe 100%);
}

.ai-ins__item.is-open .ai-ins__title {
  color: #1e3a8a;
}

.ai-ins__item.is-open .ai-ins__subtitle {
  color: #334155;
  font-weight: 500;
}

.ai-ins__item.is-open .ai-ins__decision {
  background: #ffffff;
  border-color: #93c5fd;
}

.ai-ins__item.is-open .ai-ins__label {
  color: #2563eb;
}

.ai-ins__item.is-open .ai-ins__decision p {
  color: #1e40af;
  font-weight: 600;
}

.ai-ins__item.is-open .ai-ins__step {
  color: #334155;
}

.ai-ins__item.is-open .ai-ins__step-label {
  color: #64748b;
}

.ai-ins__item.is-open .ai-ins__act {
  background: #ffffff;
  border-color: #93c5fd;
  color: #1d4ed8;
}

.ai-ins__item.is-open .ai-ins__act--main {
  background: #2563eb;
  border-color: #2563eb;
  color: #ffffff;
}

.ai-ins__item.is-open .ai-ins__act:hover {
  background: #dbeafe;
}

.ai-ins__item.is-open .ai-ins__act--main:hover {
  background: #1d4ed8;
  border-color: #1d4ed8;
  color: #ffffff;
}

.ai-ins__item.is-open .ai-ins__detail {
  border-top-color: #93c5fd;
}

.ai-ins__item.is-open .ai-ins__detail p {
  color: #334155;
}

.ai-ins__item.is-open .ai-ins__detail b {
  color: #1e40af;
  font-weight: 600;
}

.ai-ins__item.is-open .ai-ins__hint {
  color: #64748b;
}

.ai-ins__item.is-open .ai-ins__status {
  background: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
}

.ai-ins__item-top {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
}

.ai-ins__status {
  width: 8px;
  height: 8px;
  margin-top: 6px;
  border-radius: 50%;
  background: #94a3b8;
  flex-shrink: 0;
}

.ai-ins__item.is-high .ai-ins__status {
  background: #ef4444;
}

.ai-ins__item.is-medium .ai-ins__status {
  background: #f97316;
}

.ai-ins__title {
  margin: 0;
  flex: 1;
  min-width: 0;
  font-size: clamp(14px, 1.5vw, 16px);
  font-weight: 800;
  color: #0f172a;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.ai-ins__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
  flex-shrink: 0;
  max-width: 46%;
}

.ai-ins__tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 3px 8px;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.2;
  white-space: nowrap;
}

.ai-ins__tag--scope {
  background: #dbeafe;
  color: #1d4ed8;
}

.ai-ins__tag--prio {
  background: #f1f5f9;
  color: #64748b;
}

.ai-ins__tag--code-broken {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.ai-ins__tag--code-offline {
  background: #e2e8f0;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.ai-ins__tag--code-note {
  background: #fef3c7;
  color: #92400e;
  border: 1px solid #fcd34d;
}

.ai-ins__tag--code-context {
  background: #ffedd5;
  color: #9a3412;
  border: 1px solid #fdba74;
}

.ai-ins__tag--code-hanca {
  background: #cffafe;
  color: #155e75;
  border: 1px solid #67e8f9;
}

.ai-ins__tag--code-line {
  background: #dbeafe;
  color: #1e40af;
  border: 1px solid #93c5fd;
}

.ai-ins__tag--code-loss {
  background: #fce7f3;
  color: #9d174d;
  border: 1px solid #f9a8d4;
}

.ai-ins__item.is-high .ai-ins__tag--prio {
  background: #fee2e2;
  color: #b91c1c;
}

.ai-ins__item.is-medium .ai-ins__tag--prio {
  background: #ffedd5;
  color: #c2410c;
}

.ai-ins__subtitle {
  margin: 0;
  font-size: clamp(12px, 1.3vw, 14px);
  line-height: 1.45;
  color: #475569;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.ai-ins__decision {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 12px 14px;
  border-radius: 12px;
  background: #eff6ff;
  border: 1px solid #93c5fd;
}

.ai-ins__label {
  display: block;
  font-size: 10px;
  font-weight: 900;
  color: #1d4ed8;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  line-height: 1.2;
}

.ai-ins__decision p {
  margin: 0;
  font-size: clamp(13px, 1.35vw, 14.5px);
  line-height: 1.5;
  color: #1e3a8a;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.ai-ins__step {
  margin: 0;
  color: #475569;
  font-weight: 400;
}

.ai-ins__step-label {
  display: block;
  margin-bottom: 6px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #94a3b8;
}

.ai-ins__steps {
  margin: 0;
  padding-left: 18px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-ins__steps li,
.ai-ins__step-one {
  margin: 0;
  font-size: clamp(12.5px, 1.25vw, 13.5px);
  line-height: 1.45;
  color: #475569;
  font-weight: 400;
  overflow-wrap: anywhere;
}

.ai-ins__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
}

.ai-ins__act {
  border: 1px solid #bfdbfe;
  background: #fff;
  color: #1d4ed8;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
}

.ai-ins__act--main {
  background: #eff6ff;
  border-color: #93c5fd;
}

.ai-ins__act:hover {
  background: #dbeafe;
}

.ai-ins__detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.ai-ins__detail p {
  margin: 0;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
  font-weight: 400;
  overflow-wrap: anywhere;
}

.ai-ins__detail b {
  color: #64748b;
  font-weight: 600;
}

.ai-ins__hint {
  display: block;
  margin: 0;
  font-size: 11px;
  color: #94a3b8;
  font-weight: 500;
}

.ai-ins__empty {
  grid-column: 1 / -1;
  padding: 40px 16px;
  text-align: center;
  color: #64748b;
  font-size: 14px;
  font-weight: 700;
}

@media (max-width: 560px) {
  .ai-ins__item-top {
    flex-wrap: wrap;
  }

  .ai-ins__tags {
    max-width: 100%;
    width: 100%;
    padding-left: 18px;
    justify-content: flex-start;
  }
}
</style>
