<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useOperatorProductivity } from "../composables/useOperatorProductivity";
import { useOperatorCtImport } from "../composables/useOperatorCtImport";
import { useOperatorOutputTargetImport } from "../composables/useOperatorOutputTargetImport";
import { exportOperatorProductivityExcel } from "../utils/operatorProductivityExcel";
import { formatDurationHHMMSS } from "../utils/format";
import { formatCtNumber, formatProduktivitasCtPct } from "../utils/operatorCt";
import { formatOutputTarget } from "../utils/operatorOutputTarget";
import OperatorCtImportPanel from "../components/operator/OperatorCtImportPanel.vue";
import OperatorOutputTargetImportPanel from "../components/operator/OperatorOutputTargetImportPanel.vue";
import { getInitialAdminMode } from "../utils/adminMode";

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
const isAdmin = ref(false);

const {
  loading,
  errorMessage,
  keyword,
  locationFilter,
  locationOptions,
  filteredRows,
  averages,
  machineCount,
  loggedInCount,
  unloggedCount,
  loadOperatorProductivity,
} = useOperatorProductivity();

const {
  importing: ctImporting,
  importInputKey,
  importFileName,
  importPreviewRows,
  importPreviewDisplayRows,
  importDuplicateRows,
  importErrorRows,
  importStats,
  importErrorMessage,
  importSuccessMessage,
  resetImport,
  handleImportFileChange,
  submitImportExcel,
  cleanupOperatorCtImport,
} = useOperatorCtImport({
  isAdmin,
  onImported: async () => {
    await loadOperatorProductivity(localDate.value);
  },
});

const {
  importing: targetImporting,
  importInputKey: targetImportInputKey,
  importFileName: targetImportFileName,
  importPreviewRows: targetImportPreviewRows,
  importPreviewDisplayRows: targetImportPreviewDisplayRows,
  importDuplicateRows: targetImportDuplicateRows,
  importErrorRows: targetImportErrorRows,
  importStats: targetImportStats,
  importErrorMessage: targetImportErrorMessage,
  importSuccessMessage: targetImportSuccessMessage,
  resetImport: resetTargetImport,
  handleImportFileChange: handleTargetImportFileChange,
  submitImportExcel: submitTargetImportExcel,
  cleanupOperatorOutputTargetImport,
} = useOperatorOutputTargetImport({
  isAdmin,
  selectedDate: () => localDate.value,
  onImported: async () => {
    await loadOperatorProductivity(localDate.value);
  },
});

const localDate = computed({
  get: () => props.selectedDate,
  set: (value) => emit("update:selectedDate", value),
});

const page = ref(1);
const pageSize = 20;

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

function formatPct(value) {
  return `${Number(value || 0).toFixed(2)}%`;
}

function formatDisplayName(value) {
  const text = String(value || "").trim();
  if (!text) return "-";
  if (text.toLowerCase() === "not logged in") return text;

  const letters = text.replace(/[^A-Za-z]/g, "");
  if (letters && letters === letters.toUpperCase()) {
    return text
      .toLowerCase()
      .replace(/\b\w/g, (char) => char.toUpperCase());
  }

  return text;
}

function parseNotePctValue(value) {
  const text = String(value || "")
    .trim()
    .replace(/%/g, "");
  const num = Number(text);
  return Number.isFinite(num) ? num : 0;
}

function formatNotePct(value) {
  if (parseNotePctValue(value) > 100) {
    return "> 100%";
  }
  return String(value || "").trim() || "-";
}

const notePctAnomalyTitle =
  "Anomali: durasi note melebihi Mesin Menyala (note mungkin masih terbuka atau mesin sudah mati saat note berjalan).";

function isNotePctAnomaly(value) {
  return parseNotePctValue(value) > 100;
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

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());
  loadOperatorProductivity(localDate.value);
});

onBeforeUnmount(() => {
  cleanupOperatorCtImport();
  cleanupOperatorOutputTargetImport();
});
</script>

