import * as XLSX from "xlsx";

export function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

export function getVal(obj, ...keys) {
  for (const key of keys) {
    if (obj && obj[key] !== undefined && obj[key] !== null) {
      return obj[key];
    }
  }

  return undefined;
}

export function extractRows(data) {
  if (Array.isArray(data)) return data;

  return (
    data?.rows ||
    data?.Rows ||
    data?.data ||
    data?.Data ||
    data?.items ||
    data?.Items ||
    data?.machines ||
    data?.Machines ||
    []
  );
}

export function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

export function getLocationGroup(location) {
  const text = String(location || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");

  const match = text.match(/\bGM\s*([0-9]+)\b/);

  if (match) {
    return `GM${match[1]}`;
  }

  return text || "-";
}

export function formatExcelDate(dateText) {
  const text = String(dateText || "").trim();

  if (!text) return "";

  const parts = text.split("-");

  if (parts.length === 3) {
    const year = parts[0];
    const month = Number(parts[1]);
    const day = Number(parts[2]);

    return `${day}/${month}/${year}`;
  }

  return text;
}

export function formatSeconds(seconds) {
  const totalSeconds = Math.max(0, Math.floor(Number(seconds || 0)));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  return `${hours}h ${minutes}m`;
}

export function formatExportDuration(seconds) {
  const totalSeconds = Math.max(0, Math.floor(Number(seconds || 0)));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const secs = totalSeconds % 60;

  return [hours, minutes, secs]
    .map((item) => String(item).padStart(2, "0"))
    .join(":");
}

export function parseExportDateTime(value) {
  if (!value) return null;

  const raw = String(value || "").trim();

  if (!raw) return null;

  const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
  const d = new Date(normalized);

  if (Number.isNaN(d.getTime())) return null;

  return d;
}

export function formatExportTime(value) {
  const d = parseExportDateTime(value);

  if (!d) {
    const text = String(value || "").replace("T", " ");
    return text.slice(11, 16).replace(":", ".");
  }

  return d
    .toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    })
    .replace(":", ".");
}

export function isAutoLogoutText(value) {
  const text = String(value || "")
    .trim()
    .toUpperCase()
    .replace(/[_-]+/g, " ");

  return (
    text.includes("AUTO LOGOUT") ||
    text.includes("AUTOLOGOUT") ||
    text.includes("AUTO_LOGOUT")
  );
}

export function isSystemDurationNote(note, durationText) {
  const text = String(note || "")
    .trim()
    .toUpperCase();

  const duration = String(durationText || "")
    .trim()
    .toUpperCase();

  if (!text) return false;

  if (duration && text === `SELESAI ${duration}`) return true;
  if (duration && text === `SEDANG BERJALAN ${duration}`) return true;

  return (
    /^SELESAI\s+\d{2}:\d{2}:\d{2}$/.test(text) ||
    /^SEDANG BERJALAN\s+\d{2}:\d{2}:\d{2}$/.test(text)
  );
}

export function normalizeExportNote(row) {
  const reasonName = String(
    getVal(
      row,
      "reasonName",
      "ReasonName",
      "reasonLabel",
      "ReasonLabel",
      "reason_name",
      "reason_label"
    ) || ""
  ).trim();

  const note = String(getVal(row, "note", "Note") || "").trim();

  const createdAt = String(
    getVal(
      row,
      "createdAt",
      "CreatedAt",
      "created_at",
      "startTime",
      "StartTime",
      "start_time"
    ) || ""
  ).trim();

  const endTime = String(
    getVal(row, "endTime", "EndTime", "end_time") || ""
  ).trim();

  const status = String(getVal(row, "status", "Status") || "").trim();

  if (
    isAutoLogoutText(status) ||
    isAutoLogoutText(reasonName) ||
    isAutoLogoutText(note)
  ) {
    return "";
  }

  const durationSeconds = toNumber(
    getVal(
      row,
      "durationSeconds",
      "DurationSeconds",
      "duration_sec",
      "duration"
    ) || 0
  );

  const durationText =
    String(
      getVal(row, "durationText", "DurationText", "duration_text") || ""
    ).trim() || formatExportDuration(durationSeconds);

  const startClock = formatExportTime(createdAt);
  const endClock = formatExportTime(endTime);
  const statusUpper = String(status || "").toUpperCase();

  let noteText = "";

  if (statusUpper === "CLOSED" || endTime) {
    noteText = `Selesai ${durationText}`;
  } else {
    noteText = `Sedang berjalan ${durationText}`;
  }

  if (
    note &&
    !isSystemDurationNote(note, durationText) &&
    !isAutoLogoutText(note)
  ) {
    noteText = `${noteText} | ${note}`;
  }

  if (startClock && endClock) {
    return `${startClock}-${endClock} ${reasonName}: ${noteText}`;
  }

  if (startClock) {
    return `${startClock}-${reasonName}: ${noteText}`;
  }

  if (reasonName) {
    return `${reasonName}: ${noteText}`;
  }

  return noteText;
}

