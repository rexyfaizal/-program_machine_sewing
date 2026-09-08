import * as XLSX from "xlsx";

import { normalizeCtStylePart } from "./operatorCtStyle";

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

function cleanText(value) {
  return String(value || "")
    .trim()
    .replace(/\u00a0/g, " ")
    .replace(/\s+/g, " ");
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

export function normalizeOperatorCtStyleImportRows(excelRows) {
  const uniqueMap = new Map();
  const duplicateRows = [];
  const errorRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;
  let skippedInvalid = 0;

  (Array.isArray(excelRows) ? excelRows : []).forEach((row, index) => {
    const excelRowNumber = index + 2;
    const styleName = cleanText(
      getCellValue(row, "STYLE", "STYLE NAME", "NAMA STYLE")
    );
    const processName = cleanText(
      getCellValue(row, "PROSES", "PROCESS", "NAMA PROSES", "PROCESS NAME")
    );
    const ctTotal = parseCtNumber(
      getCellValue(row, "CT TOTAL", "CT TOTAL", "CT SUM", "CT")
    );

    if (!styleName && !processName) {
      skippedEmpty += 1;
      return;
    }

    if (!styleName || !processName) {
      skippedInvalid += 1;
      errorRows.push({
        excelRowNumber,
        message: "Style dan Proses wajib diisi.",
        styleName,
        processName,
      });
      return;
    }

    const key = normalizeCtStylePart(styleName) + "||" + normalizeCtStylePart(processName);
    if (uniqueMap.has(key)) {
      skippedDuplicate += 1;
      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: uniqueMap.get(key).excelRowNumber,
        styleName,
        processName,
      });
    }

    uniqueMap.set(key, {
      styleName,
      processName,
      ctTotal,
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

export async function parseOperatorCtStyleExcel(file) {
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

  return normalizeOperatorCtStyleImportRows(excelRows);
}

export function buildOperatorCtStyleTemplateRows(sourceRows = []) {
  const uniqueMap = new Map();

  (Array.isArray(sourceRows) ? sourceRows : []).forEach((row) => {
    const area = String(row?.area || "").trim().toUpperCase();
    if (area !== "GM3") return;

    const styleName = cleanText(row?.style);
    const processName = cleanText(row?.mesin);
    if (!styleName || !processName || styleName === "-" || processName === "-") {
      return;
    }

    const key = normalizeCtStylePart(styleName) + "||" + normalizeCtStylePart(processName);
    if (uniqueMap.has(key)) return;

    uniqueMap.set(key, {
      Style: styleName,
      Proses: processName,
      "CT Total": "",
    });
  });

  return [...uniqueMap.values()].sort((a, b) => {
    const styleCmp = String(a.Style).localeCompare(String(b.Style), "id", {
      numeric: true,
      sensitivity: "base",
    });
    if (styleCmp !== 0) return styleCmp;
    return String(a.Proses).localeCompare(String(b.Proses), "id", {
      numeric: true,
      sensitivity: "base",
    });
  });
}

export function downloadOperatorCtStyleTemplate(sourceRows = []) {
  let rows = buildOperatorCtStyleTemplateRows(sourceRows);

  if (!rows.length) {
    rows = [
      {
        Style: "1101723",
        Proses: "Quilting Back Panel",
        "CT Total": "",
      },
      {
        Style: "1101723",
        Proses: "Quilting Hood Mid",
        "CT Total": "",
      },
    ];
  }

  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(rows);

  sheet["!cols"] = [{ wch: 14 }, { wch: 32 }, { wch: 12 }];

  XLSX.utils.book_append_sheet(workbook, sheet, "DATA");
  XLSX.writeFile(workbook, "template-ct-style-proses-gm3.xlsx");

  return {
    rowCount: rows.length,
    fileName: "template-ct-style-proses-gm3.xlsx",
  };
}
