/** Definisi kolom tabel dashboard — dipakai picker hide/show. */
export const MACHINE_TABLE_COLUMN_DEFS = [
  { key: "no", label: "No", locked: true },
  { key: "prosesId", label: "Proses ID" },
  { key: "processName", label: "Process Name" },
  { key: "style", label: "Style" },
  { key: "branch", label: "Branch" },
  { key: "location", label: "Line" },
  { key: "spvId", label: "SPV_ID" },
  { key: "machineId", label: "Machine ID" },
  { key: "mikaId", label: "Mika ID" },
  { key: "nik", label: "Operator_ID" },
  { key: "namaOperator", label: "Nama Operator" },
  { key: "operatorNote", label: "Operator Note" },
  { key: "productivity", label: "Utilisasi Mesin" },
  { key: "status", label: "Status" },
  { key: "runtime", label: "Power On Duration" },
  { key: "procTime", label: "Running Time" },
  { key: "lossTime", label: "Loss Time" },
  { key: "output", label: "Output", adminOnly: true },
  { key: "uuid", label: "UUID / Table", adminOnly: true },
  { key: "action", label: "Action", adminOnly: true },
  { key: "ai", label: "AI", adminOnly: true },
  { key: "grafik", label: "Grafik" },
];

export const MACHINE_TABLE_COLUMNS_STORAGE_KEY =
  "sewing.dashboard.tableColumns.v1";

export function buildDefaultColumnVisibility() {
  const map = {};
  for (const col of MACHINE_TABLE_COLUMN_DEFS) {
    map[col.key] = true;
  }
  return map;
}

export function normalizeColumnVisibility(raw) {
  const next = buildDefaultColumnVisibility();
  if (!raw || typeof raw !== "object") return next;

  for (const col of MACHINE_TABLE_COLUMN_DEFS) {
    if (col.locked) {
      next[col.key] = true;
      continue;
    }
    if (Object.prototype.hasOwnProperty.call(raw, col.key)) {
      next[col.key] = raw[col.key] !== false;
    }
  }
  return next;
}

export function loadColumnVisibility() {
  try {
    const raw = localStorage.getItem(MACHINE_TABLE_COLUMNS_STORAGE_KEY);
    if (!raw) return buildDefaultColumnVisibility();
    return normalizeColumnVisibility(JSON.parse(raw));
  } catch {
    return buildDefaultColumnVisibility();
  }
}

export function saveColumnVisibility(map) {
  try {
    localStorage.setItem(
      MACHINE_TABLE_COLUMNS_STORAGE_KEY,
      JSON.stringify(normalizeColumnVisibility(map))
    );
  } catch {
    /* ignore quota / private mode */
  }
}

export function listPickerColumns(showActions = false) {
  return MACHINE_TABLE_COLUMN_DEFS.filter((col) => {
    if (col.locked) return false;
    if (col.adminOnly && !showActions) return false;
    return true;
  });
}

/** Urutan kolom per mode (untuk hitung colspan). */
export function listTableColumnKeys(showActions = false) {
  if (showActions) {
    return [
      "no",
      "prosesId",
      "processName",
      "style",
      "branch",
      "location",
      "spvId",
      "machineId",
      "mikaId",
      "nik",
      "namaOperator",
      "operatorNote",
      "productivity",
      "status",
      "runtime",
      "procTime",
      "lossTime",
      "output",
      "uuid",
      "action",
      "ai",
      "grafik",
    ];
  }

  return [
    "no",
    "prosesId",
    "processName",
    "style",
    "branch",
    "location",
    "spvId",
    "machineId",
    "mikaId",
    "nik",
    "namaOperator",
    "operatorNote",
    "runtime",
    "procTime",
    "lossTime",
    "productivity",
    "status",
    "grafik",
  ];
}

export function countVisibleColumns(visibility, showActions = false) {
  const map = normalizeColumnVisibility(visibility);
  return listTableColumnKeys(showActions).filter((key) => map[key] !== false)
    .length;
}
