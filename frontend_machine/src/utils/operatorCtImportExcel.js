import * as XLSX from "xlsx";

import { normalizeCtLine } from "./operatorCt";
import {
  formatTemplateOperatorName,
  formatTemplateOperatorNik,
  hasTemplateOperatorInfo,
  upsertOperatorTemplateRow,
} from "./operatorTemplateRows";

function normalizeHeader(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/[_]+/g, " ")
    .replace(/\s+/g, " ");
}

function getCellValue(row, ...headerNames) {
  for (const headerName of headerNames) {
    const target = normalizeHeader(headerName);

    for (const key of Object.keys(row || {})) {
      if (normalizeHeader(key) === target) {
        return row[key];
      }
    }
  }

  return "";
}

function parseCtNumber(value) {
  const text = String(value ?? "")
    .trim()
    .replace(/,/g, "");
  if (!text) return 0;

  const num = Number(text);
  if (!Number.isFinite(num) || num < 0) return 0;
  return num;
}

export function normalizeOperatorCtImportRows(excelRows) {
  const uniqueMap = new Map();
  const duplicateRows = [];
  const errorRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;
  let skippedInvalid = 0;

  (Array.isArray(excelRows) ? excelRows : []).forEach((row, index) => {
    const excelRowNumber = index + 2;
    const line = normalizeCtLine(
      getCellValue(row, "LINE", "LOCATION", "LOKASI")
    );
    const uuid = String(getCellValue(row, "UUID") || "").trim();
    const area = String(getCellValue(row, "AREA") || "").trim();
    const ctSum = parseCtNumber(getCellValue(row, "CT SUM", "CT_SUM"));
    const ctStd = parseCtNumber(getCellValue(row, "CT STD", "CT_STD"));
    const ctValue = parseCtNumber(getCellValue(row, "CT"));

    if (!uuid && !line) {
      skippedEmpty += 1;
      return;
    }

    if (!uuid || !line) {
      skippedInvalid += 1;
      errorRows.push({
        excelRowNumber,
        message: "UUID dan Line wajib diisi.",
        line,
        uuid,
      });
      return;
    }

    const key = `${uuid.toUpperCase()}||${line.toUpperCase()}`;
    if (uniqueMap.has(key)) {
      skippedDuplicate += 1;
      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: uniqueMap.get(key).excelRowNumber,
        line,
        uuid,
      });
    }

    uniqueMap.set(key, {
      uuid,
      line,
      area,
      ctSum,
      ctStd,
      ctValue,
      excelRowNumber,
    });
  });

  const result = [...uniqueMap.values()];

  return {
    rows: result,
    duplicateRows,
    errorRows,
    stats: {
      totalExcelRows: excelRows.length,
      readyRows: result.length,
      skippedEmpty,
      skippedDuplicate,
      skippedInvalid,
    },
  };
}

export async function parseOperatorCtExcel(file) {
  if (!file) {
    throw new Error("File Excel belum dipilih.");
  }

  const buffer = await file.arrayBuffer();
  const workbook = XLSX.read(buffer, { type: "array" });
  const sheetName = workbook.SheetNames[0];

  if (!sheetName) {
    throw new Error("File Excel tidak memiliki sheet.");
  }

  const sheet = workbook.Sheets[sheetName];
  const excelRows = XLSX.utils.sheet_to_json(sheet, {
    defval: "",
    raw: false,
  });

  if (!excelRows.length) {
    throw new Error("Sheet Excel kosong.");
  }

  return normalizeOperatorCtImportRows(excelRows);
}

export function buildOperatorCtTemplateRows(sourceRows = []) {
  const uniqueMap = new Map();

  (Array.isArray(sourceRows) ? sourceRows : []).forEach((row) => {
    upsertOperatorTemplateRow(uniqueMap, row, (source, { area, line, uuid }) => ({
      _hasOperatorInfo: hasTemplateOperatorInfo(source),
      Area: area,
      Line: line,
      UUID: uuid,
      NIK: formatTemplateOperatorNik(source),
      "Nama Operator": formatTemplateOperatorName(source),
      "CT SUM": "",
      "CT STD": "",
      CT: "",
    }));
  });

  return [...uniqueMap.values()]
    .map(({ _hasOperatorInfo, ...entry }) => entry)
    .sort((a, b) => {
    const lineCmp = String(a.Line).localeCompare(String(b.Line), "id", {
      numeric: true,
      sensitivity: "base",
    });
    if (lineCmp !== 0) return lineCmp;
    return String(a.UUID).localeCompare(String(b.UUID));
  });
}

function safeFilePart(value) {
  return String(value || "all")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "") || "all";
}

export function downloadOperatorCtTemplate(
  sourceRows = [],
  { locationFilter = "ALL" } = {}
) {
  const rows = buildOperatorCtTemplateRows(sourceRows);

  if (!rows.length) {
    throw new Error(
      "Tidak ada data mesin untuk template. Sesuaikan filter Area/tanggal lalu coba lagi."
    );
  }

  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(rows);

  sheet["!cols"] = [
    { wch: 8 },
    { wch: 18 },
    { wch: 22 },
    { wch: 12 },
    { wch: 28 },
    { wch: 10 },
    { wch: 10 },
    { wch: 8 },
  ];

  XLSX.utils.book_append_sheet(workbook, sheet, "DATA");
  XLSX.writeFile(
    workbook,
    `template-ct-operator-${safeFilePart(locationFilter)}.xlsx`
  );

  return {
    rowCount: rows.length,
    fileName: `template-ct-operator-${safeFilePart(locationFilter)}.xlsx`,
  };
}
