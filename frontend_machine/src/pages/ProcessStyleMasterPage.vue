<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

import ProcessStyleGccSyncPanel from "../components/process-style/ProcessStyleGccSyncPanel.vue";
import ProcessStyleMachineIdImportPanel from "../components/process-style/ProcessStyleMachineIdImportPanel.vue";
import ProcessStyleTable from "../components/process-style/ProcessStyleTable.vue";

import {
  getMachineSettings,
  importMachineIds,
  syncProcessStylesFromGcc,
} from "../api/machineApi";
import { useProcessStyleGccLive } from "../composables/useProcessStyleGccLive";
import { getInitialAdminMode } from "../utils/adminMode";
import { parseMachineIdImportFile } from "../utils/machineIdImportExcel";

import "../assets/process-style-master.css";

const isAdmin = ref(false);
const syncing = ref(false);
const syncErrorMessage = ref("");
const syncSuccessMessage = ref("");
const syncResult = ref(null);

const machineIdTemplateRows = ref([]);
const machineIdImporting = ref(false);
const machineIdImportInputKey = ref(0);
const machineIdImportFileName = ref("");
const machineIdImportPreviewRows = ref([]);
const machineIdImportStats = ref({
  totalExcelRows: 0,
  readyRows: 0,
  skippedEmpty: 0,
  skippedDuplicate: 0,
  skippedInvalid: 0,
});
const machineIdImportErrorMessage = ref("");
const machineIdImportSuccessMessage = ref("");
const machineIdImportResult = ref(null);

let syncSuccessTimer = null;
let machineIdSuccessTimer = null;

const {
  loading,
  errorMessage,
  keyword,
  rows,
  source,
  sourceMessage,
  currentPage,
  totalRows,
  totalPages,
  startIndex,
  endIndex,
  visiblePages,
  loadRows,
  handleSearch,
  scheduleSearch,
  prevPage,
  nextPage,
  goToPage,
  cleanup,
} = useProcessStyleGccLive();

async function loadMachineIdTemplateRows() {
  try {
    const list = await getMachineSettings();
    machineIdTemplateRows.value = Array.isArray(list) ? list : [];
  } catch (err) {
    console.warn("Gagal load template Machine ID:", err);
    machineIdTemplateRows.value = [];
  }
}

function resetMachineIdImport() {
  machineIdImportInputKey.value += 1;
  machineIdImportFileName.value = "";
  machineIdImportPreviewRows.value = [];
  machineIdImportStats.value = {
    totalExcelRows: 0,
    readyRows: 0,
    skippedEmpty: 0,
    skippedDuplicate: 0,
    skippedInvalid: 0,
  };
  machineIdImportErrorMessage.value = "";
  machineIdImportSuccessMessage.value = "";
  machineIdImportResult.value = null;
}

async function handleMachineIdFileChange(event) {
  const file = event?.target?.files?.[0];
  machineIdImportErrorMessage.value = "";
  machineIdImportSuccessMessage.value = "";
  machineIdImportResult.value = null;

  if (!file) {
    resetMachineIdImport();
    return;
  }

  machineIdImportFileName.value = file.name;

  try {
    const parsed = await parseMachineIdImportFile(file);
    machineIdImportPreviewRows.value = parsed.rows || [];
    machineIdImportStats.value = parsed.stats || machineIdImportStats.value;

    if (!machineIdImportPreviewRows.value.length) {
      machineIdImportErrorMessage.value =
        "Tidak ada baris valid. Pastikan kolom UUID dan Machine ID terisi.";
    }
  } catch (err) {
    machineIdImportPreviewRows.value = [];
    machineIdImportErrorMessage.value =
      err?.message || "Gagal membaca file Excel Machine ID.";
  }
}

