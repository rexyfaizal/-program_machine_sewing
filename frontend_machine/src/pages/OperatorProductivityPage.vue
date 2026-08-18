<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useOperatorProductivity } from "../composables/useOperatorProductivity";
import { exportOperatorProductivityExcel } from "../utils/operatorProductivityExcel";
import { formatDurationHHMMSS } from "../utils/format";

const props = defineProps({
  selectedDate: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update:selectedDate"]);

const exporting = ref(false);
const notice = ref("");
const noticeType = ref("ok");

const {
  loading,
  errorMessage,
  keyword,
  locationFilter,
  locationOptions,
  filteredRows,
  averages,
  loggedInCount,
  unloggedCount,
  formatLongDate,
  loadOperatorProductivity,
} = useOperatorProductivity();

const localDate = computed({
  get: () => props.selectedDate,
  set: (value) => emit("update:selectedDate", value),
});

const dateLabel = computed(() => formatLongDate(localDate.value));

const page = ref(1);
const pageSize = 10;

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(filteredRows.value.length / pageSize));
});

const pagedRows = computed(() => {
  const start = (page.value - 1) * pageSize;
  return filteredRows.value.slice(start, start + pageSize);
});

const visiblePages = computed(() => {
  const pages = [];
  const total = totalPages.value;
  const current = page.value;

  let start = Math.max(1, current - 2);
  let end = Math.min(total, current + 2);

  if (current <= 3) {
    end = Math.min(total, 5);
  }

  if (current >= total - 2) {
    start = Math.max(1, total - 4);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});

function goPage(pageNumber) {
  const next = Number(pageNumber);
  if (!Number.isFinite(next)) return;
  page.value = Math.min(totalPages.value, Math.max(1, next));
}

function formatCycle(value) {
  return Number(value || 0).toFixed(2);
}

function formatPct(value) {
  return `${Number(value || 0).toFixed(2)}%`;
}

function shiftClass(tag) {
  const text = String(tag || "Normal")
    .trim()
    .toUpperCase();

  if (text.includes("SHIFT 1") || text === "SHIFT_1") return "shift-1";
  if (text.includes("SHIFT 2") || text === "SHIFT_2") return "shift-2";
  if (text.includes("SHIFT 3") || text === "SHIFT_3") return "shift-3";
  if (text.includes("ALL")) return "shift-all";
  return "shift-normal";
}

async function handleExport() {
  if (exporting.value) return;
  exporting.value = true;
  notice.value = "";

  try {
    const result = exportOperatorProductivityExcel({
      rows: filteredRows.value,
      averages: averages.value,
      date: localDate.value,
    });
    noticeType.value = "ok";
    notice.value = `Export Excel berhasil (${result.rowCount} baris).`;
  } catch (err) {
    noticeType.value = "error";
    notice.value = err?.message || "Gagal export Excel.";
  } finally {
    exporting.value = false;
  }
}

watch(
  localDate,
  (date) => {
    page.value = 1;
    loadOperatorProductivity(date);
  },
  { immediate: false }
);

watch([keyword, locationFilter], () => {
  page.value = 1;
});

onMounted(() => {
  loadOperatorProductivity(localDate.value);
});
</script>

<template>
  <section class="operator-prod-page">
    <section class="toolbar-card">
      <div class="toolbar-left">
        <label class="date-box">
          <span class="box-label">Tanggal</span>
          <input v-model="localDate" type="date" />
        </label>

        <label class="search-box">
          <span class="box-label">Pencarian</span>
          <input
            v-model="keyword"
            type="text"
            placeholder="Cari operator, NIK, lokasi, UUID, atau proses..."
          />
        </label>

        <label class="filter-box">
          <span class="box-label">Area</span>
          <select v-model="locationFilter">
            <option
              v-for="area in locationOptions"
              :key="area"
              :value="area"
            >
              {{ area === "ALL" ? "Semua Area" : area }}
            </option>
          </select>
        </label>
      </div>

      <button
        type="button"
        class="export-btn"
        :disabled="exporting || loading || !filteredRows.length"
        @click="handleExport"
      >
        {{ exporting ? "Exporting..." : "Export Excel" }}
      </button>
    </section>

    <p v-if="notice" class="notice" :class="noticeType">{{ notice }}</p>
    <p v-if="errorMessage" class="notice error">{{ errorMessage }}</p>

    <section class="table-card">
      <div class="table-head">
        <div>
          <h3>Produktivitas Operator</h3>
          <p>
            {{ filteredRows.length }} mesin · {{ loggedInCount }} login ·
            {{ unloggedCount }} belum login
          </p>
        </div>
        <div class="date-banner">{{ dateLabel || "-" }}</div>
      </div>

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Area</th>
              <th>Location</th>
              <th>UUID</th>
              <th>Nama Operator</th>
              <th>NIK</th>
              <th>Shift</th>
              <th>Mesin</th>
              <th class="right">Output</th>
              <th class="right">Cycle Time</th>
              <th class="right">Mesin Menyala</th>
              <th class="right">Mesin Bekerja</th>
              <th class="right">Waktu Mesin Terbuang</th>
              <th class="center">Produktivitas</th>
              <th class="right">Tunggu bahan</th>
              <th class="right">Mesin Rusak</th>
              <th class="right">Ke Toilet</th>
              <th class="right">Solat</th>
              <th class="right">Others</th>
              <th>Remarks</th>
            </tr>
          </thead>

          <tbody>
            <tr v-if="loading">
              <td colspan="19" class="empty">Memuat data operator...</td>
            </tr>

            <tr v-else-if="!filteredRows.length">
              <td colspan="19" class="empty">
                Tidak ada data mesin pada tanggal ini.
              </td>
            </tr>

            <tr
              v-for="row in pagedRows"
              :key="`${row.uuid}-${row.id}-${row.loginTime}`"
            >
              <td>{{ row.area }}</td>
              <td>{{ row.locationLabel }}</td>
              <td class="mono">{{ row.uuid }}</td>
              <td :class="{ 'not-logged': !row.loggedIn }">
                {{ row.operatorName }}
              </td>
              <td>{{ row.operatorNik || "-" }}</td>
              <td>
                <span
                  v-if="row.loggedIn"
                  class="shift-tag"
                  :class="shiftClass(row.shiftTag)"
                >
                  {{ row.shiftTag || "Normal" }}
                </span>
                <span v-else class="shift-tag shift-empty">-</span>
              </td>
              <td>{{ row.mesin }}</td>
              <td class="right">{{ row.output }}</td>
              <td class="right">{{ formatCycle(row.avgCycle) }}</td>
              <td class="right mono">{{ row.powerOnText }}</td>
              <td class="right mono">{{ row.processText }}</td>
              <td class="right mono">{{ row.lossText }}</td>
              <td class="center pct">{{ formatPct(row.productivity) }}</td>
              <td class="right mono">{{ row.tungguBahanText }}</td>
              <td class="right mono">{{ row.mesinRusakText }}</td>
              <td class="right mono">{{ row.toiletText }}</td>
              <td class="right mono">{{ row.solatText }}</td>
              <td class="right mono">{{ row.otherText }}</td>
              <td class="remarks">{{ row.remarks }}</td>
            </tr>
          </tbody>

          <tfoot v-if="!loading && filteredRows.length">
            <tr class="avg-row">
              <td><strong>AVERAGE</strong></td>
              <td></td>
              <td></td>
              <td></td>
              <td></td>
              <td></td>
              <td></td>
              <td class="right">{{ averages.output }}</td>
              <td class="right">{{ formatCycle(averages.avgCycle) }}</td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.runtimeSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.procSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.lossTimeSec) }}
              </td>
              <td class="center pct">{{ formatPct(averages.productivity) }}</td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.tungguBahanSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.mesinRusakSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.toiletSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.solatSec) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.otherSec) }}
              </td>
              <td></td>
            </tr>
          </tfoot>
        </table>
      </div>

      <div v-if="!loading && filteredRows.length" class="pagination">
        <span>Page {{ page }} / {{ totalPages }}</span>
        <div class="page-controls">
          <button type="button" :disabled="page <= 1" @click="goPage(page - 1)">
            Prev
          </button>
          <button
            v-for="pageNo in visiblePages"
            :key="pageNo"
            type="button"
            class="page-number"
            :class="{ active: page === pageNo }"
            @click="goPage(pageNo)"
          >
            {{ pageNo }}
          </button>
          <button
            type="button"
            :disabled="page >= totalPages"
            @click="goPage(page + 1)"
          >
            Next
          </button>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.operator-prod-page {
  display: grid;
  gap: 16px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.toolbar-card {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: stretch;
  padding: 16px;
  border-radius: 24px;
  background: linear-gradient(135deg, #ffffff, #f3f8ff);
  border: 1px solid #dbeafe;
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.toolbar-left {
  display: grid;
  grid-template-columns: minmax(170px, 220px) minmax(0, 1fr) minmax(140px, 180px);
  gap: 14px;
  flex: 1;
  min-width: 0;
}

.date-box,
.search-box,
.filter-box {
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  border-radius: 18px;
  background: #ffffff;
  border: 1px solid #dbeafe;
}

.box-label {
  font-size: 11px;
  font-weight: 800;
  color: #64748b;
  text-transform: uppercase;
}

input,
select {
  border: 0;
  outline: 0;
  background: transparent;
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.export-btn {
  border: 0;
  background: #16a34a;
  color: #fff;
  border-radius: 16px;
  padding: 0 18px;
  font-weight: 800;
  cursor: pointer;
  white-space: nowrap;
}

.export-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.notice {
  margin: 0;
  font-weight: 700;
}

.notice.ok {
  color: #15803d;
}

.notice.error {
  color: #dc2626;
}

.table-card {
  background: #fff;
  border: 1px solid #dbe4ef;
  border-radius: 20px;
  padding: 16px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.table-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.table-head h3 {
  margin: 0;
  font-size: 18px;
}

.table-head p {
  margin: 4px 0 0;
  color: #64748b;
  font-weight: 800;
  font-size: 13px;
}

.date-banner {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  color: #fff;
  font-weight: 800;
  font-size: 16px;
  padding: 10px 20px;
  border-radius: 8px;
  min-width: 240px;
  text-align: center;
  white-space: nowrap;
}

.table-wrap {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  overflow: auto;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
}

.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.page-controls {
  display: flex;
  gap: 6px;
}

.page-controls button {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #334155;
  border-radius: 10px;
  padding: 8px 12px;
  font-weight: 800;
  cursor: pointer;
}

.page-controls button.page-number.active {
  background: #2563eb;
  border-color: #2563eb;
  color: #fff;
}

.page-controls button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

table {
  width: 100%;
  min-width: 1680px;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: auto;
}

th,
td {
  box-sizing: border-box;
}

th {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  color: #fff;
  text-align: left;
  padding: 10px 8px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

td {
  padding: 8px;
  border-bottom: 1px solid #edf2f7;
  font-size: 12px;
  vertical-align: top;
  color: #1e293b;
}

tr:nth-child(even) td {
  background: #f8fafc;
}

.right {
  text-align: right;
}

.center {
  text-align: center;
}

.mono {
  font-family: Consolas, Menlo, monospace;
  font-variant-numeric: tabular-nums;
}

.pct {
  background: #dbeafe !important;
  color: #1e40af;
  font-weight: 800;
}

.shift-tag {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 999px;
  color: #fff;
  font-size: 11px;
  font-weight: 800;
  white-space: nowrap;
}

.shift-tag.shift-1 {
  background: #2563eb;
}

.shift-tag.shift-2 {
  background: #059669;
}

.shift-tag.shift-3 {
  background: #7c3aed;
}

.shift-tag.shift-normal {
  background: #64748b;
}

.shift-tag.shift-all {
  background: #ea580c;
}

.shift-tag.shift-empty {
  background: #e2e8f0;
  color: #64748b;
}

.not-logged {
  color: #94a3b8;
  font-style: italic;
  font-weight: 700;
}

.remarks {
  min-width: 180px;
  white-space: normal;
}

.empty {
  text-align: center;
  color: #64748b;
  font-weight: 700;
  padding: 28px 12px;
}

.avg-row td {
  background: #e2e8f0 !important;
  border-top: 2px solid #94a3b8;
  font-weight: 800;
}

.avg-row .pct {
  background: #bfdbfe !important;
}

@media (max-width: 900px) {
  .toolbar-left {
    grid-template-columns: 1fr;
  }

  .toolbar-card {
    flex-direction: column;
  }
}
</style>
