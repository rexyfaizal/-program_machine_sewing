<script setup>
import { computed, ref } from "vue";
import { useMachineUtilizationResume } from "../composables/useMachineUtilizationResume";
import { exportMachineUtilizationExcel } from "../utils/machineUtilizationResume";

const props = defineProps({ selectedDate: { type: String, default: "" } });
const {
  period, building, keyword, selectedWeekNumbers, monthLabel, weekOptions, weeks,
  filteredRows, averages, loading, errorMessage, completedDays, totalDays, lastUpdate, load,
} = useMachineUtilizationResume(props.selectedDate);
const exporting = ref(false);
const exportMessage = ref("");
const machineCount = computed(() => new Set(filteredRows.value.map((row) => row.uuid.toUpperCase())).size);
const operatorCount = computed(() => new Set(filteredRows.value
  .map((row) => row.operatorNik || row.operatorName).filter(Boolean)).size);
const progress = computed(() => totalDays.value ? completedDays.value / totalDays.value * 100 : 0);

function percent(value) {
  return typeof value === "number" ? `${value.toFixed(2)}%` : "";
}

function cellClass(value) {
  if (typeof value !== "number") return "unmeasured";
  return value >= 90 ? "good" : value >= 80 ? "normal" : "bad";
}

function selectAllWeeks() {
  selectedWeekNumbers.value = weekOptions.value.map((week) => week.weekNo);
}

function downloadExcel() {
  exporting.value = true;
  exportMessage.value = "";
  try {
    exportMachineUtilizationExcel({ rows: filteredRows.value, weeks: weeks.value,
      monthLabel: monthLabel.value, period: period.value, building: building.value });
    exportMessage.value = "Excel berhasil dibuat sesuai filter yang tampil.";
  } catch (error) {
    exportMessage.value = `Gagal membuat Excel: ${error.message}`;
  } finally {
    exporting.value = false;
  }
}
</script>

