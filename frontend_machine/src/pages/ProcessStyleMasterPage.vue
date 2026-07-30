<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

import ProcessStyleFormCard from "../components/process-style/ProcessStyleFormCard.vue";
import ProcessStyleImportPanel from "../components/process-style/ProcessStyleImportPanel.vue";
import ProcessStyleTable from "../components/process-style/ProcessStyleTable.vue";

import { useProcessStyleImport } from "../composables/useProcessStyleImport";
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

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());
  await loadRows();
});

onBeforeUnmount(() => {
  cleanupProcessStyleMaster();
  cleanupProcessStyleImport();
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