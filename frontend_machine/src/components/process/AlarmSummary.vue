<script setup>
defineProps({
  alarms: { type: Array, required: true },
});
</script>

<template>
  <div class="process-panel">
    <div class="process-panel-head">
      <h3>Alarm Summary</h3>
      <span>top alarm</span>
    </div>

    <div class="alarm-list">
      <div v-for="a in alarms" :key="a.content" class="alarm-row">
        <span>{{ a.content }}</span>
        <strong>{{ a.total }}</strong>
      </div>

      <div v-if="!alarms.length" class="empty">Tidak ada alarm.</div>
    </div>
  </div>
</template>
