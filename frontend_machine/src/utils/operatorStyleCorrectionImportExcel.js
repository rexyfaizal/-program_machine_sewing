import * as XLSX from "xlsx";

import { normalizeCtLine } from "./operatorCt";
import { parseWorkDate } from "./operatorOutputTargetImportExcel";
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

function cleanText(value) {
  return String(value || "")
    .trim()
    .replace(/\u00a0/g, " ")
    .replace(/\s+/g, " ");
}

export function normalizeOperatorStyleCorrectionImportRows(
  excelRows,
  fallbackDate = ""
) {
  const uniqueMap = new Map();
  const duplicateRows = [];
  const errorRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;
  let skippedInvalid = 0;

  (Array.isArray(excelRows) ? excelRows : []).forEach((row, index) => {
    const excelRowNumber = index + 2;
    const workDate =
      parseWorkDate(getCellValue(row, "TANGGAL", "DATE", "WORK DATE")) ||
      parseWorkDate(fallbackDate);
    const line = normalizeCtLine(
      getCellValue(row, "LINE", "LOCATION", "LOKASI")
    );
    const uuid = cleanText(getCellValue(row, "UUID"));
    const operatorNik = cleanText(
      getCellValue(row, "NIK", "OPERATOR NIK", "NIK OPERATOR")
    );
    const previousStyle = cleanText(
      getCellValue(row, "STYLE LAMA", "STYLE SEBELUMNYA", "OLD STYLE")
    );
    const previousProcess = cleanText(
      getCellValue(row, "PROSES LAMA", "PROSES SEBELUMNYA", "OLD PROCESS")
    );
    const styleName = cleanText(
      getCellValue(row, "STYLE BARU", "STYLE", "STYLE NAME", "NAMA STYLE")
    );
    const processName = cleanText(
      getCellValue(
        row,
        "PROSES BARU",
        "PROSES",
        "PROCESS",
        "PROCESS NAME",
        "NAMA PROSES"
      )
    );

    if (!uuid && !styleName && !workDate) {
      skippedEmpty += 1;
      return;
    }

    if (!workDate || !uuid || !styleName) {
      skippedInvalid += 1;
      errorRows.push({
        excelRowNumber,
        message: "Tanggal, UUID, dan Style Baru wajib diisi.",
        workDate: workDate || "-",
        uuid,
        styleName,
      });
      return;
    }

    const key = `${workDate.toUpperCase()}||${uuid.toUpperCase()}||${operatorNik.toUpperCase()}`;
    if (uniqueMap.has(key)) {
      skippedDuplicate += 1;
      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: uniqueMap.get(key).excelRowNumber,
        workDate,
        uuid,
        styleName,
      });
      return;
    }

    uniqueMap.set(key, {
      excelRowNumber,
      workDate,
      line,
      uuid,
      operatorNik,
      previousStyle,
      previousProcess,
      styleName,
      processName,
    });
  });

  const result = Array.from(uniqueMap.values());

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

export async function parseOperatorStyleCorrectionExcel(file, fallbackDate = "") {
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
    raw: true,
  });

  if (!excelRows.length) {
    throw new Error("Sheet Excel kosong.");
  }

  return normalizeOperatorStyleCorrectionImportRows(excelRows, fallbackDate);
}

export function buildOperatorStyleCorrectionTemplateRows(
  sourceRows = [],
  defaultDate = ""
) {
  const workDate = parseWorkDate(defaultDate) || new Date().toISOString().slice(0, 10);
  const uniqueMap = new Map();

  (Array.isArray(sourceRows) ? sourceRows : []).forEach((row) => {
    if (!row?.loggedIn) return;

    upsertOperatorTemplateRow(uniqueMap, row, (source, { area, line, uuid }) => ({
      _hasOperatorInfo: hasTemplateOperatorInfo(source),
      Tanggal: workDate,
      Area: area,
      Line: line,
      UUID: uuid,
      NIK: formatTemplateOperatorNik(source),
      "Nama Operator": formatTemplateOperatorName(source),
      "Style Lama": cleanText(source?.style) || "",
      "Proses Lama": cleanText(source?.mesin) || "",
      "Style Baru": "",
      "Proses Baru": "",
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

export function downloadOperatorStyleCorrectionTemplate(
  sourceRows = [],
  { defaultDate = "", locationFilter = "ALL" } = {}
) {
  const workDate = parseWorkDate(defaultDate) || new Date().toISOString().slice(0, 10);
  const rows = buildOperatorStyleCorrectionTemplateRows(sourceRows, workDate);

  if (!rows.length) {
    throw new Error(
      "Tidak ada sesi login untuk template. Sesuaikan filter Area/tanggal lalu coba lagi."
    );
  }

  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(rows);

  sheet["!cols"] = [
    { wch: 12 },
    { wch: 8 },
    { wch: 16 },
    { wch: 22 },
    { wch: 12 },
    { wch: 24 },
    { wch: 14 },
    { wch: 28 },
    { wch: 14 },
    { wch: 28 },
  ];

  const fileName = `template-koreksi-style-${workDate}-${safeFilePart(locationFilter)}.xlsx`;
  XLSX.utils.book_append_sheet(workbook, sheet, "DATA");
  XLSX.writeFile(workbook, fileName);

  return {
    rowCount: rows.length,
    fileName,
  };
}
