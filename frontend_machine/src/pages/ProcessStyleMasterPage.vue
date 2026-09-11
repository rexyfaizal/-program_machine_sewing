<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

import ProcessStyleFormCard from "../components/process-style/ProcessStyleFormCard.vue";
import ProcessStyleImportPanel from "../components/process-style/ProcessStyleImportPanel.vue";
import ProcessStyleCtImportPanel from "../components/process-style/ProcessStyleCtImportPanel.vue";
import ProcessStyleTable from "../components/process-style/ProcessStyleTable.vue";

import { useProcessStyleImport } from "../composables/useProcessStyleImport";
import { useOperatorCtStyleImport } from "../composables/useOperatorCtStyleImport";
import { useProcessStyleMaster } from "../composables/useProcessStyleMaster";
import { getInitialAdminMode } from "../utils/adminMode";

import "../assets/process-style-master.css";

const isAdmin = ref(false);

const {
  loading,
  saving,
  errorMessage,
  successMessage,
  keyword,
  rows,
  form,
  currentPage,
  isEditMode,
  totalRows,
  totalPages,
  startIndex,
  endIndex,
  pagedRows,
  visiblePages,
  loadRows,
  handleSearch,
  resetForm,
  editRow,
  saveData,
  removeRow,
  prevPage,
  nextPage,
  goToPage,
  cleanupProcessStyleMaster,
} = useProcessStyleMaster({
  isAdmin,
});

const {
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
} = useProcessStyleImport({
  isAdmin,
  onImported: async () => {
    currentPage.value = 1;
    await loadRows();
  },
});

const {
  importing: ctStyleImporting,
  importInputKey: ctStyleImportInputKey,
  importFileName: ctStyleImportFileName,
  importPreviewRows: ctStyleImportPreviewRows,
  importPreviewDisplayRows: ctStyleImportPreviewDisplayRows,
  importDuplicateRows: ctStyleImportDuplicateRows,
  importStats: ctStyleImportStats,
  importErrorMessage: ctStyleImportErrorMessage,
  importSuccessMessage: ctStyleImportSuccessMessage,
  resetImport: resetCtStyleImport,
  handleImportFileChange: handleCtStyleImportFileChange,
  submitImportExcel: submitCtStyleImportExcel,
  cleanupOperatorCtStyleImport,
} = useOperatorCtStyleImport({
  isAdmin,
  onImported: async () => {
    await loadRows();
  },
});

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());
  await loadRows();
});

onBeforeUnmount(() => {
  cleanupProcessStyleMaster();
  cleanupProcessStyleImport();
  cleanupOperatorCtStyleImport();
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

    <div v-if="successMessage" class="alert success">
      {{ successMessage }}
    </div>

    <div class="import-grid">
      <ProcessStyleImportPanel
        :is-admin="isAdmin"
        :importing="importing"
        :import-input-key="importInputKey"
        :import-file-name="importFileName"
        :import-preview-rows="importPreviewRows"
        :import-preview-display-rows="importPreviewDisplayRows"
        :import-duplicate-rows="importDuplicateRows"
        :import-stats="importStats"
        :import-error-message="importErrorMessage"
        :import-success-message="importSuccessMessage"
        @reset-import="resetImport"
        @file-change="handleImportFileChange"
        @submit-import="submitImportExcel"
      />

      <ProcessStyleCtImportPanel
        :is-admin="isAdmin"
        :template-rows="rows"
        :importing="ctStyleImporting"
        :import-input-key="ctStyleImportInputKey"
        :import-file-name="ctStyleImportFileName"
        :import-preview-rows="ctStyleImportPreviewRows"
        :import-preview-display-rows="ctStyleImportPreviewDisplayRows"
        :import-duplicate-rows="ctStyleImportDuplicateRows"
        :import-stats="ctStyleImportStats"
        :import-error-message="ctStyleImportErrorMessage"
        :import-success-message="ctStyleImportSuccessMessage"
        @reset-import="resetCtStyleImport"
        @file-change="handleCtStyleImportFileChange"
        @submit-import="submitCtStyleImportExcel"
      />
    </div>

    <ProcessStyleFormCard
      :is-admin="isAdmin"
      :saving="saving"
      :form="form"
      :is-edit-mode="isEditMode"
      @update:form="form = $event"
      @save="saveData"
      @reset="resetForm"
    />

    <ProcessStyleTable
      :is-admin="isAdmin"
      :loading="loading"
      :rows="rows"
      :paged-rows="pagedRows"
      :keyword="keyword"
      :total-rows="totalRows"
      :start-index="startIndex"
      :end-index="endIndex"
      :current-page="currentPage"
      :total-pages="totalPages"
      :visible-pages="visiblePages"
      @update:keyword="keyword = $event"
      @search="handleSearch"
      @edit="editRow"
      @remove="removeRow"
      @prev="prevPage"
      @next="nextPage"
      @go="goToPage"
    />
  </section>
</template>