<template>
  <section class="operator-prod-page">
    <section v-if="isAdmin" class="ie-upload-group">
      <OperatorCtImportPanel
        embedded
        :is-admin="isAdmin"
        :template-rows="filteredRows"
        :location-filter="locationFilter"
        :importing="ctImporting"
        :import-input-key="importInputKey"
        :import-file-name="importFileName"
        :import-preview-rows="importPreviewRows"
        :import-preview-display-rows="importPreviewDisplayRows"
        :import-duplicate-rows="importDuplicateRows"
        :import-error-rows="importErrorRows"
        :import-stats="importStats"
        :import-error-message="importErrorMessage"
        :import-success-message="importSuccessMessage"
        @reset-import="resetImport"
        @file-change="handleImportFileChange"
        @submit-import="submitImportExcel"
      />

      <OperatorOutputTargetImportPanel
        embedded
        :is-admin="isAdmin"
        :selected-date="localDate"
        :template-rows="filteredRows"
        :location-filter="locationFilter"
        :importing="targetImporting"
        :import-input-key="targetImportInputKey"
        :import-file-name="targetImportFileName"
        :import-preview-rows="targetImportPreviewRows"
        :import-preview-display-rows="targetImportPreviewDisplayRows"
        :import-duplicate-rows="targetImportDuplicateRows"
        :import-error-rows="targetImportErrorRows"
        :import-stats="targetImportStats"
        :import-error-message="targetImportErrorMessage"
        :import-success-message="targetImportSuccessMessage"
        @reset-import="resetTargetImport"
        @file-change="handleTargetImportFileChange"
        @submit-import="submitTargetImportExcel"
      />
    </section>

    <section class="table-card">
      <div class="table-card-head">
        <div class="table-title-block">
          <h3>Produktivitas Operator</h3>
          <p class="table-meta">
            {{ machineCount }} mesin · {{ loggedInCount }} sesi login ·
            {{ unloggedCount }} belum login
          </p>
          <p class="table-legend">
            <span class="legend-item legend-ie">Upload IE</span>
            <span class="legend-item legend-prod">Produktivitas sistem</span>
          </p>
        </div>

        <div class="table-toolbar">
          <label class="filter-field filter-date">
            <span>Tanggal</span>
            <input v-model="localDate" type="date" />
          </label>

          <label class="filter-field filter-search">
            <span>Cari</span>
            <input
              v-model="keyword"
              type="text"
              placeholder="Operator, NIK, lokasi, UUID..."
            />
          </label>

          <div class="filter-field filter-area-segment">
            <span>Area</span>
            <div class="area-switch">
              <button
                v-for="area in locationOptions"
                :key="area"
                type="button"
                class="area-btn"
                :class="{ active: locationFilter === area }"
                @click="locationFilter = area"
              >
                {{ area === "ALL" ? "Semua" : area }}
              </button>
            </div>
          </div>

          <button
            type="button"
            class="export-btn"
            :disabled="exporting || loading || !filteredRows.length"
            @click="handleExport"
          >
            {{ exporting ? "Export..." : "Export Excel" }}
          </button>
        </div>
      </div>

      <p v-if="notice" class="notice inline-notice" :class="noticeType">
        {{ notice }}
      </p>
      <p v-if="errorMessage" class="notice error inline-notice">{{ errorMessage }}</p>

      <div class="table-wrap">
        <table class="operator-table">
          <thead>
            <tr>
              <th class="th-left th-wide freeze-col freeze-col-1">Area</th>
              <th class="th-left th-location freeze-col freeze-col-2">Location</th>
              <th class="th-left th-uuid freeze-col freeze-col-3">UUID</th>
              <th class="th-left th-operator freeze-col freeze-col-4">Nama Operator</th>
              <th class="th-narrow freeze-col freeze-col-5 freeze-col-last">NIK</th>
              <th class="th-narrow">Shift</th>
              <th class="th-left th-mesin">Mesin</th>
              <th class="th-narrow">Style</th>
              <th class="th-num">Output</th>
              <th class="th-num col-ie">Output Targetan</th>
              <th class="th-num col-ie">CT SUM</th>
              <th class="th-num col-ie">CT STD</th>
              <th class="th-num col-ie">CT</th>
              <th class="th-num col-ie">Kap/Jam</th>
              <th class="th-num">Mesin Menyala</th>
              <th class="th-num">Mesin Bekerja</th>
              <th class="th-num">Waktu Mesin Terbuang</th>
              <th class="th-num">Utilitas Mesin</th>
              <th class="th-num">Produktivitas CT</th>
              <th class="th-num">Produktivitas CT Targetan</th>
              <th class="th-num">Tunggu bahan</th>
              <th class="th-pct" title="Tunggu bahan">%</th>
              <th class="th-num">Mesin Rusak</th>
              <th class="th-pct" title="Mesin Rusak">%</th>
              <th class="th-num">Ke Toilet</th>
              <th class="th-pct" title="Ke Toilet">%</th>
              <th class="th-num">Solat</th>
              <th class="th-pct" title="Solat">%</th>
              <th class="th-num">Others</th>
              <th class="th-pct" title="Others">%</th>
              <th class="th-left th-remarks">Remarks</th>
            </tr>
          </thead>

          <tbody>
            <tr v-if="loading">
              <td colspan="31" class="empty">Memuat data operator...</td>
            </tr>

            <tr v-else-if="!filteredRows.length">
              <td colspan="31" class="empty">
                Tidak ada data mesin pada tanggal ini.
              </td>
            </tr>

            <tr
              v-for="row in pagedRows"
              :key="`${row.uuid}-${row.id}-${row.loginTime}`"
            >
              <td class="col-area freeze-col freeze-col-1">{{ row.area }}</td>
              <td class="col-location cell-wrap freeze-col freeze-col-2">
                {{ row.locationLabel }}
              </td>
              <td class="mono col-uuid cell-wrap freeze-col freeze-col-3">
                {{ row.uuid }}
              </td>
              <td
                class="col-operator cell-wrap freeze-col freeze-col-4"
                :class="{ 'not-logged': !row.loggedIn }"
              >
                {{ formatDisplayName(row.operatorName) }}
              </td>
              <td class="col-nik freeze-col freeze-col-5 freeze-col-last">
                {{ row.operatorNik || "-" }}
              </td>
              <td class="col-shift">
                <span
                  v-if="row.loggedIn"
                  class="shift-tag"
                  :class="shiftClass(row.shiftTag)"
                >
                  {{ row.shiftTag || "Normal" }}
                </span>
                <span v-else class="shift-tag shift-empty">-</span>
              </td>
              <td class="col-mesin cell-wrap">{{ row.mesin }}</td>
              <td class="col-style">{{ row.style }}</td>
              <td class="right">{{ row.output }}</td>
              <td class="right col-ie">{{ formatOutputTarget(row.outputTarget) }}</td>
              <td class="right col-ie">{{ formatCtNumber(row.ctSum) }}</td>
              <td class="right col-ie">{{ formatCtNumber(row.ctStd) }}</td>
              <td class="right col-ie">{{ formatCtNumber(row.ctValue) }}</td>
              <td class="right col-ie">{{ row.kapPerJam || "-" }}</td>
              <td class="right mono">{{ row.powerOnText }}</td>
              <td class="right mono">{{ row.processText }}</td>
              <td class="right mono">{{ row.lossText }}</td>
              <td class="center pct">{{ formatPct(row.productivity) }}</td>
              <td class="center pct">
                {{ formatProduktivitasCtPct(row.produktivitasCt) }}
              </td>
              <td class="center pct">
                {{ formatProduktivitasCtPct(row.produktivitasCtTargetan) }}
              </td>
              <td class="right mono">{{ row.tungguBahanText }}</td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(row.tungguBahanPct) }"
                :title="
                  isNotePctAnomaly(row.tungguBahanPct) ? notePctAnomalyTitle : ''
                "
              >
                {{ formatNotePct(row.tungguBahanPct) }}
              </td>
              <td class="right mono">{{ row.mesinRusakText }}</td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(row.mesinRusakPct) }"
                :title="
                  isNotePctAnomaly(row.mesinRusakPct) ? notePctAnomalyTitle : ''
                "
              >
                {{ formatNotePct(row.mesinRusakPct) }}
              </td>
              <td class="right mono">{{ row.toiletText }}</td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(row.toiletPct) }"
                :title="isNotePctAnomaly(row.toiletPct) ? notePctAnomalyTitle : ''"
              >
                {{ formatNotePct(row.toiletPct) }}
              </td>
              <td class="right mono">{{ row.solatText }}</td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(row.solatPct) }"
                :title="isNotePctAnomaly(row.solatPct) ? notePctAnomalyTitle : ''"
              >
                {{ formatNotePct(row.solatPct) }}
              </td>
              <td class="right mono">{{ row.otherText }}</td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(row.otherPct) }"
                :title="isNotePctAnomaly(row.otherPct) ? notePctAnomalyTitle : ''"
              >
                {{ formatNotePct(row.otherPct) }}
              </td>
              <td class="remarks cell-wrap">{{ row.remarks }}</td>
            </tr>
          </tbody>

          <tfoot v-if="!loading && filteredRows.length">
            <tr class="avg-row">
              <td class="freeze-col freeze-col-1"><strong>AVERAGE</strong></td>
              <td class="freeze-col freeze-col-2"></td>
              <td class="freeze-col freeze-col-3"></td>
              <td class="freeze-col freeze-col-4"></td>
              <td class="freeze-col freeze-col-5 freeze-col-last"></td>
              <td></td>
              <td></td>
              <td></td>
              <td class="right">{{ averages.output }}</td>
              <td class="right col-ie">{{ formatOutputTarget(averages.outputTarget) }}</td>
              <td class="col-ie"></td>
              <td class="col-ie"></td>
              <td class="col-ie"></td>
              <td class="col-ie"></td>
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
              <td class="center pct">
                {{ formatProduktivitasCtPct(averages.produktivitasCt) }}
              </td>
              <td class="center pct">
                {{ formatProduktivitasCtPct(averages.produktivitasCtTargetan) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.tungguBahanSec) }}
              </td>
              <td
                class="center note-pct"
                :class="{
                  'note-pct-anomaly': isNotePctAnomaly(averages.tungguBahanPct),
                }"
                :title="
                  isNotePctAnomaly(averages.tungguBahanPct)
                    ? notePctAnomalyTitle
                    : ''
                "
              >
                {{ formatNotePct(averages.tungguBahanPct) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.mesinRusakSec) }}
              </td>
              <td
                class="center note-pct"
                :class="{
                  'note-pct-anomaly': isNotePctAnomaly(averages.mesinRusakPct),
                }"
                :title="
                  isNotePctAnomaly(averages.mesinRusakPct)
                    ? notePctAnomalyTitle
                    : ''
                "
              >
                {{ formatNotePct(averages.mesinRusakPct) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.toiletSec) }}
              </td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(averages.toiletPct) }"
                :title="
                  isNotePctAnomaly(averages.toiletPct) ? notePctAnomalyTitle : ''
                "
              >
                {{ formatNotePct(averages.toiletPct) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.solatSec) }}
              </td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(averages.solatPct) }"
                :title="
                  isNotePctAnomaly(averages.solatPct) ? notePctAnomalyTitle : ''
                "
              >
                {{ formatNotePct(averages.solatPct) }}
              </td>
              <td class="right mono">
                {{ formatDurationHHMMSS(averages.otherSec) }}
              </td>
              <td
                class="center note-pct"
                :class="{ 'note-pct-anomaly': isNotePctAnomaly(averages.otherPct) }"
                :title="
                  isNotePctAnomaly(averages.otherPct) ? notePctAnomalyTitle : ''
                "
              >
                {{ formatNotePct(averages.otherPct) }}
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

