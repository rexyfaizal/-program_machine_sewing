import * as XLSX from "xlsx";
import { formatDurationHHMMSS } from "./format";

function safeFileName(value) {
  return String(value || "")
    .trim()
    .replace(/[\\/:*?"<>|]/g, "-")
    .replace(/\s+/g, "_")
    .slice(0, 80);
}

function parseNotePctValue(value) {
  const text = String(value || "")
    .trim()
    .replace(/%/g, "");
  const num = Number(text);
  return Number.isFinite(num) ? num : 0;
}

function formatExportNotePct(value) {
  if (!String(value || "").trim()) return "";
  if (parseNotePctValue(value) > 100) return "> 100%";
  return String(value).trim();
}

function buildRows(list, averages) {
  const rows = (Array.isArray(list) ? list : []).map((row) => ({
    Area: row.area || "-",
    Location: row.locationLabel || "-",
    UUID: row.uuid || "-",
    "Nama Operator": row.operatorName || "-",
    NIK: row.operatorNik || "-",
    Shift: row.shiftTag || "Normal",
    Mesin: row.mesin || "-",
    Style: row.style || "-",
    Output: Number(row.output || 0),
    "Avg Proses": Number(Number(row.avgCycle || 0).toFixed(2)),
    "Mesin Menyala": row.powerOnText || "00:00:00",
    "Mesin Bekerja": row.processText || "00:00:00",
    "Waktu Mesin Terbuang": row.lossText || "00:00:00",
    Produktivitas: Number(Number(row.productivity || 0).toFixed(2)),
    "Tunggu bahan": row.tungguBahanText || "00:00:00",
    "% Tunggu bahan": formatExportNotePct(row.tungguBahanPct),
    "Mesin Rusak": row.mesinRusakText || "00:00:00",
    "% Mesin Rusak": formatExportNotePct(row.mesinRusakPct),
    "Ke Toilet": row.toiletText || "00:00:00",
    "% Toilet": formatExportNotePct(row.toiletPct),
    Solat: row.solatText || "00:00:00",
    "% Solat": formatExportNotePct(row.solatPct),
    Others: row.otherText || "00:00:00",
    "% Others": formatExportNotePct(row.otherPct),
    Remarks: row.remarks === "-" ? "" : row.remarks,
  }));

  rows.push({
    Area: "AVERAGE",
    Location: "",
    UUID: "",
    "Nama Operator": "",
    NIK: "",
    Shift: "",
    Mesin: "",
    Style: "",
    Output: Number(averages?.output || 0),
    "Avg Proses": Number(averages?.avgCycle || 0),
    "Mesin Menyala": formatDurationHHMMSS(averages?.runtimeSec || 0),
    "Mesin Bekerja": formatDurationHHMMSS(averages?.procSec || 0),
    "Waktu Mesin Terbuang": formatDurationHHMMSS(averages?.lossTimeSec || 0),
    Produktivitas: Number(averages?.productivity || 0),
    "Tunggu bahan": formatDurationHHMMSS(averages?.tungguBahanSec || 0),
    "% Tunggu bahan": formatExportNotePct(averages?.tungguBahanPct),
    "Mesin Rusak": formatDurationHHMMSS(averages?.mesinRusakSec || 0),
    "% Mesin Rusak": formatExportNotePct(averages?.mesinRusakPct),
    "Ke Toilet": formatDurationHHMMSS(averages?.toiletSec || 0),
    "% Toilet": formatExportNotePct(averages?.toiletPct),
    Solat: formatDurationHHMMSS(averages?.solatSec || 0),
    "% Solat": formatExportNotePct(averages?.solatPct),
    Others: formatDurationHHMMSS(averages?.otherSec || 0),
    "% Others": formatExportNotePct(averages?.otherPct),
    Remarks: "",
  });

  return rows;
}

export function exportOperatorProductivityExcel({
  rows = [],
  averages = {},
  date = "",
} = {}) {
  const list = Array.isArray(rows) ? rows : [];
  if (!list.length) {
    throw new Error("Tidak ada data operator untuk diexport.");
  }

  const sheetRows = buildRows(list, averages);
  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(sheetRows);

  sheet["!cols"] = [
    { wch: 8 },
    { wch: 28 },
    { wch: 20 },
    { wch: 28 },
    { wch: 14 },
    { wch: 12 },
    { wch: 28 },
    { wch: 10 },
    { wch: 12 },
    { wch: 16 },
    { wch: 16 },
    { wch: 20 },
    { wch: 14 },
    { wch: 14 },
    { wch: 12 },
    { wch: 14 },
    { wch: 12 },
    { wch: 12 },
    { wch: 10 },
    { wch: 12 },
    { wch: 10 },
    { wch: 12 },
    { wch: 10 },
    { wch: 36 },
  ];

  XLSX.utils.book_append_sheet(workbook, sheet, "Produktivitas Operator");

  const datePart = safeFileName(date || "tanggal");
  const fileName = `produktivitas-operator-${datePart}.xlsx`;
  XLSX.writeFile(workbook, fileName);

  return {
    fileName,
    rowCount: list.length,
  };
}