export function buildOperatorExportMap(reportData) {
  const tempMap = new Map();

  extractRows(reportData).forEach((row) => {
    const uuid = String(getVal(row, "uuid", "UUID") || "").trim();

    if (!uuid) return;

    const rowStatus = String(getVal(row, "status", "Status") || "").trim();
    const key = normalizeText(uuid);

    if (!tempMap.has(key)) {
      tempMap.set(key, {
        operatorRows: [],
        rowKeySet: new Set(),
      });
    }

    const current = tempMap.get(key);

    const operatorNik = String(
      getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
    ).trim();

    const operatorName = String(
      getVal(row, "operatorName", "OperatorName", "operator_name") || ""
    ).trim();

    const operator =
      operatorNik && operatorName
        ? `${operatorNik} - ${operatorName}`
        : operatorName || operatorNik || "";

    const notes = extractRows(
      getVal(row, "notes", "Notes", "lastNotes", "LastNotes") || []
    )
      .map(normalizeExportNote)
      .filter(Boolean);

    const activeLossReasonCode = String(
      getVal(
        row,
        "activeLossReasonCode",
        "ActiveLossReasonCode",
        "active_loss_reason_code"
      ) || ""
    ).trim();

    const activeLossReasonLabel = String(
      getVal(
        row,
        "activeLossReasonLabel",
        "ActiveLossReasonLabel",
        "active_loss_reason_label"
      ) || ""
    ).trim();

    const activeLossStartTime = String(
      getVal(
        row,
        "activeLossStartTime",
        "ActiveLossStartTime",
        "active_loss_start_time"
      ) || ""
    ).trim();

    const activeLossDurationSeconds = toNumber(
      getVal(
        row,
        "activeLossDurationSeconds",
        "ActiveLossDurationSeconds",
        "active_loss_duration_seconds"
      ) || 0
    );

    if (
      activeLossReasonLabel &&
      !isAutoLogoutText(activeLossReasonLabel) &&
      !isAutoLogoutText(rowStatus) &&
      !notes.length
    ) {
      notes.push(
        normalizeExportNote({
          reasonCode: activeLossReasonCode,
          reasonName: activeLossReasonLabel,
          createdAt: activeLossStartTime,
          durationSeconds: activeLossDurationSeconds,
          status: "ACTIVE",
        })
      );
    }

    if (!notes.length && operator && !isAutoLogoutText(rowStatus)) {
      notes.push("");
    }

    notes.forEach((note) => {
      const cleanNote = String(note || "").trim();

      if (isAutoLogoutText(cleanNote)) {
        return;
      }

      const rowKey = `${operator}||${cleanNote}`;

      if (current.rowKeySet.has(rowKey)) {
        return;
      }

      current.rowKeySet.add(rowKey);
      current.operatorRows.push({
        operator,
        note: cleanNote,
      });
    });
  });

  const map = new Map();

  tempMap.forEach((value, key) => {
    map.set(key, {
      operatorRows: value.operatorRows,
    });
  });

  return map;
}

