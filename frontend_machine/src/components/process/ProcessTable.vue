<script setup>
import { computed, ref } from "vue";
import { exportProcessDetailExcel } from "../../utils/processDetailExportExcel";
import { formatDurationHHMMSS } from "../../utils/format";

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
  visiblePages: {
    type: Array,
    default: () => [],
  },
  machineName: {
    type: String,
    default: "",
  },
  uuid: {
    type: String,
    default: "",
  },
  location: {
    type: String,
    default: "",
  },
  selectedDate: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["prev", "next", "go", "notice"]);

const exporting = ref(false);

const startNo = computed(() => {
  return (props.page - 1) * props.pageSize;
});

const pageButtons = computed(() => {
  if (props.visiblePages.length) return props.visiblePages;

  const pages = [];
  const total = Math.max(1, Number(props.totalPages || 1));

  for (let i = 1; i <= Math.min(total, 5); i++) {
    pages.push(i);
  }

  return pages;
});

/** Total dari semua event (bukan hanya halaman aktif). */
const totalGapSec = computed(() => {
  return (props.events || []).reduce((sum, event) => {
    if (event?.gapSec == null) return sum;
    return sum + Math.max(0, Number(event.gapSec) || 0);
  }, 0);
});

const totalLossSec = computed(() => {
  return (props.events || []).reduce((sum, event) => {
    if (event?.lossTimeSec == null) return sum;
    return sum + Math.max(0, Number(event.lossTimeSec) || 0);
  }, 0);
});

const totalGapTime = computed(() => formatDurationHHMMSS(totalGapSec.value));
const totalLossTime = computed(() => formatDurationHHMMSS(totalLossSec.value));

function handleExportExcel() {
  if (exporting.value) return;

  exporting.value = true;

  try {
    const result = exportProcessDetailExcel({
      events: props.events,
      machineName: props.machineName,
      uuid: props.uuid,
      location: props.location,
      date: props.selectedDate,
    });

    emit(
      "notice",
      `Export Excel berhasil (${result.rowCount} baris).`,
      "ok"
    );
  } catch (err) {
    emit("notice", err?.message || "Gagal export Excel.", "error");
  } finally {
    exporting.value = false;
  }
}
</script>

<template>
  <section class="process-panel">
    <div class="process-panel-head">
      <div>
        <h3>Detail Output</h3>
        <p class="panel-sub">{{ events.length }} proses</p>
      </div>

      <div class="process-totals" v-if="events.length">
        <div class="process-total-item">
          <span>Total Jeda Waktu Antar Output</span>
          <strong>{{ totalGapTime }}</strong>
        </div>
        <div class="process-total-item">
          <span>Total Waktu Losstime</span>
          <strong>{{ totalLossTime }}</strong>
        </div>
      </div>

      <button
        type="button"
        class="export-excel-btn"
        :disabled="exporting || !events.length"
        @click="handleExportExcel"
      >
        {{ exporting ? "Exporting..." : "Export Excel" }}
      </button>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>No</th>
            <th>Program</th>
            <th>Start</th>
            <th>End</th>
            <th>Jeda Waktu Antar Output</th>
            <th>Detail Losstime</th>
            <th>Waktu Losstime</th>
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
              <template v-if="e.gapSec == null">
                <strong>-</strong>
              </template>
              <template v-else>
                <strong>{{ e.gapTime || "00:00:00" }}</strong>
                <small>{{ e.gapSec }}s</small>
              </template>
            </td>

            <td class="loss-detail">
              <template v-if="!e.detailLossTime || e.detailLossTime === '-'">
                -
              </template>
              <template v-else>
                <div
                  v-for="(line, lineIdx) in String(e.detailLossTime).split('\n')"
                  :key="`${e.startTime}-loss-${lineIdx}`"
                  class="loss-line"
                >
                  {{ line }}
                </div>
              </template>
            </td>

            <td class="right">
              <template v-if="e.lossTimeSec == null">
                <strong>-</strong>
              </template>
              <template v-else>
                <strong>{{ e.lossTime || "00:00:00" }}</strong>
                <small>{{ e.lossTimeSec }}s</small>
              </template>
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
            <td colspan="14" class="empty">
              Tidak ada proses pada mesin/tanggal ini.
            </td>
          </tr>
        </tbody>

        <tfoot v-if="events.length">
          <tr class="process-total-row">
            <td colspan="4" class="right">
              <strong>TOTAL</strong>
            </td>
            <td class="right">
              <strong>{{ totalGapTime }}</strong>
            </td>
            <td></td>
            <td class="right">
              <strong>{{ totalLossTime }}</strong>
            </td>
            <td colspan="7"></td>
          </tr>
        </tfoot>
      </table>
    </div>

    <div class="process-pagination">
      <span>
        Page {{ page }} / {{ totalPages }}
      </span>

      <div class="process-page-controls">
        <button @click="emit('prev')" :disabled="page <= 1">
          Prev
        </button>

        <button
          v-for="pageNo in pageButtons"
          :key="pageNo"
          class="page-number"
          :class="{ active: page === pageNo }"
          @click="emit('go', pageNo)"
        >
          {{ pageNo }}
        </button>

        <button @click="emit('next')" :disabled="page >= totalPages">
          Next
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel-sub {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.export-excel-btn {
  border: 0;
  background: #16a34a;
  color: #fff;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 800;
  cursor: pointer;
  white-space: nowrap;
}

.export-excel-btn:hover:not(:disabled) {
  background: #15803d;
}

.export-excel-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.loss-detail {
  min-width: 160px;
  white-space: normal;
  font-size: 12px;
  line-height: 1.45;
  font-weight: 700;
  color: #334155;
}

.loss-line + .loss-line {
  margin-top: 4px;
}

.process-totals {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-left: auto;
  margin-right: 12px;
}

.process-total-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 170px;
  padding: 8px 12px;
  border-radius: 12px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
}

.process-total-item span {
  color: #1e40af;
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.process-total-item strong {
  color: #0f172a;
  font-size: 16px;
  font-variant-numeric: tabular-nums;
}

.process-total-row td {
  background: #f8fafc;
  border-top: 2px solid #cbd5e1;
  font-variant-numeric: tabular-nums;
}
</style>
