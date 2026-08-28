import { ref, computed } from "vue";

import { importOperatorOutputTarget } from "../api/machineApi";
import { parseOperatorOutputTargetExcel } from "../utils/operatorOutputTargetImportExcel";

export function useOperatorOutputTargetImport({ isAdmin, selectedDate, onImported }) {
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
      const fallbackDate =
        typeof selectedDate === "function"
          ? selectedDate()
          : selectedDate?.value || selectedDate || "";
      const parsed = await parseOperatorOutputTargetExcel(file, fallbackDate);
      importPreviewRows.value = parsed.rows;
      importDuplicateRows.value = parsed.duplicateRows;
      importErrorRows.value = parsed.errorRows;
      importStats.value = parsed.stats;

      if (!importPreviewRows.value.length) {
        showImportError(
          "Tidak ada baris valid. Pastikan kolom Tanggal, UUID, dan Line terisi."
        );
      }
    } catch (err) {
      resetImport(true);
      showImportError(err?.message || "Gagal membaca file Excel.");
    }
  }

  async function submitImportExcel() {
    if (!isAdmin?.value) {
      showImportError("Upload output target hanya untuk mode admin / IE.");
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
          workDate: row.workDate,
          uuid: row.uuid,
          line: row.line,
          area: row.area,
          outputTarget: row.outputTarget,
        })),
      };

      const result = await importOperatorOutputTarget(payload);
      const upserted = Number(result?.upserted || 0);
      const skipped = Number(result?.skipped || 0);

      showImportSuccess(
        `Import output target berhasil (${upserted} tersimpan, ${skipped} dilewati).`
      );

      resetImport(true);

      if (typeof onImported === "function") {
        await onImported();
      }
    } catch (err) {
      showImportError(err?.message || "Gagal import output target ke server.");
    } finally {
      importing.value = false;
    }
  }

  function cleanupOperatorOutputTargetImport() {
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
    cleanupOperatorOutputTargetImport,
  };
}
