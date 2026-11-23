import { computed, ref, watch } from "vue";

export function useMachineTable(props) {
  const currentPage = ref(1);
  /** @type {import('vue').Ref<'desc' | 'asc'>} */
  const utilSortDir = ref("desc");

  const totalRows = computed(() => props.machines.length);

  const totalPages = computed(() => {
    return Math.max(1, Math.ceil(totalRows.value / props.pageSize));
  });

  const startIndex = computed(() => {
    return (currentPage.value - 1) * props.pageSize;
  });

  const endIndex = computed(() => {
    return Math.min(startIndex.value + props.pageSize, totalRows.value);
  });

  const sortedMachines = computed(() => {
    const dir = utilSortDir.value === "asc" ? 1 : -1;
    return [...props.machines].sort((a, b) => {
      const av = Number(a.productivity || 0);
      const bv = Number(b.productivity || 0);
      if (av === bv) return 0;
      return av > bv ? dir : -dir;
    });
  });

  function toggleUtilSort() {
    utilSortDir.value = utilSortDir.value === "desc" ? "asc" : "desc";
    currentPage.value = 1;
  }

  const pagedMachines = computed(() => {
    return sortedMachines.value.slice(startIndex.value, endIndex.value);
  });

  const totalCols = computed(() => {
    return props.showActions ? 19 : 15;
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

  watch(
    () => totalRows.value,
    () => {
      if (currentPage.value > totalPages.value) {
        currentPage.value = totalPages.value;
      }

      if (currentPage.value < 1) {
        currentPage.value = 1;
      }
    }
  );

  watch(
    () => props.pageSize,
    () => {
      if (currentPage.value > totalPages.value) {
        currentPage.value = totalPages.value;
      }
    }
  );

  watch(
    () => props.locationValue,
    () => {
      currentPage.value = 1;
    }
  );

  watch(
    () => props.keyword,
    () => {
      currentPage.value = 1;
    }
  );

  function nextPage() {
    if (currentPage.value < totalPages.value) {
      currentPage.value++;
    }
  }

  function prevPage() {
    if (currentPage.value > 1) {
      currentPage.value--;
    }
  }

  function goToPage(page) {
    currentPage.value = page;
  }

  return {
    currentPage,
    utilSortDir,
    totalRows,
    totalPages,
    startIndex,
    endIndex,
    sortedMachines,
    pagedMachines,
    totalCols,
    visiblePages,
    toggleUtilSort,
    nextPage,
    prevPage,
    goToPage,
  };
}