<template>
  <section class="utilization-resume">
    <div class="resume-controls">
      <div class="resume-heading">
        <div>
          <p class="eyebrow">LAPORAN MINGGUAN</p>
          <h2>Resume Utilisasi Mesin</h2>
          <p>Bandingkan utilisasi mesin per operator dan proses di {{ monthLabel || 'bulan pilihan' }}.</p>
        </div>
        <button class="export-button" type="button" :disabled="loading || exporting || !!errorMessage || !filteredRows.length" @click="downloadExcel">
          {{ exporting ? 'Membuat Excel...' : 'Export Excel' }}
        </button>
      </div>

      <div class="resume-filters">
        <label>
          <span>Bulan laporan</span>
          <input v-model="period" type="month" min="1900-01" max="2100-12" aria-label="Bulan laporan" />
        </label>
        <label>
          <span>Gedung</span>
          <select v-model="building" aria-label="Gedung">
            <option value="ALL">Semua gedung</option>
            <option value="GM1">GM1</option>
            <option value="GM2">GM2</option>
            <option value="GM3">GM3</option>
          </select>
        </label>
        <label class="search-field">
          <span>Pencarian</span>
          <input v-model="keyword" type="search" placeholder="Lokasi, UUID, nama, NIK, atau proses..." aria-label="Cari resume utilisasi" />
        </label>
        <button class="reload-button" type="button" :disabled="loading" @click="load({ refresh: true })">Muat ulang</button>
      </div>

      <fieldset class="week-filter">
        <legend>Minggu yang ditampilkan</legend>
        <label v-for="week in weekOptions" :key="week.weekNo" :class="{ selected: selectedWeekNumbers.includes(week.weekNo) }">
          <input v-model="selectedWeekNumbers" type="checkbox" :value="week.weekNo"
            :disabled="selectedWeekNumbers.length === 1 && selectedWeekNumbers.includes(week.weekNo)" />
          <span><strong>W{{ week.weekNo }}</strong><small>{{ week.rangeLabel }}</small></span>
        </label>
        <button type="button" @click="selectAllWeeks">Semua minggu</button>
      </fieldset>
    </div>

    <div v-if="loading" class="load-status" role="status" aria-live="polite">
      <span>Memuat laporan {{ completedDays }} / {{ totalDays }} hari kerja...</span>
      <progress :value="progress" max="100" aria-label="Progres laporan"></progress>
    </div>
    <div v-if="errorMessage" class="resume-error" role="alert">{{ errorMessage }}</div>
    <div v-if="exportMessage" class="export-notice" role="status">{{ exportMessage }}</div>

    <div class="resume-summary">
      <div><span>Mesin</span><strong>{{ machineCount }}</strong></div>
      <div><span>Operator</span><strong>{{ operatorCount }}</strong></div>
      <div><span>Baris laporan</span><strong>{{ filteredRows.length }}</strong></div>
      <p v-if="lastUpdate">Diperbarui {{ lastUpdate }}</p>
    </div>

    <div class="resume-table-wrap" :aria-busy="loading">
      <table class="resume-table">
        <caption>Resume Utilisasi Mesin · {{ monthLabel }} · {{ building === 'ALL' ? 'Semua gedung' : building }}</caption>
        <thead>
          <tr>
            <th scope="col">Area</th>
            <th scope="col">Location</th>
            <th scope="col">UUID</th>
            <th scope="col">NAMA</th>
            <th scope="col">NIK</th>
            <th scope="col">PROSES</th>
            <th v-for="week in weeks" :key="week.weekNo" scope="col" class="week-heading">
              AVG W{{ week.weekNo }}<small>{{ week.rangeLabel }}</small>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filteredRows" :key="row.key">
            <td>{{ row.area === '-' ? '' : row.area }}</td>
            <td>{{ row.location }}</td>
            <td class="uuid-cell">{{ row.uuid }}</td>
            <td>{{ row.operatorName }}</td>
            <td class="nik-cell">{{ row.operatorNik }}</td>
            <td>{{ row.processName }}</td>
            <td v-for="week in weeks" :key="week.weekNo" class="percent-cell"
              :class="cellClass(row.weeks[week.weekNo]?.value)"
              :title="row.weeks[week.weekNo]?.days ? `Rata-rata ${row.weeks[week.weekNo].days} hari dengan durasi nyala` : 'Belum ada durasi nyala pada minggu ini'">
              {{ percent(row.weeks[week.weekNo]?.value) }}
            </td>
          </tr>
          <tr v-if="!filteredRows.length">
            <td class="empty-state" :colspan="6 + weeks.length">
              {{ loading ? 'Mengambil data mesin dan operator...' : errorMessage ? 'Laporan belum dapat ditampilkan. Muat ulang untuk mencoba kembali.' : 'Belum ada data untuk bulan, minggu, dan gedung yang dipilih.' }}
            </td>
          </tr>
        </tbody>
        <tfoot v-if="filteredRows.length">
          <tr>
            <th colspan="6" scope="row">Average</th>
            <td v-for="week in weeks" :key="week.weekNo" class="percent-cell">{{ percent(averages[week.weekNo]) }}</td>
          </tr>
        </tfoot>
      </table>
    </div>
    <p class="calculation-note">
      AVG mingguan adalah rata-rata utilisasi harian (waktu proses ÷ durasi mesin nyala).
      Baris operator memakai durasi selama sesi login; mesin tanpa operator memakai pengukuran harian.
      Sel kosong berarti belum ada durasi nyala; 0% berarti mesin menyala tanpa proses. Hari Minggu tidak dihitung.
    </p>
  </section>
</template>

