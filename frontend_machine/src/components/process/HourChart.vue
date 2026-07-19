<script setup>
defineProps({
  hours: { type: Array, required: true },
  maxHourProc: { type: Number, required: true },
  selectedDate: { type: String, required: true },
});
</script>

<template>
  <div class="process-panel">
    <div class="process-panel-head">
      <h3>Runtime / Process Time per Jam</h3>
      <span>{{ selectedDate }}</span>
    </div>

    <div class="runtime-chart">
      <div v-for="h in hours" :key="h.hour" class="hour-item">
        <div
          class="hour-bar"
          :style="{
            height:
              Math.max(8, Math.round((Number(h.procSec || 0) / maxHourProc) * 150)) +
              'px',
          }"
        >
          {{ h.procTime }}
        </div>
        <strong>{{ h.hour }}</strong>
        <small>{{ h.output }} pcs</small>
      </div>

      <div v-if="!hours.length" class="empty">Belum ada proses di tanggal ini.</div>
    </div>
  </div>
</template>
