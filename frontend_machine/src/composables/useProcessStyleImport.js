import { computed, ref } from "vue";

import { importProcessStyles } from "../api/machineApi";
import { parseProcessStyleExcel } from "../utils/processStyleImportExcel";

export function useProcessStyleImport({ isAdmin, onImported }) {
  const importing = ref(false);
  const importInputKey = ref(0);

  const importFileName = ref("");
  const importPreviewRows = ref([]);
  const importDuplicateRows = ref([]);

  const importErrorMessage = ref("");
  const importSuccessMessage = ref("");

  const importStats = ref({
    totalExcelRows: 0,
    readyRows: 0,
    skippedEmpty: 0,
    skippedDuplicate: 0,
  });

  let importSuccessTimer = null;

  const importPreviewDisplayRows = computed(() => {
    return importPreviewRows.value.slice(0, 10);
  });

  function showImportSuccess(message) {
    importSuccessMessage.value = message;
    importErrorMessage.value = "";

    if (importSuccessTimer) {
      clearTimeout(importSuccessTimer);
    }

    importSuccessTimer = setTimeout(() => {
      importSuccessMessage.value = "";
    }, 3500);
  }

  function showImportError(message) {
    importErrorMessage.value = message;
    importSuccessMessage.value = "";
  }

  function resetImport(keepMessage = false) {
    importFileName.value = "";
    importPreviewRows.value = [];
    importDuplicateRows.value = [];

    importStats.value = {
      totalExcelRows: 0,
      readyRows: 0,
      skippedEmpty: 0,
      skippedDuplicate: 0,
    };

    importInputKey.value += 1;

    if (!keepMessage) {
      importErrorMessage.value = "";
      importSuccessMessage.value = "";
    }
  }

  async function handleImportFileChange(event) {
    importErrorMessage.value = "";
    importSuccessMessage.value = "";
    importPreviewRows.value = [];
    importDuplicateRows.value = [];

    const file = event.target.files?.[0];

    if (!file) {
      resetImport();
      return;
    }

    if (!isAdmin.value) {
      resetImport();
      showImportError("Akses upload hanya untuk admin / IE.");
      return;
    }

    try {
      const result = await parseProcessStyleExcel(file);

      importFileName.value = result.fileName;
      importPreviewRows.value = result.rows;
      importDuplicateRows.value = result.duplicateRows;
      importStats.value = result.stats;

      showImportSuccess(
        `File terbaca. Siap import ${result.rows.length} data. Kolom LINE diabaikan.`
      );
    } catch (err) {
      resetImport();
      showImportError(err.message || "Gagal membaca file Excel.");
    }
  }

  async function submitImportExcel() {
    if (!isAdmin.value) {
      showImportError("Akses import hanya untuk admin / IE.");
      return;
    }

    if (!importPreviewRows.value.length) {
      showImportError("Belum ada data yang siap diimport.");
      return;
    }

    importing.value = true;
    importErrorMessage.value = "";
    importSuccessMessage.value = "Mengimport data Master IE...";

    try {
      const totalReadyRows = importPreviewRows.value.length;
      const result = await importProcessStyles(importPreviewRows.value);

      showImportSuccess(
        `${result?.message || "Import berhasil"} | Total: ${
          result?.total ?? totalReadyRows
        }, Inserted: ${result?.inserted ?? 0}, Skipped: ${result?.skipped ?? 0}`
      );

      resetImport(true);

      if (typeof onImported === "function") {
        await onImported();
      }
    } catch (err) {
      showImportError(`Gagal import Excel: ${err.message}`);
    } finally {
      importing.value = false;
    }
  }

  function cleanupProcessStyleImport() {
    if (importSuccessTimer) {
      clearTimeout(importSuccessTimer);
    }
  }

  return {
    importing,
    importInputKey,
    importFileName,
    importPreviewRows,
    importPreviewDisplayRows,
    importDuplicateRows,
    importStats,
    importErrorMessage,
    importSuccessMessage,
    resetImport,
    handleImportFileChange,
    submitImportExcel,
    cleanupProcessStyleImport,
  };
}