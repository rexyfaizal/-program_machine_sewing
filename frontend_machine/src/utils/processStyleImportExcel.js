import * as XLSX from "xlsx";

export function normalizeHeader(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

export function getCellValue(row, headerName) {
  const target = normalizeHeader(headerName);

  for (const key of Object.keys(row || {})) {
    if (normalizeHeader(key) === target) {
      return row[key];
    }
  }

  return "";
}

export function normalizeProcessStyleImportRows(excelRows) {
  const result = [];
  const unique = new Map();
  const duplicateRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;

  excelRows.forEach((row, index) => {
    const excelRowNumber = index + 2;

    const style = String(getCellValue(row, "STYLE") || "").trim();
    const processName = String(getCellValue(row, "NAMA PROSES") || "").trim();

    if (!style || !processName) {
      skippedEmpty += 1;
      return;
    }

    const key = `${style.toUpperCase()}||${processName.toUpperCase()}`;

    if (unique.has(key)) {
      skippedDuplicate += 1;

      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: unique.get(key),
        style,
        processName,
      });

      return;
    }

    unique.set(key, excelRowNumber);

    result.push({
      style,
      processName,
    });
  });

  return {
    rows: result,
    duplicateRows,
    stats: {
      totalExcelRows: excelRows.length,
      readyRows: result.length,
      skippedEmpty,
      skippedDuplicate,
    },
  };
}

export async function parseProcessStyleExcel(file) {
  if (!file) {
    throw new Error("File Excel belum dipilih.");
  }

  const ext = String(file.name.split(".").pop() || "").toLowerCase();

  if (!["xls", "xlsx"].includes(ext)) {
    throw new Error("File harus format .xls atau .xlsx.");
  }

  const buffer = await file.arrayBuffer();

  const workbook = XLSX.read(buffer, {
    type: "array",
  });

  const firstSheetName = workbook.SheetNames[0];

  if (!firstSheetName) {
    throw new Error("Sheet Excel tidak ditemukan.");
  }

  const worksheet = workbook.Sheets[firstSheetName];

  const excelRows = XLSX.utils.sheet_to_json(worksheet, {
    defval: "",
    raw: false,
  });

  const headerKeys = Object.keys(excelRows[0] || {}).map(normalizeHeader);

  if (!headerKeys.includes("STYLE") || !headerKeys.includes("NAMA PROSES")) {
    throw new Error(
      "Header Excel wajib memiliki kolom STYLE dan NAMA PROSES. Kolom LINE akan diabaikan."
    );
  }

  const normalized = normalizeProcessStyleImportRows(excelRows);

  if (!normalized.rows.length) {
    throw new Error(
      "Tidak ada data valid. Pastikan kolom STYLE dan NAMA PROSES terisi."
    );
  }

  return {
    fileName: file.name,
    rows: normalized.rows,
    duplicateRows: normalized.duplicateRows,
    stats: normalized.stats,
  };
}