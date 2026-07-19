<script setup>
defineProps({
  groups: { type: Array, required: true },
  maxGroupProc: { type: Number, required: true },
});
</script>

<template>
  <div class="process-panel">
    <div class="process-panel-head">
      <h3>Process Count / Process Time / Abnormal</h3>
      <span>per program jahit</span>
    </div>

    <div class="program-list">
      <div v-for="g in groups.slice(0, 12)" :key="g.fileName" class="program-row">
        <strong :title="g.fileName">{{ g.fileName }}</strong>

        <div class="program-track">
          <div
            class="program-fill"
            :style="{
              width:
                Math.max(2, (Number(g.procSec || 0) / maxGroupProc) * 100) + '%',
            }"
          ></div>
        </div>

        <div class="program-info">
          <b>{{ g.output }}</b>
          <small>{{ g.procTime }}</small>
          <small :class="{ redText: Number(g.incomplete || 0) > 0 }">
            {{ g.incomplete }} abnormal
          </small>
        </div>
      </div>

      <div v-if="!groups.length" class="empty">Tidak ada program.</div>
    </div>
  </div>
</template>
