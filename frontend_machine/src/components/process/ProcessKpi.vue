<script setup>
import { formatDurationHHMMSS } from "../../utils/format";

defineProps({
  detailMachine: { type: Object, required: true },
  selectedMachine: { type: Object, default: null },
  eventsCount: { type: Number, default: 0 },
  avgNodeDistance: { type: [String, Number], default: "-" },
});
</script>

<template>
  <section class="process-kpis">
    <div class="process-kpi green">
      <span>Total Workcount</span>
      <strong>{{ detailMachine.output || selectedMachine?.output || 0 }}</strong>
    </div>

    <div class="process-kpi amber">
      <span>Plan Rate</span>
      <strong>{{ Number(detailMachine.productivityPct || selectedMachine?.productivity || 0).toFixed(2) }}%</strong>
    </div>

    <div class="process-kpi">
      <span>Runtime</span>
      <strong>{{ formatDurationHHMMSS(detailMachine.runtimeSec || selectedMachine?.runtime || 0) }}</strong>
    </div>

    <div class="process-kpi">
      <span>Process Time</span>
      <strong>{{ formatDurationHHMMSS(detailMachine.procSec || selectedMachine?.procTime || 0) }}</strong>
    </div>

    <div class="process-kpi blue">
      <span>Axis Speed</span>
      <strong>{{ avgNodeDistance }}</strong>
    </div>

    <div class="process-kpi purple">
      <span>Total Process</span>
      <strong>{{ detailMachine.cycles || eventsCount || 0 }}</strong>
    </div>

    <div class="process-kpi purple">
      <span>Abnormal</span>
      <strong>{{ detailMachine.incomplete || selectedMachine?.abnormal || 0 }}</strong>
    </div>
  </section>
</template>
