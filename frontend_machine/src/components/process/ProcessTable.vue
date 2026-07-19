<script setup>
import { computed } from "vue";

const props = defineProps({
  events: {
    type: Array,
    required: true,
    default: () => [],
  },
  pagedEvents: {
    type: Array,
    required: true,
    default: () => [],
  },
  page: {
    type: Number,
    required: true,
    default: 1,
  },
  totalPages: {
    type: Number,
    required: true,
    default: 1,
  },
  pageSize: {
    type: Number,
    default: 20,
  },
});

const emit = defineEmits(["prev", "next"]);

const startNo = computed(() => {
  return (props.page - 1) * props.pageSize;
});
</script>

<template>
  <section class="process-panel">
    <div class="process-panel-head">
      <h3>Detail Proses Harian</h3>
      <span>{{ events.length }} proses</span>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>No</th>
            <th>Program</th>
            <th>Start</th>
            <th>End</th>
            <th>Proc Time</th>
            <th>Count</th>
            <th>Stitch</th>
            <th>Node Distance</th>
            <th>SPM</th>
            <th>Status</th>
            <th>Reason</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="(e, index) in pagedEvents" :key="`${e.no}-${e.startTime}-${index}`">
            <td class="right">
              {{ startNo + index + 1 }}
            </td>

            <td>
              <strong>{{ e.fileName || "-" }}</strong>
            </td>

            <td class="mono">
              {{ e.startTime || "-" }}
            </td>

            <td class="mono">
              {{ e.endTime || "-" }}
            </td>

            <td class="right">
              <strong>{{ e.procTime || "00:00" }}</strong>
              <small>{{ e.procSec || 0 }}s</small>
            </td>

            <td class="right">
              {{ e.procCounts || 0 }}
            </td>

            <td class="right">
              {{ e.endStitch || 0 }}/{{ e.fileStitches || 0 }}
            </td>

            <td class="right">
              {{ e.nodeDistance || 0 }}
            </td>

            <td class="right">
              {{ e.spm || 0 }}
            </td>

            <td>
              <span
                class="process-status"
                :class="e.status === 'OK' ? 'ok' : 'bad'"
              >
                {{ e.status || "ABNORMAL" }}
              </span>
            </td>

            <td>
              {{ e.abnormalReason || "-" }}
            </td>
          </tr>

          <tr v-if="!pagedEvents.length">
            <td colspan="11" class="empty">
              Tidak ada proses pada mesin/tanggal ini.
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="process-pagination">
      <span>
        Page {{ page }} / {{ totalPages }}
      </span>

      <div>
        <button @click="emit('prev')" :disabled="page <= 1">
          Prev
        </button>

        <button @click="emit('next')" :disabled="page >= totalPages">
          Next
        </button>
      </div>
    </div>
  </section>
</template>