export function statusFromProductivity(productivity) {
  const value = Number(productivity || 0);

  if (value >= 90) return "GOOD";
  if (value >= 80) return "NORMAL";
  return "BAD";
}

export function toIsoDate(date) {
  const offset = date.getTimezoneOffset();
  const local = new Date(date.getTime() - offset * 60000);

  return local.toISOString().slice(0, 10);
}

export function getOrderedRange(startValue, endValue, fallbackValue = "") {
  const start = String(startValue || fallbackValue || "").trim();
  const end = String(endValue || start || "").trim();

  if (!start && !end) {
    return {
      start: "",
      end: "",
    };
  }

  if (start && end && end < start) {
    return {
      start: end,
      end: start,
    };
  }

  return {
    start,
    end: end || start,
  };
}

export function getWeekdaysBetween(startDateText, endDateText) {
  const start = new Date(`${startDateText}T00:00:00`);
  const end = new Date(`${endDateText}T00:00:00`);
  const dates = [];

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return dates;
  }

  const cursor = new Date(start);

  while (cursor <= end) {
    const day = cursor.getDay();

    if (day >= 1 && day <= 5) {
      dates.push(toIsoDate(cursor));
    }

    cursor.setDate(cursor.getDate() + 1);
  }

  return dates;
}

export function getDayName(dateText) {
  const d = new Date(`${dateText}T00:00:00`);

  return d.toLocaleDateString("id-ID", {
    weekday: "long",
  });
}

export function setWorksheetWidth(worksheet, widths) {
  worksheet["!cols"] = widths.map((wch) => ({ wch }));
}

export function setAutoFilter(worksheet, rowCount, colCount) {
  if (!rowCount || !colCount) return;

  const lastCol = XLSX.utils.encode_col(colCount - 1);
  const lastRow = rowCount + 1;

  worksheet["!autofilter"] = {
    ref: `A1:${lastCol}${lastRow}`,
  };
}

export function normalizeRangeItem(
  row,
  dateText,
  operatorMap = new Map(),
  getManualSettingByUuid = () => null
) {
  const uuid = String(getVal(row, "uuid", "UUID") || "");
  const setting = getManualSettingByUuid(uuid);

  const procTime = toNumber(
    getVal(
      row,
      "procSec",
      "ProcSec",
      "procTimeSec",
      "ProcTimeSec",
      "procTime",
      "ProcTime",
      "process_time",
      "ProcessTime"
    ) || 0
  );

  let runtime = toNumber(
    getVal(
      row,
      "runtimeSec",
      "RuntimeSec",
      "runtime",
      "Runtime",
      "RunTime",
      "runTime"
    ) || 0
  );

  if (procTime > runtime) {
    runtime = procTime;
  }

  const lossTime = Math.max(0, runtime - procTime);
  const productivity =
    runtime > 0 ? Math.min((procTime / runtime) * 100, 100) : 0;

  const backendName = String(
    getVal(
      row,
      "nickName",
      "NickName",
      "machineName",
      "MachineName",
      "name",
      "Name"
    ) || uuid
  );

  const machineName = setting?.customName || backendName;

  const locationFromApi = String(getVal(row, "location", "Location") || "");
  const location = setting?.location || locationFromApi || "-";

  const operatorInfo = operatorMap.get(normalizeText(uuid)) || {};
  const operatorRows = Array.isArray(operatorInfo.operatorRows)
    ? operatorInfo.operatorRows
    : [];

  return {
    dateText,
    tanggal: formatExcelDate(dateText),
    hari: getDayName(dateText),
    mesin: machineName,
    location,
    locationGroup: getLocationGroup(location),
    operatorRows,
    runtime,
    procTime,
    lossTime,
    productivity: Number(productivity.toFixed(2)),
    status: statusFromProductivity(productivity),
  };
}

export function filterRangeItemsByLocation(items, selectedLocation) {
  const selected = String(selectedLocation || "ALL").trim();

  if (selected === "ALL") {
    return items;
  }

  return items.filter((item) => {
    return getLocationGroup(item.location) === selected;
  });
}

