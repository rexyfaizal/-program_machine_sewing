import { computed, ref, watch } from "vue";
import {
  downTimeSec,
  getDisplayMachineName,
  getOperatorDisplayRows,
  getOperatorNoteItems,
} from "../utils/machineTableFormat";
import { parseLocationParts } from "../utils/gm3Shift";

const NUMERIC_SORT_KEYS = new Set([
  "no",
  "productivity",
  "runtime",
  "procTime",
  "lossTime",
  "output",
]);

function branchLabel(machine) {
  const { factory } = parseLocationParts(machine?.location);
  return factory || "";
}

function locationLabel(machine) {
  const raw = String(machine?.location || "").trim();
  if (!raw || raw === "-") return "";
  const { lineName } = parseLocationParts(raw);
  return lineName || raw;
}

function firstOperator(machine) {
  const rows = getOperatorDisplayRows(machine);
  return rows[0] || null;
}

function noteSortText(machine) {
  const items = getOperatorNoteItems(machine);
  if (!items.length) return "";
  return items
    .map((note) =>
      [
        [note.reasonName, note.body].filter(Boolean).join(" "),
        note.timeRange,
      ]
        .filter(Boolean)
        .join(" | ") || String(note.text || "")
    )
    .join(" | ");
}

function sortValue(machine, key, index) {
  switch (key) {
    case "no":
      return index;
    case "prosesId": {
      const op = firstOperator(machine);
      return String(op?.prosesId || machine?.operatorProsesId || "").trim();
    }
    case "processName":
      return String(getDisplayMachineName(machine) || "").trim();
    case "branch":
      return String(branchLabel(machine) || "").trim();
    case "location":
      return String(locationLabel(machine) || "").trim();
    case "machineId":
      return String(machine?.machineId || "").trim();
    case "nik": {
      const op = firstOperator(machine);
      return String(op?.nik || "").trim();
    }
    case "namaOperator": {
      const op = firstOperator(machine);
      return String(op?.name || op?.label || machine?.pic || "").trim();
    }
    case "operatorNote":
      return noteSortText(machine);
    case "productivity":
      return Number(machine?.productivity || 0);
    case "status":
      return String(machine?.status || "BAD").trim().toUpperCase();
    case "runtime":
      return Number(machine?.runtime || 0);
    case "procTime":
      return Number(machine?.procTime || 0);
    case "lossTime":
      return downTimeSec(machine);
    case "output":
      return Number(machine?.output || 0);
    case "uuid":
      return String(machine?.uuid || "").trim();
    default:
      return "";
  }
}

function compareValues(a, b, key) {
  if (NUMERIC_SORT_KEYS.has(key)) {
    const av = Number(a);
    const bv = Number(b);
    if (av === bv) return 0;
    return av > bv ? 1 : -1;
  }

  const as = String(a ?? "").toLowerCase();
  const bs = String(b ?? "").toLowerCase();
  if (as === bs) return 0;
  return as > bs ? 1 : -1;
}

export function useMachineTable(props) {
  const currentPage = ref(1);
  /** @type {import('vue').Ref<string>} */
  const sortKey = ref("productivity");
  /** @type {import('vue').Ref<'desc' | 'asc'>} */
  const sortDir = ref("desc");

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
    const key = sortKey.value;
    const dir = sortDir.value === "asc" ? 1 : -1;

    // Nomor tetap dari ranking default (Utilisasi ↓), ikut pindah saat kolom lain di-sort.
    const baseRanked = props.machines.map((machine, index) => ({ machine, index }));
    baseRanked.sort((a, b) => {
      const cmp = compareValues(
        sortValue(a.machine, "productivity", a.index),
        sortValue(b.machine, "productivity", b.index),
        "productivity"
      );
      if (cmp !== 0) return cmp * -1; // desc
      return a.index - b.index;
    });

    const indexed = baseRanked.map((row, rank) => ({
      machine: row.machine,
      index: row.index,
      rowNo: rank + 1,
    }));

    indexed.sort((a, b) => {
      const sortIndex = key === "no" ? a.rowNo - 1 : a.index;
      const sortIndexB = key === "no" ? b.rowNo - 1 : b.index;
      const cmp = compareValues(
        key === "no" ? a.rowNo : sortValue(a.machine, key, a.index),
        key === "no" ? b.rowNo : sortValue(b.machine, key, b.index),
        key
      );
      if (cmp !== 0) return cmp * dir;
      return sortIndex - sortIndexB;
    });

    return indexed.map((row) => ({
      ...row.machine,
      __rowNo: row.rowNo,
    }));
  });

  function setSort(key, dir) {
    if (!key) return;

    if (dir === "asc" || dir === "desc") {
      sortKey.value = key;
      sortDir.value = dir;
    } else if (sortKey.value === key) {
      sortDir.value = sortDir.value === "desc" ? "asc" : "desc";
    } else {
      sortKey.value = key;
      sortDir.value = NUMERIC_SORT_KEYS.has(key) ? "desc" : "asc";
    }

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
    sortKey,
    sortDir,
    /** @deprecated gunakan sortDir — tetap diekspos agar tidak pecah sementara */
    utilSortDir: sortDir,
    totalRows,
    totalPages,
    startIndex,
    endIndex,
    sortedMachines,
    pagedMachines,
    totalCols,
    visiblePages,
    setSort,
    nextPage,
    prevPage,
    goToPage,
  };
}
