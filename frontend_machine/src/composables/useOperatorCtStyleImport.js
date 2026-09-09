import { ref, computed } from "vue";

import { importOperatorCtStyle } from "../api/machineApi";
import { parseOperatorCtStyleExcel } from "../utils/operatorCtStyleImportExcel";

export function useOperatorCtStyleImport({ isAdmin, onImported }) {
  const importing = ref(false);
  const importInputKey = ref(0);

  const importFileName = ref("");
  const importPreviewRows = ref([]);
  const importDuplicateRows = ref([]);
  const importErrorRows = ref([]);

  const importErrorMessage = ref("");
  const importSuccessMessage = ref("");

  const importStats = ref({
    totalExcelRows: 0,
    readyRows: 0,
    skippedEmpty: 0,
    skippedDuplicate: 0,
    skippedInvalid: 0,
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
    importErrorRows.value = [];

    importStats.value = {
      totalExcelRows: 0,
      readyRows: 0,
      skippedEmpty: 0,
      skippedDuplicate: 0,
      skippedInvalid: 0,
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
    importErrorRows.value = [];

    const file = event.target.files?.[0];
    if (!file) {
      resetImport();
      return;
    }

    importFileName.value = file.name;

    try {
      const parsed = await parseOperatorCtStyleExcel(file);
      importPreviewRows.value = parsed.rows;
      importDuplicateRows.value = parsed.duplicateRows;
      importErrorRows.value = parsed.errorRows;
      importStats.value = parsed.stats;

      if (!importPreviewRows.value.length) {
        showImportError(
          "Tidak ada baris valid. Pastikan kolom Style, Proses, dan CT Total terisi."
        );
      }
    } catch (err) {
      resetImport(true);
      showImportError(err?.message || "Gagal membaca file Excel CT GM3.");
    }
  }

  async function submitImportExcel() {
    if (!isAdmin?.value) {
      showImportError("Upload CT GM3 hanya untuk mode admin / IE.");
      return;
    }

    if (!importPreviewRows.value.length) {
      showImportError("Tidak ada data valid untuk diimport.");
      return;
    }

    importing.value = true;
    importErrorMessage.value = "";

    try {
      const payload = {
        rows: importPreviewRows.value.map((row) => ({
          styleName: row.styleName,
          processName: row.processName,
          ctTotal: row.ctTotal,
        })),
      };

      const result = await importOperatorCtStyle(payload);
      const upserted = Number(result?.upserted || 0);
      const skipped = Number(result?.skipped || 0);

      showImportSuccess(
        `Import CT GM3 berhasil (${upserted} tersimpan, ${skipped} dilewati).`
      );

      resetImport(true);

      if (typeof onImported === "function") {
        await onImported();
      }
    } catch (err) {
      showImportError(err?.message || "Gagal import CT GM3 ke server.");
    } finally {
      importing.value = false;
    }
  }

  function cleanupOperatorCtStyleImport() {
    if (importSuccessTimer) {
      clearTimeout(importSuccessTimer);
      importSuccessTimer = null;
    }
  }

  return {
    importing,
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
    cleanupOperatorCtStyleImport,
  };
}