export function safeFileName(value) {
  return String(value || "")
    .trim()
    .replace(/[\\/:*?"<>|]/g, "-")
    .replace(/\s+/g, "_");
}

function makeBlankRow(baseRow) {
  const blank = {};

  Object.keys(baseRow).forEach((key) => {
    blank[key] = "";
  });

  return blank;
}

export function expandOperatorRows(baseRow, operatorRows) {
  const rows = Array.isArray(operatorRows) ? operatorRows : [];

  if (!rows.length) {
    return [
      {
        ...baseRow,
        Operator: "",
        "Operator Note": "",
      },
    ];
  }

  let lastOperator = "";

  return rows.map((item, index) => {
    const operator = String(item.operator || "").trim();
    const note = String(item.note || "").trim();

    const row = index === 0 ? { ...baseRow } : makeBlankRow(baseRow);

    row.Operator = operator && operator !== lastOperator ? operator : "";
    row["Operator Note"] = note;

    if (operator) {
      lastOperator = operator;
    }

    return row;
  });
}

export function buildRangeDetailRows(items) {
  return items
    .slice()
    .sort((a, b) => {
      const dateA = String(a.tanggal || "");
      const dateB = String(b.tanggal || "");
      const areaCompare = String(a.locationGroup).localeCompare(
        String(b.locationGroup)
      );
      const locCompare = String(a.location).localeCompare(String(b.location));

      if (dateA !== dateB) return dateA.localeCompare(dateB);
      if (areaCompare !== 0) return areaCompare;
      if (locCompare !== 0) return locCompare;

      return String(a.mesin).localeCompare(String(b.mesin));
    })
    .flatMap((item) => {
      const baseRow = {
        Tanggal: item.tanggal,
        Hari: item.hari,
        Area: getLocationGroup(item.location),
        Mesin: item.mesin,
        Location: item.location,
        "Power On Duration": formatSeconds(item.runtime),
        "Running Time": formatSeconds(item.procTime),
        "Loss Time": formatSeconds(item.lossTime),
        Produktivitas: item.productivity,
        Status: item.status,
      };

      return expandOperatorRows(baseRow, item.operatorRows);
    });
}

export function buildRangeSummaryBaseRows(items, rangeStart, rangeEnd) {
  const map = new Map();

  items.forEach((item) => {
    const area = getLocationGroup(item.location);
    const key = `${area}||${item.location}||${item.mesin}`;

    if (!map.has(key)) {
      map.set(key, {
        periode: `${formatExcelDate(rangeStart)} - ${formatExcelDate(rangeEnd)}`,
        area,
        location: item.location,
        mesin: item.mesin,
        totalPowerOn: 0,
        totalRunning: 0,
        totalLoss: 0,
        operatorRows: [],
        operatorRowKeySet: new Set(),
      });
    }

    const current = map.get(key);

    current.totalPowerOn += Number(item.runtime || 0);
    current.totalRunning += Number(item.procTime || 0);
    current.totalLoss += Number(item.lossTime || 0);

    const operatorRows = Array.isArray(item.operatorRows)
      ? item.operatorRows
      : [];

    operatorRows.forEach((operatorRow) => {
      const operator = String(operatorRow.operator || "").trim();
      const note = String(operatorRow.note || "").trim();

      if (isAutoLogoutText(operator) || isAutoLogoutText(note)) {
        return;
      }

      const rowKey = `${operator}||${note}`;

      if (current.operatorRowKeySet.has(rowKey)) {
        return;
      }

      current.operatorRowKeySet.add(rowKey);
      current.operatorRows.push({
        operator,
        note,
      });
    });
  });

  return [...map.values()]
    .map((item) => {
      const productivity =
        item.totalPowerOn > 0
          ? Math.min((item.totalRunning / item.totalPowerOn) * 100, 100)
          : 0;

      return {
        periode: item.periode,
        area: item.area,
        location: item.location,
        mesin: item.mesin,
        totalPowerOn: item.totalPowerOn,
        totalRunning: item.totalRunning,
        totalLoss: item.totalLoss,
        productivity: Number(productivity.toFixed(2)),
        status: statusFromProductivity(productivity),
        operatorRows: item.operatorRows,
        __lossSec: item.totalLoss,
      };
    })
    .sort((a, b) => {
      const areaCompare = String(a.area).localeCompare(String(b.area));
      const locCompare = String(a.location).localeCompare(String(b.location));

      if (areaCompare !== 0) return areaCompare;
      if (locCompare !== 0) return locCompare;

      return Number(a.productivity || 0) - Number(b.productivity || 0);
    });
}

export function buildRangeSummaryRows(summaryBaseRows) {
  return summaryBaseRows.flatMap((item) => {
    const baseRow = {
      Periode: item.periode,
      Area: item.area,
      Location: item.location,
      Mesin: item.mesin,
      "Total Power On Duration": formatSeconds(item.totalPowerOn),
      "Total Running Time": formatSeconds(item.totalRunning),
      "Total Loss Time": formatSeconds(item.totalLoss),
      Produktivitas: item.productivity,
      Status: item.status,
    };

    return expandOperatorRows(baseRow, item.operatorRows);
  });
}

export function buildBadPriorityRows(summaryBaseRows, selectedLocation) {
  const badRows = summaryBaseRows
    .filter((row) => row.status === "BAD")
    .sort((a, b) => Number(a.productivity || 0) - Number(b.productivity || 0));

  if (!badRows.length) {
    return [
      {
        Rank: "",
        Area: selectedLocation === "ALL" ? "All GM" : selectedLocation,
        Location: "",
        Mesin: "Tidak ada mesin BAD",
        Produktivitas: "",
        "Total Loss Time": "",
        Rekomendasi: "Semua mesin tidak berstatus BAD pada periode ini.",
        Operator: "",
        "Operator Note": "",
      },
    ];
  }

  return badRows.flatMap((row, index) => {
    let rekomendasi = "Cek penyebab running time rendah.";

    if (Number(row.__lossSec || 0) > 14400) {
      rekomendasi = "Loss time tinggi. Cek idle, operator, dan kondisi mesin.";
    }

    const baseRow = {
      Rank: index + 1,
      Area: row.area,
      Location: row.location,
      Mesin: row.mesin,
      Produktivitas: row.productivity,
      "Total Loss Time": formatSeconds(row.totalLoss),
      Rekomendasi: rekomendasi,
    };

    return expandOperatorRows(baseRow, row.operatorRows);
  });
}

export function createRangeWorkbook(summaryRows, detailRows, badPriorityRows) {
  const workbook = XLSX.utils.book_new();

  const summarySheet = XLSX.utils.json_to_sheet(summaryRows);
  setWorksheetWidth(summarySheet, [
    24,
    12,
    18,
    42,
    24,
    22,
    18,
    16,
    12,
    32,
    90,
  ]);
  setAutoFilter(summarySheet, summaryRows.length, 11);

  const detailSheet = XLSX.utils.json_to_sheet(detailRows);
  setWorksheetWidth(detailSheet, [
    12,
    14,
    12,
    42,
    18,
    22,
    18,
    16,
    16,
    12,
    32,
    90,
  ]);
  setAutoFilter(detailSheet, detailRows.length, 12);

  const badSheet = XLSX.utils.json_to_sheet(badPriorityRows);
  setWorksheetWidth(badSheet, [8, 12, 18, 42, 16, 18, 48, 32, 90]);
  setAutoFilter(badSheet, badPriorityRows.length, 9);

  XLSX.utils.book_append_sheet(workbook, summarySheet, "Range Summary");
  XLSX.utils.book_append_sheet(workbook, detailSheet, "Daily Detail");
  XLSX.utils.book_append_sheet(workbook, badSheet, "BAD Priority");

  return workbook;
}

export function writeWorkbook(workbook, fileName) {
  XLSX.writeFile(workbook, fileName);
}