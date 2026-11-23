<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

import ProcessStyleGccSyncPanel from "../components/process-style/ProcessStyleGccSyncPanel.vue";
import ProcessStyleTable from "../components/process-style/ProcessStyleTable.vue";

import { syncProcessStylesFromGcc } from "../api/machineApi";
import { useProcessStyleGccLive } from "../composables/useProcessStyleGccLive";
import { getInitialAdminMode } from "../utils/adminMode";

import "../assets/process-style-master.css";

const isAdmin = ref(false);
const syncing = ref(false);
const syncErrorMessage = ref("");
const syncSuccessMessage = ref("");
const syncResult = ref(null);

let syncSuccessTimer = null;

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
  await loadRows();
});

onBeforeUnmount(() => {
  cleanup();
  if (syncSuccessTimer) clearTimeout(syncSuccessTimer);
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