<style scoped>
.utilization-resume { display: grid; gap: 16px; }
.resume-controls { background: #fff; border: 1px solid #dbe7fb; border-radius: 20px; padding: 24px; box-shadow: 0 8px 24px #234d8210; }
.resume-heading { display: flex; justify-content: space-between; align-items: center; gap: 20px; margin-bottom: 22px; }
.eyebrow { margin: 0 0 6px; color: #2563eb; font-size: 11px; font-weight: 800; letter-spacing: .12em; }
h2 { font-size: 23px; margin: 0 0 6px; }
.resume-heading p:not(.eyebrow) { margin: 0; color: #64748b; font-size: 13px; }
button, input, select { font: inherit; }
button { cursor: pointer; }
button:disabled { cursor: default; opacity: .5; }
.export-button, .reload-button { border: 0; border-radius: 10px; padding: 12px 20px; font-weight: 750; white-space: nowrap; }
.export-button { background: #15803d; color: white; }
.reload-button { background: #eaf1ff; color: #1d4ed8; }
.resume-filters { display: flex; flex-wrap: wrap; align-items: end; gap: 14px; }
.resume-filters label { display: grid; gap: 7px; min-width: 170px; }
.resume-filters label > span, legend { font-size: 12px; font-weight: 750; color: #475569; }
.resume-filters input, .resume-filters select { height: 43px; background: white; border: 1px solid #cbd5e1; border-radius: 9px; padding: 8px 12px; color: #0f172a; min-width: 0; }
.resume-filters .search-field { flex: 1; min-width: 240px; }
input:focus-visible, select:focus-visible, button:focus-visible { outline: 2px solid #2563eb; outline-offset: 2px; }
.week-filter { display: flex; flex-wrap: wrap; gap: 10px; padding: 10px 0 0; margin: 20px 0 0; border: 0; }
.week-filter legend { padding: 0; margin-bottom: 3px; }
.week-filter label { display: flex; align-items: center; gap: 9px; padding: 9px 12px; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer; }
.week-filter label.selected { background: #eff6ff; border-color: #93c5fd; }
.week-filter input { accent-color: #2563eb; }
.week-filter strong { font-size: 13px; }
.week-filter small { display: block; color: #64748b; font-size: 11px; margin-top: 3px; }
.week-filter button { border: 0; background: transparent; color: #2563eb; padding: 9px; font-size: 12px; font-weight: 700; }
.resume-summary { display: flex; flex-wrap: wrap; align-items: center; gap: 14px 32px; padding: 0 4px; }
.resume-summary div { display: flex; align-items: baseline; gap: 9px; }
.resume-summary span { font-size: 12px; color: #64748b; }
.resume-summary strong { font-size: 23px; }
.resume-summary p { margin: 0 0 0 auto; font-size: 11px; color: #64748b; }
.resume-table-wrap { overflow: auto; max-height: 70vh; border: 1px solid #b8c8df; border-radius: 12px; background: white; }
.resume-table { width: 100%; min-width: 1120px; border-spacing: 0; border-collapse: separate; font-size: 12px; }
caption { text-align: left; padding: 15px; font-weight: 750; color: #334155; background: #f8fafc; }
th, td { border-right: 1px solid #cbd5e1; border-bottom: 1px solid #cbd5e1; padding: 9px 10px; vertical-align: middle; }
thead th { position: sticky; top: 0; z-index: 1; text-align: center; background: #dbeafe; color: #142e52; font-size: 11px; letter-spacing: .025em; }
thead th:nth-child(2) { min-width: 175px; }
thead th:nth-child(4) { min-width: 170px; }
thead th:nth-child(6) { min-width: 210px; }
.week-heading { min-width: 125px; }
.week-heading small { display: block; margin-top: 5px; font-weight: 500; font-size: 10px; letter-spacing: 0; }
.uuid-cell, .nik-cell { white-space: nowrap; font-variant-numeric: tabular-nums; }
.uuid-cell { font-size: 11px; }
.percent-cell { text-align: center; font-variant-numeric: tabular-nums; font-weight: 650; }
.percent-cell.bad { background: #f5b9b3; color: #702b2b; }
.percent-cell.normal { background: #fef3c7; color: #854d0e; }
.percent-cell.good { background: #dcfce7; color: #166534; }
.percent-cell.unmeasured { background: #f8e7e5; }
tbody tr:hover td:not(.percent-cell) { background: #f8fbff; }
tfoot th, tfoot td { background: #fef08a; font-weight: 800; padding-top: 13px; padding-bottom: 13px; }
tfoot th { text-align: right; }
.empty-state { text-align: center; color: #64748b; padding: 45px 16px; }
.calculation-note { margin: 0 4px; color: #64748b; font-size: 12px; line-height: 1.8; }
.load-status { display: flex; align-items: center; gap: 20px; padding: 14px 18px; border-radius: 10px; background: #dbeafe; color: #1e40af; font-size: 13px; }
.load-status progress { flex: 1; height: 8px; accent-color: #2563eb; }
.resume-error, .export-notice { padding: 14px 18px; border-radius: 10px; font-size: 13px; }
.resume-error { background: #fee2e2; color: #991b1b; }
.export-notice { background: #dcfce7; color: #166534; }
@media (max-width: 700px) {
  .resume-controls { padding: 18px 14px; }
  .resume-heading { align-items: start; flex-direction: column; }
  .resume-filters label { flex: 1; min-width: 130px; }
  .resume-filters .search-field { flex-basis: 100%; }
  .resume-summary { gap: 12px 20px; }
  .resume-summary p { width: 100%; margin-left: 0; }
  .load-status { flex-direction: column; align-items: stretch; gap: 10px; }
}
</style>
