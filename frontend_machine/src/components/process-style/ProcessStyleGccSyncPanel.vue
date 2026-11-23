<script setup>
const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
  },
  syncing: {
    type: Boolean,
    default: false,
  },
  syncErrorMessage: {
    type: String,
    default: "",
  },
  syncSuccessMessage: {
    type: String,
    default: "",
  },
  syncResult: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["sync"]);
</script>

<template>
  <section class="import-card gcc-sync-card">
    <div class="import-head">
      <div class="import-head-text">
        <h2>Sinkron GCC → Tabel Lokal</h2>
        <p>
          Mode <strong>incremental</strong>: insert data baru, update yang berubah,
          hapus yang sudah tidak ada di GCC. Tabel
          <strong>gcc_smv_process</strong> dipakai Master IE &amp; login operator.
        </p>
      </div>
    </div>

    <div class="gcc-sync-actions">
      <button
        type="button"
        class="btn-primary"
        :disabled="!props.isAdmin || props.syncing"
        @click="emit('sync')"
      >
        {{ props.syncing ? "Menyinkronkan..." : "Sinkron dari GCC" }}
      </button>
      <small>
        Ambil data GCC masih perlu waktu (paginasi). DB hanya upsert, tidak hapus-semua dulu.
      </small>
    </div>

    <div v-if="props.syncErrorMessage" class="alert error">
      {{ props.syncErrorMessage }}
    </div>

    <div v-if="props.syncSuccessMessage" class="alert success">
      {{ props.syncSuccessMessage }}
    </div>

    <div v-if="props.syncResult" class="gcc-sync-stats">
      <span>Header GCC: {{ props.syncResult.headersFetched || 0 }}/{{ props.syncResult.totalHeaders || 0 }}</span>
      <span>Halaman: {{ props.syncResult.pagesFetched || 0 }}</span>
      <span>Baru: {{ props.syncResult.stylesInserted || 0 }}</span>
      <span>Update: {{ props.syncResult.stylesUpdated || 0 }}</span>
      <span>Dihapus: {{ props.syncResult.stylesDeleted || 0 }}</span>
    </div>
  </section>
</template>

<style scoped>
.gcc-sync-card {
  grid-column: 1 / -1;
}

.gcc-sync-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
}

.gcc-sync-actions small {
  color: #64748b;
  font-weight: 700;
  font-size: 12px;
}

.gcc-sync-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.gcc-sync-stats span {
  padding: 7px 10px;
  border-radius: 999px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1e40af;
  font-size: 12px;
  font-weight: 900;
}
</style>
