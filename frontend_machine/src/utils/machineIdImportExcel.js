import * as XLSX from "xlsx";

function normalizeHeader(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/[_-]+/g, " ")
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
  return String(value ?? "")
    .trim()
    .replace(/\u00a0/g, " ")
    .replace(/\s+/g, " ");
}

export function normalizeMachineIdImportRows(excelRows) {
  const uniqueMap = new Map();
  const duplicateRows = [];
  const errorRows = [];

  let skippedEmpty = 0;
  let skippedDuplicate = 0;
  let skippedInvalid = 0;

  (Array.isArray(excelRows) ? excelRows : []).forEach((row, index) => {
    const excelRowNumber = index + 2;
    const uuid = cleanText(getCellValue(row, "UUID", "MESIN UUID", "MACHINE UUID"));
    const machineId = cleanText(
      getCellValue(row, "MACHINE ID", "MACHINEID", "MESIN ID", "ID MESIN", "ID MACHINE")
    );

    if (!uuid && !machineId) {
      skippedEmpty += 1;
      return;
    }

    if (!uuid || !machineId) {
      skippedInvalid += 1;
      errorRows.push({
        excelRowNumber,
        message: "UUID dan Machine ID wajib diisi.",
        uuid,
        machineId,
      });
      return;
    }

    const key = uuid.toLowerCase();
    if (uniqueMap.has(key)) {
      skippedDuplicate += 1;
      duplicateRows.push({
        excelRowNumber,
        duplicateOfRowNumber: uniqueMap.get(key).excelRowNumber,
        uuid,
        machineId,
      });
    }

    uniqueMap.set(key, {
      uuid,
      machineId,
      excelRowNumber,
    });
  });

  const rows = [...uniqueMap.values()];

  return {
    rows,
    duplicateRows,
    errorRows,
    stats: {
      totalExcelRows: Array.isArray(excelRows) ? excelRows.length : 0,
      readyRows: rows.length,
      skippedEmpty,
      skippedDuplicate,
      skippedInvalid,
    },
  };
}

export async function parseMachineIdImportFile(file) {
  const buffer = await file.arrayBuffer();
  const workbook = XLSX.read(buffer, { type: "array" });
  const sheetName = workbook.SheetNames[0];
  if (!sheetName) {
    throw new Error("Sheet Excel kosong.");
  }

  const sheet = workbook.Sheets[sheetName];
  const excelRows = XLSX.utils.sheet_to_json(sheet, { defval: "" });
  return normalizeMachineIdImportRows(excelRows);
}

export function downloadMachineIdTemplate(existingRows = []) {
  const rows = (Array.isArray(existingRows) ? existingRows : [])
    .map((item) => ({
      UUID: String(item?.uuid || item?.UUID || "").trim(),
      "Machine ID": String(item?.machineId || item?.MachineID || "").trim(),
    }))
    .filter((row) => row.UUID);

  const data =
    rows.length > 0
      ? rows
      : [
          { UUID: "21EXAMPLEUUID0001", "Machine ID": "JUKI-0001" },
          { UUID: "21EXAMPLEUUID0002", "Machine ID": "JUKI-0002" },
        ];

  const worksheet = XLSX.utils.json_to_sheet(data);
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "MachineID");
  XLSX.writeFile(workbook, "template-machine-id.xlsx");

  return { rowCount: data.length };
}
