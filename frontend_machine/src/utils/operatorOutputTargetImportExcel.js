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

function pad2(value) {
  return String(value).padStart(2, "0");
}

function formatDateParts(year, month, day) {
  if (!year || !month || !day) return "";
  return `${year}-${pad2(month)}-${pad2(day)}`;
}

function parseExcelSerialDate(serial) {
  const num = Number(serial);
  if (!Number.isFinite(num) || num <= 0) return "";

  const parsed = XLSX.SSF.parse_date_code(num);
  if (!parsed) return "";

  return formatDateParts(parsed.y, parsed.m, parsed.d);
}

export function parseWorkDate(value) {
  if (value === null || value === undefined || value === "") return "";

  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return formatDateParts(
      value.getFullYear(),
      value.getMonth() + 1,
      value.getDate()
    );
  }

  if (typeof value === "number") {
    return parseExcelSerialDate(value);
  }

  const text = String(value).trim();
  if (!text) return "";

  if (/^\d+(\.\d+)?$/.test(text)) {
    const fromSerial = parseExcelSerialDate(Number(text));
    if (fromSerial) return fromSerial;
  }

  const isoMatch = text.match(/^(\d{4})-(\d{1,2})-(\d{1,2})$/);
  if (isoMatch) {
    return formatDateParts(
      Number(isoMatch[1]),
      Number(isoMatch[2]),
      Number(isoMatch[3])
    );
  }

  const slashMatch = text.match(/^(\d{1,2})\/(\d{1,2})\/(\d{4})$/);
  if (slashMatch) {
    return formatDateParts(
      Number(slashMatch[3]),
      Number(slashMatch[2]),
      Number(slashMatch[1])
    );
  }

  const dashMatch = text.match(/^(\d{1,2})-(\d{1,2})-(\d{4})$/);
  if (dashMatch) {
    return formatDateParts(
      Number(dashMatch[3]),
      Number(dashMatch[2]),
      Number(dashMatch[1])
    );
  }

  return "";
}

function parseOutputTargetNumber(value) {
  const text = String(value ?? "")
    .trim()
    .replace(/,/g, "");
  if (!text) return 0;

  const num = Number(text);
  if (!Number.isFinite(num) || num < 0) return 0;
  return Math.round(num);
}

export function normalizeOperatorOutputTargetImportRows(excelRows, fallbackDate = "") {
  const uniqueMap = new Map();
  const duplicateRows = [];
  const errorRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;
  let skippedInvalid = 0;

  (Array.isArray(excelRows) ? excelRows : []).forEach((row, index) => {
    const excelRowNumber = index + 2;
    const workDate =
      parseWorkDate(getCellValue(row, "TANGGAL", "DATE", "WORK DATE", "WORK_DATE")) ||
      parseWorkDate(fallbackDate);
    const line = normalizeCtLine(
      getCellValue(row, "LINE", "LOCATION", "LOKASI")
    );
    const uuid = String(getCellValue(row, "UUID") || "").trim();
    const area = String(getCellValue(row, "AREA") || "").trim();
    const outputTarget = parseOutputTargetNumber(
      getCellValue(
        row,
        "OUTPUT TARGETAN",
        "OUTPUT TARGET",
        "TARGET",
        "TARGET OUTPUT"
      )
    );

    if (!uuid && !line && !workDate) {
      skippedEmpty += 1;
      return;
    }

    if (!workDate || !uuid || !line) {
      skippedInvalid += 1;
      errorRows.push({
        excelRowNumber,
        message: "Tanggal, UUID, dan Line wajib diisi.",
        workDate: workDate || "-",
        line,
        uuid,
      });
      return;
    }

    const key = `${workDate.toUpperCase()}||${uuid.toUpperCase()}||${line.toUpperCase()}`;
    if (uniqueMap.has(key)) {
      skippedDuplicate += 1;
      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: uniqueMap.get(key).excelRowNumber,
        workDate,
        line,
        uuid,
      });
      return;
    }

    uniqueMap.set(key, {
      excelRowNumber,
      workDate,
      line,
      uuid,
      area,
      outputTarget,
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

export async function parseOperatorOutputTargetExcel(file, fallbackDate = "") {
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

  return normalizeOperatorOutputTargetImportRows(excelRows, fallbackDate);
}

export function buildOperatorOutputTargetTemplateRows(
  sourceRows = [],
  defaultDate = ""
) {
  const workDate = parseWorkDate(defaultDate) || new Date().toISOString().slice(0, 10);
  const uniqueMap = new Map();

  (Array.isArray(sourceRows) ? sourceRows : []).forEach((row) => {
    upsertOperatorTemplateRow(uniqueMap, row, (source, { area, line, uuid }) => ({
      _hasOperatorInfo: hasTemplateOperatorInfo(source),
      Tanggal: workDate,
      Area: area,
      Line: line,
      UUID: uuid,
      NIK: formatTemplateOperatorNik(source),
      "Nama Operator": formatTemplateOperatorName(source),
      "Output Targetan": "",
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

export function downloadOperatorOutputTargetTemplate(
  sourceRows = [],
  { defaultDate = "", locationFilter = "ALL" } = {}
) {
  const workDate = parseWorkDate(defaultDate) || new Date().toISOString().slice(0, 10);
  const rows = buildOperatorOutputTargetTemplateRows(sourceRows, workDate);

  if (!rows.length) {
    throw new Error(
      "Tidak ada data mesin untuk template. Sesuaikan filter Area/tanggal lalu coba lagi."
    );
  }

  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(rows);

  sheet["!cols"] = [
    { wch: 12 },
    { wch: 8 },
    { wch: 18 },
    { wch: 22 },
    { wch: 12 },
    { wch: 28 },
    { wch: 16 },
  ];

  const fileName = `template-output-targetan-${workDate}-${safeFilePart(locationFilter)}.xlsx`;
  XLSX.utils.book_append_sheet(workbook, sheet, "DATA");
  XLSX.writeFile(workbook, fileName);

  return {
    rowCount: rows.length,
    fileName,
  };
}
