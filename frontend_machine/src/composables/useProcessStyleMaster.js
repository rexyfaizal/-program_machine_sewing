import { computed, ref, watch } from "vue";

import {
  createProcessStyle,
  deleteProcessStyle,
  getProcessStyleList,
  updateProcessStyle,
} from "../api/machineApi";

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
      getVal(
        row,
        "processName",
        "ProcessName",
        "proses",
        "PROSES",
        "namaProses",
        "NamaProses",
        "NAMA PROSES"
      ) || ""
    ),
    createdAt: String(
      getVal(row, "createdAt", "CreatedAt", "created_at", "Created_At") || ""
    ),
  };
}

export function useProcessStyleMaster({ isAdmin }) {
  const loading = ref(false);
  const saving = ref(false);

  const errorMessage = ref("");
  const successMessage = ref("");

  const keyword = ref("");
  const rows = ref([]);

  const form = ref({
    id: null,
    styleName: "",
    processName: "",
  });

  const currentPage = ref(1);
  const pageSize = 10;

  let searchTimer = null;
  let successTimer = null;

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

  function showSuccess(message) {
    successMessage.value = message;
    errorMessage.value = "";

    if (successTimer) {
      clearTimeout(successTimer);
    }

    successTimer = setTimeout(() => {
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

  function cleanupProcessStyleMaster() {
    if (searchTimer) {
      clearTimeout(searchTimer);
    }

    if (successTimer) {
      clearTimeout(successTimer);
    }
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

  return {
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
  };
}