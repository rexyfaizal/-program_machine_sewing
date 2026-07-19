<script setup>
defineProps({
  events: {
    type: Array,
    required: true,
    default: () => [],
  },
});

function formatTimeOnly(value) {
  if (!value) return "-";

  const text = String(value);
  return text.length >= 16 ? text.slice(11, 16) : text;
}
</script>

<template>
  <section class="process-panel timeline-panel">
    <div class="process-panel-head">
      <h3>Process Detail Timeline</h3>
    </div>

    <div class="timeline-wrap">
      <div class="timeline-row">
        <div
          v-for="(e, index) in events.slice(0, 30)"
          :key="`${e.no}-${e.startTime}-${index}`"
          class="timeline-step"
          :class="{ abnormal: e.status === 'ABNORMAL' }"
        >
          <div class="step-no">{{ index + 1 }}</div>

          <strong :title="e.fileName">
            {{ e.fileName || "-" }}
          </strong>

          <b>{{ e.procTime || "00:00" }}</b>

          <small>
            {{ formatTimeOnly(e.startTime) }} - {{ formatTimeOnly(e.endTime) }}
          </small>

          <span
            class="process-status"
            :class="e.status === 'OK' ? 'ok' : 'bad'"
          >
            {{ e.status || "ABNORMAL" }}
          </span>
        </div>

        <div v-if="!events.length" class="empty">
          Belum ada process detail.
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.timeline-panel {
  width: 100%;
  max-width: 760px;
  margin: 0;
}

.process-panel {
  background: white;
  border: 1px solid #dbe4ef;
  border-radius: 20px;
  padding: 18px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
}

.process-panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.process-panel-head h3 {
  margin: 0;
  font-size: 17px;
  color: #0f172a;
}

.process-panel-head span {
  color: #64748b;
  font-size: 12px;
  font-weight: 800;
}

.timeline-wrap {
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 4px 4px 14px;
  border-radius: 14px;
}

.timeline-row {
  display: flex;
  gap: 14px;
  width: max-content;
}

.timeline-step {
  width: 180px;
  min-width: 180px;
  min-height: 150px;
  background: #f8fafc;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
  padding: 14px;
}

.timeline-step.abnormal {
  background: #fff7f7;
  border-color: #fecaca;
}

.step-no {
  color: #1d4ed8;
  font-size: 18px;
  font-weight: 900;
  margin-bottom: 8px;
}

.timeline-step strong {
  display: block;
  color: #0f172a;
  font-size: 12px;
  font-weight: 900;
  line-height: 1.35;
  min-height: 32px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.timeline-step b {
  display: block;
  color: #0f172a;
  font-size: 18px;
  margin-top: 8px;
}

.timeline-step small {
  display: block;
  color: #64748b;
  font-size: 12px;
  margin-top: 5px;
}

.process-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 900;
  margin-top: 10px;
}

.process-status.ok {
  background: #dcfce7;
  color: #166534;
}

.process-status.bad {
  background: #fee2e2;
  color: #991b1b;
}

.timeline-wrap::-webkit-scrollbar {
  height: 12px;
}

.timeline-wrap::-webkit-scrollbar-track {
  background: #e5e7eb;
  border-radius: 999px;
}

.timeline-wrap::-webkit-scrollbar-thumb {
  background: #8b8f98;
  border-radius: 999px;
}

.timeline-wrap::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}

.empty {
  min-width: 400px;
  text-align: center;
  padding: 30px;
  color: #64748b;
  font-weight: 800;
}

@media (max-width: 900px) {
  .timeline-panel {
    max-width: 100%;
  }
}
</style>