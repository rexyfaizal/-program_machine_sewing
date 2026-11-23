import { computed, ref } from "vue";

import { getProcessStyleGccList } from "../api/machineApi";

export function useProcessStyleGccLive() {
  const loading = ref(false);
  const errorMessage = ref("");
  const keyword = ref("");
  const rows = ref([]);
  const source = ref("gcc-live");
  const sourceMessage = ref("");
  const currentPage = ref(1);
  const pageSize = ref(20);
  const pagination = ref({
    total: 0,
    perPage: 20,
    currentPage: 1,
    lastPage: 1,
    from: 0,
    to: 0,
  });

  let searchTimer = null;

  const totalRows = computed(() => Number(pagination.value.total || 0));
  const totalPages = computed(() => Math.max(1, Number(pagination.value.lastPage || 1)));
  const startIndex = computed(() => {
    const from = Number(pagination.value.from || 0);
    return from > 0 ? from - 1 : 0;
  });
  const endIndex = computed(() => Number(pagination.value.to || 0));

  const visiblePages = computed(() => {
    const pages = [];
    const total = totalPages.value;
    const current = currentPage.value;
    let start = Math.max(1, current - 2);
    let end = Math.min(total, current + 2);
    if (current <= 3) end = Math.min(total, 5);
    if (current >= total - 2) start = Math.max(1, total - 4);
    for (let i = start; i <= end; i++) pages.push(i);
    return pages;
  });

  async function loadRows() {
    loading.value = true;
    errorMessage.value = "";
    try {
      const data = await getProcessStyleGccList({
        page: currentPage.value,
        limit: pageSize.value,
        q: keyword.value,
      });
      rows.value = Array.isArray(data?.rows) ? data.rows : [];
      source.value = String(data?.source || "gcc-live");
      sourceMessage.value = String(data?.message || "");
      const p = data?.pagination || {};
      pagination.value = {
        total: Number(p.total || 0),
        perPage: Number(p.perPage || pageSize.value),
        currentPage: Number(p.currentPage || currentPage.value),
        lastPage: Number(p.lastPage || 1),
        from: Number(p.from || 0),
        to: Number(p.to || 0),
      };
      currentPage.value = pagination.value.currentPage || 1;
    } catch (err) {
      rows.value = [];
      errorMessage.value = err?.message || "Gagal mengambil data Master IE dari GCC.";
    } finally {
      loading.value = false;
    }
  }

  function handleSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    currentPage.value = 1;
    loadRows();
  }

  function scheduleSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      currentPage.value = 1;
      loadRows();
    }, 350);
  }

  function prevPage() {
    if (currentPage.value <= 1) return;
    currentPage.value -= 1;
    loadRows();
  }

  function nextPage() {
    if (currentPage.value >= totalPages.value) return;
    currentPage.value += 1;
    loadRows();
  }

  function goToPage(page) {
    const target = Number(page || 1);
    if (target < 1 || target > totalPages.value) return;
    currentPage.value = target;
    loadRows();
  }

  function cleanup() {
    if (searchTimer) clearTimeout(searchTimer);
  }

  return {
    loading,
    errorMessage,
    keyword,
    rows,
    source,
    sourceMessage,
    currentPage,
    pageSize,
    pagination,
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
  };
}