async function handleMachineIdImport() {
  if (!isAdmin.value || machineIdImporting.value) return;
  if (!machineIdImportPreviewRows.value.length) {
    machineIdImportErrorMessage.value = "Belum ada baris siap import.";
    return;
  }

  machineIdImporting.value = true;
  machineIdImportErrorMessage.value = "";
  machineIdImportSuccessMessage.value = "";
  machineIdImportResult.value = null;

  try {
    const result = await importMachineIds(machineIdImportPreviewRows.value);
    machineIdImportResult.value = result || null;
    machineIdImportSuccessMessage.value =
      result?.message ||
      `Import selesai. ${Number(result?.upserted || 0)} Machine ID tersimpan.`;

    if (machineIdSuccessTimer) clearTimeout(machineIdSuccessTimer);
    machineIdSuccessTimer = setTimeout(() => {
      machineIdImportSuccessMessage.value = "";
    }, 6000);

    await loadMachineIdTemplateRows();
  } catch (err) {
    machineIdImportErrorMessage.value =
      err?.message || "Gagal import Machine ID.";
  } finally {
    machineIdImporting.value = false;
  }
}

async function handleSyncGcc() {
  if (!isAdmin.value || syncing.value) return;

  const ok = window.confirm(
    "Sinkron incremental GCC → gcc_smv_process.\n\n- Data baru: ditambah\n- Data lama: diupdate jika berubah\n- Yang hilang di GCC: dihapus\n\nLanjutkan?"
  );
  if (!ok) return;

  syncing.value = true;
  syncErrorMessage.value = "";
  syncSuccessMessage.value = "";
  syncResult.value = null;

  try {
    const result = await syncProcessStylesFromGcc();
    syncResult.value = result || null;
    syncSuccessMessage.value =
      result?.message ||
      `Sinkron lokal selesai. ${Number(result?.stylesInserted || 0)} baris tersimpan untuk login.`;

    if (syncSuccessTimer) clearTimeout(syncSuccessTimer);
    syncSuccessTimer = setTimeout(() => {
      syncSuccessMessage.value = "";
    }, 5000);

    await loadRows();
  } catch (err) {
    syncErrorMessage.value = err?.message || "Gagal sinkron lokal dari GCC.";
  } finally {
    syncing.value = false;
  }
}

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());
  await Promise.all([loadRows(), loadMachineIdTemplateRows()]);
});

onBeforeUnmount(() => {
  cleanup();
  if (syncSuccessTimer) clearTimeout(syncSuccessTimer);
  if (machineIdSuccessTimer) clearTimeout(machineIdSuccessTimer);
});
</script>

<template>
  <section class="master-ie-page">
    <div v-if="!isAdmin" class="alert error">
      Akses halaman ini hanya untuk admin / IE. Buka dengan mode admin.
    </div>

    <div v-if="errorMessage" class="alert error">
      {{ errorMessage }}
    </div>

    <ProcessStyleGccSyncPanel
      :is-admin="isAdmin"
      :syncing="syncing"
      :sync-error-message="syncErrorMessage"
      :sync-success-message="syncSuccessMessage"
      :sync-result="syncResult"
      @sync="handleSyncGcc"
    />

    <ProcessStyleMachineIdImportPanel
      :is-admin="isAdmin"
      :template-rows="machineIdTemplateRows"
      :importing="machineIdImporting"
      :import-input-key="machineIdImportInputKey"
      :import-file-name="machineIdImportFileName"
      :import-preview-rows="machineIdImportPreviewRows"
      :import-stats="machineIdImportStats"
      :import-error-message="machineIdImportErrorMessage"
      :import-success-message="machineIdImportSuccessMessage"
      :import-result="machineIdImportResult"
      @reset-import="resetMachineIdImport"
      @file-change="handleMachineIdFileChange"
      @submit-import="handleMachineIdImport"
    />

    <ProcessStyleTable
      :loading="loading || syncing"
      :rows="rows"
      :keyword="keyword"
      :source="source"
      :source-message="sourceMessage"
      :total-rows="totalRows"
      :start-index="startIndex"
      :end-index="endIndex"
      :current-page="currentPage"
      :total-pages="totalPages"
      :visible-pages="visiblePages"
      @update:keyword="keyword = $event"
      @search="handleSearch"
      @schedule-search="scheduleSearch"
      @prev="prevPage"
      @next="nextPage"
      @go="goToPage"
    />
  </section>
</template>
