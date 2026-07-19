<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import {
  createProcessStyle,
  deleteProcessStyle,
  getProcessStyleList,
  updateProcessStyle,
} from "../api/machineApi";
import { getInitialAdminMode } from "../utils/adminMode";

const loading = ref(false);
const saving = ref(false);
const errorMessage = ref("");
const successMessage = ref("");
const keyword = ref("");

const isAdmin = ref(false);

const rows = ref([]);

const form = ref({
  id: null,
  styleName: "",
  processName: "",
});

const currentPage = ref(1);
const pageSize = 10;

let searchTimer = null;

const isEditMode = computed(() => Boolean(form.value.id));

const totalRows = computed(() => rows.value.length);

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(totalRows.value / pageSize));
});

const startIndex = computed(() => {
  return (currentPage.value - 1) * pageSize;
});

const endIndex = computed(() => {
  return Math.min(startIndex.value + pageSize, totalRows.value);
});

const pagedRows = computed(() => {
  return rows.value.slice(startIndex.value, endIndex.value);
});

const visiblePages = computed(() => {
  const pages = [];
  const total = totalPages.value;
  const current = currentPage.value;

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

function getVal(obj, ...keys) {
  for (const key of keys) {
    if (obj && obj[key] !== undefined && obj[key] !== null) {
      return obj[key];
    }
  }

  return undefined;
}

function normalizeRow(row) {
  return {
    id: Number(getVal(row, "id", "ID") || 0),
    styleName: String(
      getVal(row, "styleName", "StyleName", "style", "STYLE") || ""
    ),
    processName: String(
      getVal(row, "processName", "ProcessName", "proses", "PROSES") || ""
    ),
    createdAt: String(
      getVal(row, "createdAt", "CreatedAt", "created_at", "Created_At") || ""
    ),
  };
}

function formatDateTime(value) {
  if (!value) return "-";

  const d = new Date(value);

  if (Number.isNaN(d.getTime())) {
    return String(value).replace("T", " ").slice(0, 19);
  }

  return d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function showSuccess(message) {
  successMessage.value = message;
  errorMessage.value = "";

  setTimeout(() => {
    successMessage.value = "";
  }, 2500);
}

function showError(message) {
  errorMessage.value = message;
  successMessage.value = "";
}

async function loadRows() {
  loading.value = true;
  errorMessage.value = "";

  try {
    const data = await getProcessStyleList(keyword.value);
    rows.value = data.map(normalizeRow).filter((item) => item.id);

    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value;
    }

    if (currentPage.value < 1) {
      currentPage.value = 1;
    }
  } catch (err) {
    showError(`Gagal mengambil data Master IE: ${err.message}`);
  } finally {
    loading.value = false;
  }
}

function scheduleSearch() {
  if (searchTimer) {
    clearTimeout(searchTimer);
  }

  searchTimer = setTimeout(() => {
    currentPage.value = 1;
    loadRows();
  }, 350);
}

function handleSearch() {
  if (searchTimer) {
    clearTimeout(searchTimer);
  }

  currentPage.value = 1;
  loadRows();
}

function resetForm() {
  form.value = {
    id: null,
    styleName: "",
    processName: "",
  };
}

function editRow(row) {
  form.value = {
    id: row.id,
    styleName: row.styleName,
    processName: row.processName,
  };

  window.scrollTo({
    top: 0,
    behavior: "smooth",
  });
}

async function saveData() {
  if (!isAdmin.value) {
    showError("Akses hanya untuk admin / IE.");
    return;
  }

  const styleName = String(form.value.styleName || "").trim();
  const processName = String(form.value.processName || "").trim();

  if (!styleName) {
    showError("Style wajib diisi.");
    return;
  }

  if (!processName) {
    showError("Proses wajib diisi.");
    return;
  }

  saving.value = true;

  try {
    if (isEditMode.value) {
      await updateProcessStyle(form.value.id, {
        styleName,
        processName,
      });

      showSuccess("Data berhasil diupdate.");
    } else {
      await createProcessStyle({
        styleName,
        processName,
      });

      showSuccess("Data berhasil ditambahkan.");
    }

    resetForm();
    currentPage.value = 1;
    await loadRows();
  } catch (err) {
    showError(`Gagal menyimpan data: ${err.message}`);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row) {
  if (!isAdmin.value) {
    showError("Akses hanya untuk admin / IE.");
    return;
  }

  const ok = window.confirm(
    `Hapus proses "${row.processName}" untuk style "${row.styleName}"?`
  );

  if (!ok) return;

  loading.value = true;

  try {
    await deleteProcessStyle(row.id);
    showSuccess("Data berhasil dihapus.");
    await loadRows();
  } catch (err) {
    showError(`Gagal menghapus data: ${err.message}`);
  } finally {
    loading.value = false;
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--;
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
  }
}

function goToPage(page) {
  currentPage.value = page;
}

watch(keyword, () => {
  scheduleSearch();
});

watch(totalRows, () => {
  if (currentPage.value > totalPages.value) {
    currentPage.value = totalPages.value;
  }

  if (currentPage.value < 1) {
    currentPage.value = 1;
  }
});

onMounted(async () => {
  isAdmin.value = await Promise.resolve(getInitialAdminMode());
  await loadRows();
});

onBeforeUnmount(() => {
  if (searchTimer) {
    clearTimeout(searchTimer);
  }
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

    <section class="form-card">
      <div class="form-title">
        <h2>{{ isEditMode ? "Edit Data" : "Tambah Data" }}</h2>

        <button
          v-if="isEditMode"
          type="button"
          class="btn-soft"
          @click="resetForm"
        >
          Batal Edit
        </button>
      </div>

      <form class="form-grid" @submit.prevent="saveData">
        <label>
          <span>Style</span>
          <input
            v-model="form.styleName"
            type="text"
            placeholder="Contoh: 1101482"
            :disabled="!isAdmin || saving"
          />
        </label>

        <label>
          <span>Proses</span>
          <input
            v-model="form.processName"
            type="text"
            placeholder="Contoh: Quilting Front Kanan Horizontal"
            :disabled="!isAdmin || saving"
          />
        </label>

        <button
          type="submit"
          class="btn-primary btn-save-green"
          :disabled="!isAdmin || saving"
        >
          {{ saving ? "Menyimpan..." : isEditMode ? "Update" : "Simpan" }}
        </button>
      </form>
    </section>

    <section class="table-card">
      <div class="table-toolbar">
        <div class="table-title-area">
          <div>
            <h2>List Master IE</h2>
            <p>
              Menampilkan {{ totalRows ? startIndex + 1 : 0 }} - {{ endIndex }}
              dari {{ totalRows }} data.
            </p>
          </div>

          <span class="table-count-badge">
            {{ totalRows }} data
          </span>
        </div>

        <div class="search-box">
          <input
            v-model="keyword"
            type="text"
            placeholder="Cari style / proses..."
            @keyup.enter="handleSearch"
          />

          <button type="button" class="btn-primary small" @click="handleSearch">
            Refresh
          </button>
        </div>
      </div>

      <div v-if="loading" class="loading-box">
        Mengambil data...
      </div>

      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>No</th>
              <th>Style</th>
              <th>Proses</th>
              <th>Created At</th>
              <th v-if="isAdmin">Action</th>
            </tr>
          </thead>

          <tbody>
            <tr v-if="!rows.length">
              <td :colspan="isAdmin ? 5 : 4" class="empty">
                Belum ada data.
              </td>
            </tr>

            <tr v-for="(row, index) in pagedRows" :key="row.id">
              <td>{{ startIndex + index + 1 }}</td>

              <td>
                <strong>{{ row.styleName }}</strong>
              </td>

              <td>{{ row.processName }}</td>

              <td>{{ formatDateTime(row.createdAt) }}</td>

              <td v-if="isAdmin">
                <div class="actions">
                  <button type="button" class="btn-edit" @click="editRow(row)">
                    Edit
                  </button>

                  <button
                    type="button"
                    class="btn-delete"
                    @click="removeRow(row)"
                  >
                    Hapus
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="!loading && rows.length" class="pagination">
        <div class="page-info">
          Menampilkan {{ startIndex + 1 }} - {{ endIndex }} dari {{ totalRows }} data
        </div>

        <div class="page-controls">
          <button type="button" @click="prevPage" :disabled="currentPage <= 1">
            Prev
          </button>

          <button
            v-for="page in visiblePages"
            :key="page"
            type="button"
            class="page-number"
            :class="{ active: currentPage === page }"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>

          <button
            type="button"
            @click="nextPage"
            :disabled="currentPage >= totalPages"
          >
            Next
          </button>
        </div>

        <div class="page-total">
          Page {{ currentPage }} / {{ totalPages }}
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.master-ie-page {
  display: grid;
  gap: 18px;
}

.alert {
  padding: 14px 16px;
  border-radius: 18px;
  font-weight: 900;
  border: 1px solid #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

.alert.error {
  background: #fff1f2;
  border-color: #fecaca;
  color: #b91c1c;
}

.alert.success {
  background: #ecfdf5;
  border-color: #bbf7d0;
  color: #166534;
}

.form-card,
.table-card {
  border-radius: 24px;
  background: #ffffff;
  border: 1px solid #dbeafe;
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.06);
  overflow: hidden;
}

.form-title,
.table-toolbar {
  padding: 18px 20px;
  border-bottom: 1px solid #dbeafe;
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: center;
}

.form-title h2,
.table-toolbar h2 {
  margin: 0;
  color: #0f172a;
  font-size: 18px;
  letter-spacing: -0.02em;
}

.table-toolbar p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 12px;
  font-weight: 800;
}

.table-title-area {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.table-count-badge {
  flex: 0 0 auto;
  border-radius: 999px;
  background: #2563eb;
  color: #ffffff;
  padding: 9px 14px;
  font-size: 12px;
  font-weight: 1000;
  white-space: nowrap;
  box-shadow: 0 10px 20px rgba(37, 99, 235, 0.18);
}

.form-grid {
  padding: 20px;
  display: grid;
  grid-template-columns: 180px minmax(260px, 1fr) auto;
  gap: 12px;
  align-items: end;
}

.form-grid label {
  display: grid;
  gap: 7px;
}

.form-grid span {
  color: #64748b;
  font-size: 12px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.form-grid input,
.search-box input {
  width: 100%;
  min-height: 48px;
  border: 1px solid #cbd5e1;
  border-radius: 14px;
  padding: 0 14px;
  color: #0f172a;
  font-size: 14px;
  font-weight: 800;
  outline: none;
}

.form-grid input:focus,
.search-box input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 5px rgba(37, 99, 235, 0.12);
}

.btn-primary,
.btn-soft,
.btn-edit,
.btn-delete {
  min-height: 44px;
  border-radius: 13px;
  padding: 0 16px;
  font-size: 13px;
  font-weight: 1000;
  cursor: pointer;
}

.btn-primary {
  border: 0;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #ffffff;
  box-shadow: 0 12px 22px rgba(37, 99, 235, 0.18);
}

.btn-primary.btn-save-green {
  background: linear-gradient(135deg, #16a34a, #15803d);
  color: #ffffff;
  box-shadow: 0 12px 22px rgba(22, 163, 74, 0.22);
}

.btn-primary.btn-save-green:hover {
  filter: brightness(1.04);
}

.btn-primary.small {
  min-height: 42px;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-soft {
  border: 1px solid #bfdbfe;
  background: #ffffff;
  color: #0f172a;
}

.search-box {
  display: grid;
  grid-template-columns: minmax(220px, 320px) auto;
  gap: 10px;
}

.loading-box,
.empty {
  padding: 24px;
  text-align: center;
  color: #64748b;
  font-weight: 900;
}

.table-wrap {
  width: 100%;
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;
}

th {
  background: #2563eb;
  color: white;
  padding: 13px 14px;
  text-align: left;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

td {
  padding: 13px 14px;
  border-bottom: 1px solid #e2e8f0;
  color: #0f172a;
  font-size: 13px;
  font-weight: 750;
  vertical-align: top;
}

td strong {
  font-weight: 1000;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-edit {
  border: 1px solid #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

.btn-delete {
  border: 1px solid #fecaca;
  background: #fff1f2;
  color: #b91c1c;
}

.pagination {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 14px;
  align-items: center;
  padding: 14px 18px 16px;
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
  border-top: 1px solid #e2e8f0;
  background: #fbfdff;
}

.page-controls {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.page-controls button {
  min-height: 38px;
  border: 1px solid #dbe4ef;
  background: #ffffff;
  color: #0f172a;
  border-radius: 10px;
  padding: 0 12px;
  font-weight: 1000;
  cursor: pointer;
}

.page-controls button:hover:not(:disabled) {
  background: #f1f5f9;
}

.page-controls button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.page-number.active {
  background: #2563eb;
  color: #ffffff;
  border-color: #2563eb;
}

.page-total {
  text-align: right;
}

@media (max-width: 900px) {
  .form-title,
  .table-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .table-title-area {
    justify-content: space-between;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .search-box {
    grid-template-columns: 1fr;
  }

  .pagination {
    grid-template-columns: 1fr;
    text-align: center;
  }

  .page-total {
    text-align: center;
  }

  .page-controls {
    flex-wrap: wrap;
  }
}
</style>