import * as XLSX from "xlsx";
import { formatDurationHHMMSS } from "./format";

function safeFileName(value) {
  return String(value || "")
    .trim()
    .replace(/[\\/:*?"<>|]/g, "-")
    .replace(/\s+/g, "_")
    .slice(0, 80);
}

function stitchText(event) {
  const endStitch = Number(event?.endStitch || 0);
  const fileStitches = Number(event?.fileStitches || 0);
  return `${endStitch}/${fileStitches}`;
}

function sumDurationSec(events, key) {
  return (Array.isArray(events) ? events : []).reduce((sum, event) => {
    if (event?.[key] == null) return sum;
    return sum + Math.max(0, Number(event[key]) || 0);
  }, 0);
}

export function buildProcessDetailExportRows(
  events = [],
  { uuid = "", location = "" } = {}
) {
  const machineUuid = String(uuid || "").trim();
  const machineLocation = String(location || "").trim() || "-";
  const list = Array.isArray(events) ? events : [];

  const rows = list.map((event, index) => ({
    No: index + 1,
    UUID: machineUuid || "-",
    "Nama Operator": String(event?.operatorName || "").trim() || "-",
    Location: machineLocation,
    Program: String(event?.fileName || "").trim() || "-",
    Start: String(event?.startTime || "").trim() || "-",
    End: String(event?.endTime || "").trim() || "-",
    "Jeda Waktu Antar Output":
      event?.gapSec == null ? "-" : String(event?.gapTime || "00:00:00"),
    "Detail Losstime":
      event?.lossTimeSec == null
        ? "-"
        : String(event?.detailLossTime || "-").trim() || "-",
    "Waktu Losstime":
      event?.lossTimeSec == null
        ? "-"
        : String(event?.lossTime || "00:00:00"),
    "Losstime (detik)":
      event?.lossTimeSec == null ? "" : Number(event.lossTimeSec),
    "Proc Time": String(event?.procTime || "00:00"),
    "Proc (detik)": Number(event?.procSec || 0),
    Count: Number(event?.procCounts || 0),
    Stitch: stitchText(event),
    "Node Distance": Number(event?.nodeDistance || 0),
    SPM: Number(event?.spm || 0),
    Status: String(event?.status || "ABNORMAL"),
    Reason: String(event?.abnormalReason || "").trim() || "-",
  }));

  const totalGapSec = sumDurationSec(list, "gapSec");
  const totalLossSec = sumDurationSec(list, "lossTimeSec");

  rows.push({
    No: "",
    UUID: "",
    "Nama Operator": "",
    Location: "",
    Program: "TOTAL",
    Start: "",
    End: "",
    "Jeda Waktu Antar Output": formatDurationHHMMSS(totalGapSec),
    "Detail Losstime": "",
    "Waktu Losstime": formatDurationHHMMSS(totalLossSec),
    "Losstime (detik)": totalLossSec,
    "Proc Time": "",
    "Proc (detik)": "",
    Count: "",
    Stitch: "",
    "Node Distance": "",
    SPM: "",
    Status: "",
    Reason: "",
  });

  return rows;
}

export function exportProcessDetailExcel({
  events = [],
  machineName = "",
  uuid = "",
  location = "",
  date = "",
} = {}) {
  const rows = buildProcessDetailExportRows(events, { uuid, location });

  if (!rows.length) {
    throw new Error("Tidak ada data proses untuk diexport.");
  }

  const workbook = XLSX.utils.book_new();
  const sheet = XLSX.utils.json_to_sheet(rows);

  sheet["!cols"] = [
    { wch: 6 },
    { wch: 22 },
    { wch: 22 },
    { wch: 18 },
    { wch: 28 },
    { wch: 20 },
    { wch: 20 },
    { wch: 22 },
    { wch: 36 },
    { wch: 14 },
    { wch: 14 },
    { wch: 12 },
    { wch: 12 },
    { wch: 8 },
    { wch: 12 },
    { wch: 14 },
    { wch: 10 },
    { wch: 12 },
    { wch: 28 },
  ];

  XLSX.utils.book_append_sheet(workbook, sheet, "Detail Output");

  const namePart = safeFileName(machineName || uuid || "mesin");
  const datePart = safeFileName(date || "tanggal");
  const fileName = `detail-output-${namePart}-${datePart}.xlsx`;

  XLSX.writeFile(workbook, fileName);

  return {
    fileName,
    rowCount: Math.max(0, rows.length - 1),
  };
}