.ie-upload-group {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  padding: 12px 14px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.ie-upload-group > :deep(.import-panel:not(:last-child)) {
  padding-right: 14px;
  border-right: 1px solid #e2e8f0;
}

.ie-upload-group > :deep(.import-panel:not(:first-child)) {
  padding-left: 14px;
}

@media (max-width: 1100px) {
  .ie-upload-group {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .ie-upload-group > :deep(.import-panel:not(:last-child)) {
    padding-right: 0;
    padding-bottom: 12px;
    border-right: 0;
    border-bottom: 1px solid #e2e8f0;
  }

  .ie-upload-group > :deep(.import-panel:not(:first-child)) {
    padding-left: 0;
    padding-top: 0;
  }
}

.table-card-head {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 12px;
}

.table-title-block h3 {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.table-meta {
  margin: 4px 0 0;
  color: #64748b;
  font-weight: 700;
  font-size: 13px;
}

.table-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 8px;
}

.filter-field {
  display: grid;
  gap: 4px;
  padding: 8px 10px;
  border: 1px solid #dbe4ef;
  border-radius: 10px;
  background: #f8fafc;
  min-width: 0;
}

.filter-field span {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  color: #64748b;
  letter-spacing: 0.03em;
}

.filter-field input,
.filter-field select {
  border: 0;
  outline: 0;
  background: transparent;
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
  min-width: 0;
}

.filter-date {
  flex: 0 0 150px;
}

.filter-search {
  flex: 1 1 220px;
}

.filter-area-segment {
  flex: 0 1 auto;
  min-width: 220px;
  background: #fff;
  border-color: #dbeafe;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.035);
}

.area-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.area-btn {
  border: 0;
  border-radius: 8px;
  background: #f1f5f9;
  color: #475569;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  padding: 6px 12px;
  white-space: nowrap;
  line-height: 1.2;
}

.area-btn:hover {
  background: #e2e8f0;
}

.area-btn.active {
  background: #2563eb;
  color: #fff;
}

.table-toolbar .export-btn {
  flex: 0 0 auto;
  height: 56px;
  padding: 0 16px;
  border: 0;
  border-radius: 10px;
  background: #16a34a;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  white-space: nowrap;
}

.table-toolbar .export-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.inline-notice {
  margin: 0 0 10px;
  font-size: 13px;
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

.table-legend {
  display: flex;
  gap: 12px;
  margin-top: 8px !important;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 700;
  color: #475569;
}

.legend-item::before {
  content: "";
  width: 14px;
  height: 14px;
  border-radius: 4px;
  border: 1px solid rgba(15, 23, 42, 0.08);
}

.legend-ie::before {
  background: #fff7ed;
}

.legend-prod::before {
  background: #dbeafe;
}

@media (max-width: 900px) {
  .table-toolbar {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .filter-search {
    grid-column: 1 / -1;
  }

  .filter-area-segment {
    grid-column: 1 / -1;
  }

  .table-toolbar .export-btn {
    grid-column: 1 / -1;
    width: 100%;
    height: 44px;
  }
}

.table-wrap {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  overflow-x: auto;
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

table.operator-table {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: auto;
  --freeze-w-1: 56px;
  --freeze-w-2: 148px;
  --freeze-w-3: 132px;
  --freeze-w-4: 168px;
  --freeze-w-5: 84px;
}

th,
td {
  box-sizing: border-box;
}

.operator-table thead th.th-wide.freeze-col-1,
.operator-table td.freeze-col-1 {
  min-width: var(--freeze-w-1);
  max-width: var(--freeze-w-1);
  width: var(--freeze-w-1);
}

.operator-table thead th.th-location.freeze-col-2,
.operator-table td.freeze-col-2 {
  min-width: var(--freeze-w-2);
  max-width: var(--freeze-w-2);
  width: var(--freeze-w-2);
}

.operator-table thead th.th-uuid.freeze-col-3,
.operator-table td.freeze-col-3 {
  min-width: var(--freeze-w-3);
  max-width: var(--freeze-w-3);
  width: var(--freeze-w-3);
}

.operator-table thead th.th-operator.freeze-col-4,
.operator-table td.freeze-col-4 {
  min-width: var(--freeze-w-4);
  max-width: var(--freeze-w-4);
  width: var(--freeze-w-4);
}

.operator-table thead th.freeze-col-5,
.operator-table td.freeze-col-5 {
  min-width: var(--freeze-w-5);
  max-width: var(--freeze-w-5);
  width: var(--freeze-w-5);
}

.operator-table .freeze-col {
  position: sticky;
  background-clip: padding-box;
}

.operator-table thead th.freeze-col {
  z-index: 4;
}

.operator-table tbody td.freeze-col,
.operator-table tfoot td.freeze-col {
  z-index: 2;
  background: #fff;
}

.operator-table tbody tr:nth-child(even) td.freeze-col {
  background: #f8fafc;
}

.operator-table tfoot td.freeze-col {
  background: #e2e8f0 !important;
}

.operator-table .freeze-col-1 {
  left: 0;
}

.operator-table .freeze-col-2 {
  left: var(--freeze-w-1);
}

.operator-table .freeze-col-3 {
  left: calc(var(--freeze-w-1) + var(--freeze-w-2));
}

.operator-table .freeze-col-4 {
  left: calc(var(--freeze-w-1) + var(--freeze-w-2) + var(--freeze-w-3));
}

.operator-table .freeze-col-5 {
  left: calc(var(--freeze-w-1) + var(--freeze-w-2) + var(--freeze-w-3) + var(--freeze-w-4));
}

.operator-table .freeze-col-last {
  box-shadow: 4px 0 8px -4px rgba(15, 23, 42, 0.18);
}

.operator-table thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  color: #fff;
  padding: 8px 6px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  line-height: 1.25;
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
  vertical-align: middle;
  text-align: center;
  border-right: 1px solid rgba(255, 255, 255, 0.12);
}

.operator-table thead th:last-child {
  border-right: 0;
}

.operator-table thead th.th-left {
  text-align: left;
}

.operator-table thead th.th-narrow {
  width: 56px;
  min-width: 56px;
}

.operator-table thead th.th-pct {
  width: 42px;
  min-width: 42px;
}

.operator-table thead th.th-num {
  width: 78px;
  min-width: 72px;
}

.operator-table thead th.th-wide {
  min-width: 56px;
}

.operator-table thead th.th-location {
  min-width: 148px;
}

.operator-table thead th.th-uuid {
  min-width: 132px;
}

.operator-table thead th.th-operator {
  min-width: 168px;
}

.operator-table thead th.th-mesin {
  min-width: 220px;
}

.operator-table thead th.th-remarks {
  min-width: 140px;
}

td {
  padding: 10px 10px;
  border-bottom: 1px solid #e8eef4;
  font-size: 12px;
  vertical-align: top;
  color: #1e293b;
  line-height: 1.45;
}

.cell-wrap {
  white-space: normal;
  word-break: break-word;
  overflow-wrap: anywhere;
  hyphens: auto;
}

.col-area {
  font-weight: 700;
  color: #475569;
  white-space: nowrap;
  vertical-align: middle;
}

.col-nik,
.col-style {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  vertical-align: middle;
}

.col-uuid {
  font-size: 11px;
  color: #475569;
  word-break: break-all;
}

.col-operator {
  font-weight: 600;
  font-size: 12px;
}

.operator-table td.col-location {
  font-size: 12px;
  color: #334155;
}

.operator-table td.col-mesin {
  font-size: 12px;
  color: #334155;
}

.right,
.center,
.col-nik,
.col-style,
.col-shift,
td.col-ie {
  white-space: nowrap;
  vertical-align: middle;
}

.col-shift {
  text-align: center;
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

th.col-ie {
  background: linear-gradient(135deg, #ea580c 0%, #c2410c 100%);
}

td.col-ie {
  background: #fff7ed !important;
  color: #7c2d12;
  font-weight: 700;
}

tr:nth-child(even) td.col-ie {
  background: #ffedd5 !important;
}

.avg-row td.col-ie {
  background: #fed7aa !important;
}

.pct {
  background: #dbeafe !important;
  color: #1e40af;
  font-weight: 800;
}

.note-pct {
  font-weight: 800;
  color: #334155;
  white-space: nowrap;
}

.note-pct.note-pct-anomaly {
  background: #fee2e2 !important;
  color: #b91c1c;
  font-weight: 900;
  box-shadow: inset 0 0 0 1px #fca5a5;
}

.avg-row .note-pct.note-pct-anomaly {
  background: #fecaca !important;
}

.shift-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  padding: 2px 7px;
  border-radius: 6px;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
  line-height: 1.2;
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
  font-weight: 600;
}

.remarks {
  min-width: 140px;
  max-width: 220px;
  font-size: 12px;
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
</style>
