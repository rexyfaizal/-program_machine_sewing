import * as XLSX from "xlsx";
import {
  getExportShiftLabel,
  getGM3ShiftWindow,
  isDateInShiftWindow,
  resolveLineShiftForLocation,
} from "./gm3Shift";

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

export function formatExportUsageDuration(loginTime, logoutTime, status) {
  const start = parseExportDateTime(loginTime);

  if (!start) return "";

  const statusUpper = String(status || "").toUpperCase();
  const end = parseExportDateTime(logoutTime);
  const isActive = ["ACTIVE", "OPEN"].includes(statusUpper) && !end;
  const endMs = isActive ? Date.now() : end?.getTime();

  if (!endMs || endMs < start.getTime()) return "";

  const totalSeconds = Math.max(0, Math.floor((endMs - start.getTime()) / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  if (hours > 0) {
    return `${hours}j ${minutes}m`;
  }

  return `${minutes}m`;
}

/** Teks pemakaian operator seperti dashboard: Login 06.12–13.30 · Pakai 7j 18m */
export function formatOperatorUsageText(loginTime, logoutTime, status) {
  const loginClock = formatExportTime(loginTime);
  const logoutClock = formatExportTime(logoutTime);
  const usage = formatExportUsageDuration(loginTime, logoutTime, status);
  const statusUpper = String(status || "").toUpperCase();
  const isActive =
    ["ACTIVE", "OPEN"].includes(statusUpper) && !String(logoutTime || "").trim();

  if (loginClock && logoutClock && usage) {
    return `Login ${loginClock}–${logoutClock} · Pakai ${usage}`;
  }

  if (loginClock && isActive && usage) {
    return `Login ${loginClock} · Aktif ${usage}`;
  }

  if (loginClock && usage) {
    return `Login ${loginClock} · Pakai ${usage}`;
  }

  if (loginClock) {
    return `Login ${loginClock}`;
  }

  return "";
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

  let statusText = "";

  if (statusUpper === "CLOSED" || endTime) {
    statusText = `Selesai ${durationText}`;
  } else {
    statusText = `Sedang berjalan ${durationText}`;
  }

  // Bersihkan deskripsi bebas: buang teks sistem / duplikat reason / durasi dobel.
  let freeNote = note;
  const freeUpper = freeNote.toUpperCase();
  const reasonUpper = reasonName.toUpperCase();

  if (
    !freeNote ||
    isSystemDurationNote(freeNote, durationText) ||
    isAutoLogoutText(freeNote) ||
    freeUpper === reasonUpper ||
    freeUpper === `${reasonUpper} - SELESAI` ||
    freeUpper === `${reasonUpper} SELESAI` ||
    /^[A-Z0-9 \-_/]+ - SELESAI$/.test(freeUpper) ||
    /^SELESAI(\s+\d{2}:\d{2}:\d{2})?$/.test(freeUpper) ||
    /^SEDANG BERJALAN(\s+\d{2}:\d{2}:\d{2})?$/.test(freeUpper)
  ) {
    freeNote = "";
  } else {
    freeNote = freeNote
      .replace(/\s*-\s*Selesai\s+\d{2}:\d{2}:\d{2}\s*$/i, "")
      .replace(/\s+Selesai\s+\d{2}:\d{2}:\d{2}\s*$/i, "")
      .replace(/\s*-\s*Sedang berjalan\s+\d{2}:\d{2}:\d{2}\s*$/i, "")
      .trim();

    if (
      !freeNote ||
      freeNote.toUpperCase() === reasonUpper ||
      isSystemDurationNote(freeNote, durationText)
    ) {
      freeNote = "";
    }
  }

  let body = statusText;

  if (freeNote) {
    body = `${statusText} | ${freeNote}`;
  }

  const label = reasonName || "Other";

  if (startClock && endClock) {
    return `${startClock}-${endClock} ${label}: ${body}`;
  }

  if (startClock) {
    return `${startClock}-${label}: ${body}`;
  }

  if (reasonName) {
    return `${label}: ${body}`;
  }

  return body;
}

/**
 * Bangun map operator per UUID mesin.
 * options.workDate + options.shiftCode → filter note/session hanya yang overlap shift.
 * options.shiftConfigMap + options.getLocationByUuid → window per line.
 * ALL / kosong / tanpa window → seharian (perilaku lama).
 */
export function buildOperatorExportMap(reportData, options = {}) {
  const workDate = String(options.workDate || "").trim();
  const shiftCode = String(options.shiftCode || "ALL").trim().toUpperCase();
  const shiftConfigMap = options.shiftConfigMap || new Map();
  const getLocationByUuid =
    typeof options.getLocationByUuid === "function"
      ? options.getLocationByUuid
      : () => "";

  const resolveScheduleForUuid = (uuid) => {
    const location = getLocationByUuid(uuid);
    const resolved = resolveLineShiftForLocation(location, shiftConfigMap);
    if (!resolved.useShift) return null;
    return resolved.schedule;
  };

  const noteInShift = (noteRow, schedule) => {
    const window = getGM3ShiftWindow(workDate, shiftCode, schedule);
    if (!window) return true;

    const startAt =
      getVal(
        noteRow,
        "createdAt",
        "CreatedAt",
        "created_at",
        "startTime",
        "StartTime",
        "start_time"
      ) || "";

    return isDateInShiftWindow(startAt, workDate, shiftCode, schedule);
  };

  const sessionInShift = (row, schedule) => {
    const window = getGM3ShiftWindow(workDate, shiftCode, schedule);
    if (!window) return true;

    const loginAt =
      getVal(
        row,
        "loginTime",
        "LoginTime",
        "login_time",
        "startTime",
        "StartTime",
        "createdAt",
        "CreatedAt"
      ) || "";

    const logoutAt =
      getVal(row, "logoutTime", "LogoutTime", "logout_time", "endTime", "EndTime") ||
      "";

    // Session masuk shift jika login di dalam window, atau overlap window.
    if (isDateInShiftWindow(loginAt, workDate, shiftCode, schedule)) {
      return true;
    }

    if (!loginAt || !logoutAt) {
      return false;
    }

    const loginRaw = String(loginAt).trim();
    const logoutRaw = String(logoutAt).trim();
    const loginDate = new Date(
      loginRaw.includes("T") ? loginRaw : loginRaw.replace(" ", "T")
    );
    const logoutDate = new Date(
      logoutRaw.includes("T") ? logoutRaw : logoutRaw.replace(" ", "T")
    );

    if (Number.isNaN(loginDate.getTime()) || Number.isNaN(logoutDate.getTime())) {
      return false;
    }

    return (
      loginDate.getTime() < window.end.getTime() &&
      logoutDate.getTime() > window.start.getTime()
    );
  };

  const tempMap = new Map();

  extractRows(reportData).forEach((row) => {
    const uuid = String(getVal(row, "uuid", "UUID") || "").trim();

    if (!uuid) return;

    const schedule = resolveScheduleForUuid(uuid);
    const filterByShift = Boolean(
      schedule && getGM3ShiftWindow(workDate, shiftCode, schedule)
    );

    const rowStatus = String(getVal(row, "status", "Status") || "").trim();
    const key = normalizeText(uuid);

    if (!tempMap.has(key)) {
      tempMap.set(key, {
        operatorRows: [],
        rowKeySet: new Set(),
        processName: "",
        styleName: "",
        hasActiveProcess: false,
      });
    }

    const current = tempMap.get(key);
    const isActive = ["ACTIVE", "OPEN"].includes(rowStatus.toUpperCase());

    const operatorNik = String(
      getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
    ).trim();

    const operatorName = String(
      getVal(row, "operatorName", "OperatorName", "operator_name") || ""
    )
      .trim()
      .toUpperCase();

    const processName = String(
      getVal(
        row,
        "processName",
        "ProcessName",
        "process_name",
        "process",
        "Process"
      ) || ""
    ).trim();

    const styleName = String(
      getVal(row, "styleName", "StyleName", "style_name", "style", "Style") ||
        ""
    ).trim();

    const rawNotes = extractRows(
      getVal(row, "notes", "Notes", "lastNotes", "LastNotes") || []
    ).filter((noteRow) => noteInShift(noteRow, schedule));

    // Skip session yang sama sekali tidak relevan dengan shift terpilih.
    if (filterByShift && !rawNotes.length && !sessionInShift(row, schedule)) {
      return;
    }

    // Sama seperti dashboard: utamakan process/style dari session ACTIVE di shift ini.
    if (
      isActive &&
      processName &&
      (!filterByShift || sessionInShift(row, schedule))
    ) {
      current.processName = processName;
      current.styleName = styleName;
      current.hasActiveProcess = true;
    } else if (!current.hasActiveProcess) {
      if (processName && !current.processName) {
        current.processName = processName;
      }
      if (styleName && !current.styleName) {
        current.styleName = styleName;
      }
    }

    const operator =
      operatorNik && operatorName
        ? `${operatorNik} - ${operatorName}`
        : operatorName || operatorNik || "";

    const loginTime = String(
      getVal(row, "loginTime", "LoginTime", "login_time") || ""
    ).trim();

    const logoutTime = String(
      getVal(row, "logoutTime", "LogoutTime", "logout_time") || ""
    ).trim();

    const usageText = formatOperatorUsageText(loginTime, logoutTime, rowStatus);

    const notes = rawNotes.map(normalizeExportNote).filter(Boolean);

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
      !notes.length &&
      (!filterByShift ||
        isDateInShiftWindow(
          activeLossStartTime,
          workDate,
          shiftCode,
          schedule
        ))
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

    if (
      !notes.length &&
      operator &&
      !isAutoLogoutText(operator) &&
      !isAutoLogoutText(operatorName) &&
      (!filterByShift || sessionInShift(row, schedule))
    ) {
      notes.push("");
    }

    // Setelah filter shift, session tanpa note relevan → jangan tampilkan operator.
    if (!notes.length) {
      return;
    }

    notes.forEach((note) => {
      const cleanNote = String(note || "").trim();

      if (isAutoLogoutText(cleanNote)) {
        return;
      }

      const rowKey = `${operator}||${loginTime}||${processName}||${styleName}||${cleanNote}`;

      if (current.rowKeySet.has(rowKey)) {
        return;
      }

      current.rowKeySet.add(rowKey);
      current.operatorRows.push({
        operator,
        operatorNik,
        operatorName,
        note: cleanNote,
        processName,
        styleName,
        loginTime,
        logoutTime,
        usageText,
        hasSessionStats: Boolean(
          getVal(row, "hasSessionStats", "HasSessionStats")
        ),
        runtimeSec: toNumber(
          getVal(row, "runtimeSec", "RuntimeSec", "runtime_sec") || 0
        ),
        procSec: toNumber(
          getVal(row, "procSec", "ProcSec", "proc_sec") || 0
        ),
        lossTimeSec: toNumber(
          getVal(
            row,
            "lossTimeSec",
            "LossTimeSec",
            "loss_time_sec",
            "lossSec",
            "LossSec"
          ) || 0
        ),
        productivityPct: toNumber(
          getVal(
            row,
            "productivityPct",
            "ProductivityPct",
            "productivity_pct"
          ) || 0
        ),
        productivityStatus: String(
          getVal(
            row,
            "productivityStatus",
            "ProductivityStatus",
            "productivity_status"
          ) || ""
        ).trim(),
      });
    });
  });

  const map = new Map();

  tempMap.forEach((value, key) => {
    map.set(key, {
      operatorRows: value.operatorRows,
      processName: value.processName,
      styleName: value.styleName,
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
  getManualSettingByUuid = () => null,
  shiftCode = "",
  shiftConfigMap = new Map()
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
  ).trim();

  const locationFromApi = String(getVal(row, "location", "Location") || "");
  const location = setting?.location || locationFromApi || "-";
  const locationGroup = getLocationGroup(location);
  const lineShift = resolveLineShiftForLocation(location, shiftConfigMap);

  const operatorInfo = operatorMap.get(normalizeText(uuid)) || {};
  const operatorRows = Array.isArray(operatorInfo.operatorRows)
    ? operatorInfo.operatorRows
    : [];

  const processFromOperatorRows = String(
    operatorRows.find((item) => String(item.processName || "").trim())
      ?.processName || ""
  ).trim();

  const styleFromOperatorRows = String(
    operatorRows.find((item) => String(item.styleName || "").trim())
      ?.styleName || ""
  ).trim();

  const processFromOperatorInfo = String(operatorInfo.processName || "").trim();
  const styleFromOperatorInfo = String(operatorInfo.styleName || "").trim();

  // Sama dashboard: display name = process operator, fallback custom/backend name.
  const operatorProcessName =
    processFromOperatorInfo || processFromOperatorRows;
  const operatorStyleName = styleFromOperatorInfo || styleFromOperatorRows;

  const originalMesin = setting?.customName || backendName || uuid;
  const machineName = operatorProcessName || originalMesin;

  const apiShiftCode = String(
    getVal(row, "shiftCode", "ShiftCode", "shift_code") || ""
  ).trim();

  const apiShiftName = String(
    getVal(row, "shiftName", "ShiftName", "shift_name") || ""
  ).trim();

  const resolvedShiftCode = apiShiftCode || shiftCode || "";
  const resolvedShiftName =
    apiShiftName ||
    getExportShiftLabel(locationGroup, resolvedShiftCode, {
      useShift: lineShift.useShift,
      schedule: lineShift.schedule,
      explicitDisabled: !lineShift.useShift,
    });

  return {
    dateText,
    tanggal: formatExcelDate(dateText),
    hari: getDayName(dateText),
    uuid,
    mesin: machineName,
    originalMesin,
    styleName: operatorStyleName,
    location,
    locationGroup,
    shiftCode: resolvedShiftCode,
    shiftName: resolvedShiftName,
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

function resolveOperatorNikName(item) {
  const nik = String(item?.operatorNik || "").trim();
  let name = String(item?.operatorName || "").trim();

  if (nik || name) {
    return {
      operatorNik: nik,
      operatorName: name.toUpperCase(),
    };
  }

  const combined = String(item?.operator || "").trim();

  if (!combined) {
    return { operatorNik: "", operatorName: "" };
  }

  const parts = combined.split(" - ");

  if (parts.length >= 2) {
    return {
      operatorNik: String(parts[0] || "").trim(),
      operatorName: parts.slice(1).join(" - ").trim().toUpperCase(),
    };
  }

  if (/^\d+$/.test(combined)) {
    return { operatorNik: combined, operatorName: "" };
  }

  return { operatorNik: "", operatorName: combined.toUpperCase() };
}

function isUsefulOperatorRow(item) {
  const { operatorNik, operatorName } = resolveOperatorNikName(item);
  const note = String(item?.note || "").trim();

  return Boolean(operatorNik || operatorName || note);
}

export function expandOperatorRows(baseRow, operatorRows, options = {}) {
  const fallbackMesin = String(
    options.fallbackMesin || baseRow.Mesin || ""
  ).trim();
  const fallbackStyle = String(
    options.fallbackStyle || baseRow.Style || ""
  ).trim();

  let rows = (Array.isArray(operatorRows) ? operatorRows : []).filter(
    isUsefulOperatorRow
  );

  // Utamakan baris yang punya note (seperti contoh Excel).
  const rowsWithNotes = rows.filter((item) => String(item?.note || "").trim());
  if (rowsWithNotes.length) {
    rows = rowsWithNotes;
  }

  if (!rows.length) {
    return [
      {
        ...baseRow,
        Mesin: fallbackMesin,
        Style: fallbackStyle,
        "Operator NIK": "",
        "Operator Name": "",
        "Waktu Pemakaian": "",
        "Operator Note": "",
      },
    ];
  }

  // Urutkan: login → NIK → Name → Note.
  const sortedRows = rows.slice().sort((a, b) => {
    const timeA = parseExportDateTime(a?.loginTime)?.getTime() || 0;
    const timeB = parseExportDateTime(b?.loginTime)?.getTime() || 0;
    if (timeA !== timeB) return timeA - timeB;

    const left = resolveOperatorNikName(a);
    const right = resolveOperatorNikName(b);

    const nikCmp = left.operatorNik.localeCompare(right.operatorNik);
    if (nikCmp !== 0) return nikCmp;

    const nameCmp = left.operatorName.localeCompare(right.operatorName);
    if (nameCmp !== 0) return nameCmp;

    return String(a?.note || "").localeCompare(String(b?.note || ""));
  });

  // Pola:
  // - Baris pertama tiap session operator: NIK + Name + Waktu Pemakaian + Note
  // - Baris note berikutnya: NIK/Name/Waktu kosong, Note tetap terisi
  let lastSessionKey = "";

  return sortedRows.map((item) => {
    const { operatorNik, operatorName } = resolveOperatorNikName(item);
    const processName = String(item?.processName || "").trim();
    const styleName = String(item?.styleName || "").trim();
    const loginTime = String(item?.loginTime || "").trim();
    const usageText = String(item?.usageText || "").trim();
    const sessionKey = `${operatorNik}||${operatorName}||${loginTime}||${usageText}`;
    const isFirstOfSession = sessionKey !== lastSessionKey;

    if (isFirstOfSession) {
      lastSessionKey = sessionKey;
    }

    const hasSessionStats = Boolean(item?.hasSessionStats);
    const sessionPowerOn = hasSessionStats
      ? formatSeconds(item.runtimeSec)
      : baseRow["Power On Duration"];
    const sessionRunning = hasSessionStats
      ? formatSeconds(item.procSec)
      : baseRow["Running Time"];
    const sessionLoss = hasSessionStats
      ? formatSeconds(item.lossTimeSec)
      : baseRow["Loss Time"];
    const sessionProductivity = hasSessionStats
      ? Number(item.productivityPct || 0)
      : baseRow.Produktivitas;
    const sessionStatus = hasSessionStats
      ? String(item.productivityStatus || statusFromProductivity(item.productivityPct)).trim()
      : baseRow.Status;

    return {
      ...baseRow,
      Mesin: isFirstOfSession ? processName || fallbackMesin : "",
      Style: isFirstOfSession ? styleName || fallbackStyle : "",
      "Power On Duration": isFirstOfSession ? sessionPowerOn : "",
      "Running Time": isFirstOfSession ? sessionRunning : "",
      "Loss Time": isFirstOfSession ? sessionLoss : "",
      Produktivitas: isFirstOfSession ? sessionProductivity : "",
      Status: isFirstOfSession ? sessionStatus : "",
      "Operator NIK": isFirstOfSession ? operatorNik : "",
      "Operator Name": isFirstOfSession ? operatorName : "",
      "Waktu Pemakaian": isFirstOfSession ? usageText : "",
      "Operator Note": String(item?.note || "").trim(),
    };
  });
}

export function buildRangeDetailRows(items, shiftCode = "") {
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
      const area = getLocationGroup(item.location);
      const shift =
        item.shiftName ||
        getExportShiftLabel(area, item.shiftCode || shiftCode, {
          useShift: area === "GM3" || Boolean(item.shiftCode),
          schedule: null,
        });

      const baseRow = {
        Tanggal: item.tanggal,
        Hari: item.hari,
        Shift: shift,
        Area: area,
        Location: item.location,
        Mesin: item.mesin,
        Style: item.styleName || "",
        "Power On Duration": formatSeconds(item.runtime),
        "Running Time": formatSeconds(item.procTime),
        "Loss Time": formatSeconds(item.lossTime),
        Produktivitas: item.productivity,
        Status: item.status,
      };

      return expandOperatorRows(baseRow, item.operatorRows, {
        fallbackMesin: item.originalMesin || item.mesin,
        fallbackStyle: item.styleName || "",
      });
    });
}

export function buildRangeSummaryBaseRows(items, rangeStart, rangeEnd, shiftCode = "") {
  const map = new Map();

  items.forEach((item) => {
    const area = getLocationGroup(item.location);
    const shift =
      item.shiftName ||
      getExportShiftLabel(area, item.shiftCode || shiftCode, {
        useShift: Boolean(item.shiftName) || area === "GM3",
        schedule: null,
      });
    const uuid = String(item.uuid || "").trim() || `${item.location}||${item.originalMesin || item.mesin}`;
    const key = `${area}||${shift}||${uuid}`;

    if (!map.has(key)) {
      map.set(key, {
        periode: `${formatExcelDate(rangeStart)} - ${formatExcelDate(rangeEnd)}`,
        shift,
        area,
        location: item.location,
        mesin: item.mesin,
        styleName: item.styleName || "",
        originalMesin: item.originalMesin || item.mesin,
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

    // Update nama tampilan ke process/style terbaru (selaras dashboard).
    if (item.mesin && item.mesin !== item.originalMesin) {
      current.mesin = item.mesin;
    }
    if (item.styleName) {
      current.styleName = item.styleName;
    }
    if (item.location) {
      current.location = item.location;
    }

    const operatorRows = Array.isArray(item.operatorRows)
      ? item.operatorRows
      : [];

    operatorRows.forEach((operatorRow) => {
      const { operatorNik, operatorName } = resolveOperatorNikName(operatorRow);
      const operator = String(operatorRow.operator || "").trim();
      const note = String(operatorRow.note || "").trim();
      const processName = String(operatorRow.processName || "").trim();
      const styleName = String(operatorRow.styleName || "").trim();
      const loginTime = String(operatorRow.loginTime || "").trim();
      const logoutTime = String(operatorRow.logoutTime || "").trim();
      const usageText = String(operatorRow.usageText || "").trim();

      if (
        isAutoLogoutText(operator) ||
        isAutoLogoutText(operatorName) ||
        isAutoLogoutText(note)
      ) {
        return;
      }

      const rowKey = `${operatorNik}||${operatorName}||${loginTime}||${processName}||${styleName}||${note}`;

      if (current.operatorRowKeySet.has(rowKey)) {
        return;
      }

      current.operatorRowKeySet.add(rowKey);
      current.operatorRows.push({
        operator,
        operatorNik,
        operatorName,
        note,
        processName,
        styleName,
        loginTime,
        logoutTime,
        usageText,
        hasSessionStats: Boolean(operatorRow.hasSessionStats),
        runtimeSec: toNumber(operatorRow.runtimeSec || 0),
        procSec: toNumber(operatorRow.procSec || 0),
        lossTimeSec: toNumber(operatorRow.lossTimeSec || 0),
        productivityPct: toNumber(operatorRow.productivityPct || 0),
        productivityStatus: String(operatorRow.productivityStatus || "").trim(),
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
        shift: item.shift,
        area: item.area,
        location: item.location,
        mesin: item.mesin,
        styleName: item.styleName,
        originalMesin: item.originalMesin,
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
      const shiftCompare = String(a.shift).localeCompare(String(b.shift));
      const locCompare = String(a.location).localeCompare(String(b.location));
      const mesinCompare = String(a.mesin).localeCompare(String(b.mesin));

      if (areaCompare !== 0) return areaCompare;
      if (shiftCompare !== 0) return shiftCompare;
      if (locCompare !== 0) return locCompare;
      if (mesinCompare !== 0) return mesinCompare;

      return Number(a.productivity || 0) - Number(b.productivity || 0);
    });
}

export function buildRangeSummaryRows(summaryBaseRows) {
  return summaryBaseRows.flatMap((item) => {
    const baseRow = {
      Periode: item.periode,
      Shift: item.shift,
      Area: item.area,
      Location: item.location,
      Mesin: item.mesin,
      Style: item.styleName || "",
      "Power On Duration": formatSeconds(item.totalPowerOn),
      "Running Time": formatSeconds(item.totalRunning),
      "Loss Time": formatSeconds(item.totalLoss),
      Produktivitas: item.productivity,
      Status: item.status,
    };

    return expandOperatorRows(baseRow, item.operatorRows, {
      fallbackMesin: item.originalMesin || item.mesin,
      fallbackStyle: item.styleName || "",
    });
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
        Shift: "",
        Area: selectedLocation === "ALL" ? "All GM" : selectedLocation,
        Location: "",
        Mesin: "Tidak ada mesin BAD",
        Style: "",
        "Power On Duration": "",
        "Running Time": "",
        "Loss Time": "",
        Produktivitas: "",
        Status: "",
        Rekomendasi: "Semua mesin tidak berstatus BAD pada periode ini.",
        "Operator NIK": "",
        "Operator Name": "",
        "Waktu Pemakaian": "",
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
      Shift: row.shift,
      Area: row.area,
      Location: row.location,
      Mesin: row.mesin,
      Style: row.styleName || "",
      "Power On Duration": formatSeconds(row.totalPowerOn),
      "Running Time": formatSeconds(row.totalRunning),
      "Loss Time": formatSeconds(row.totalLoss),
      Produktivitas: row.productivity,
      Status: row.status,
      Rekomendasi: rekomendasi,
    };

    return expandOperatorRows(baseRow, row.operatorRows, {
      fallbackMesin: row.originalMesin || row.mesin,
      fallbackStyle: row.styleName || "",
    });
  });
}

export function createRangeWorkbook(summaryRows, detailRows, badPriorityRows) {
  const workbook = XLSX.utils.book_new();

  const summaryCols = 15;
  const detailCols = 16;
  const badCols = 16;

  const summarySheet = XLSX.utils.json_to_sheet(summaryRows);
  setWorksheetWidth(summarySheet, [
    24, 28, 10, 18, 42, 24, 18, 16, 14, 14, 12, 14, 28, 36, 50,
  ]);
  setAutoFilter(summarySheet, summaryRows.length, summaryCols);

  const detailSheet = XLSX.utils.json_to_sheet(detailRows);
  setWorksheetWidth(detailSheet, [
    12, 12, 28, 10, 18, 42, 24, 18, 16, 14, 14, 12, 14, 28, 36, 50,
  ]);
  setAutoFilter(detailSheet, detailRows.length, detailCols);

  const badSheet = XLSX.utils.json_to_sheet(badPriorityRows);
  setWorksheetWidth(badSheet, [
    8, 28, 10, 18, 42, 24, 18, 16, 14, 14, 12, 40, 14, 28, 36, 50,
  ]);
  setAutoFilter(badSheet, badPriorityRows.length, badCols);

  XLSX.utils.book_append_sheet(workbook, summarySheet, "Range Summary");
  XLSX.utils.book_append_sheet(workbook, detailSheet, "Daily Detail");
  XLSX.utils.book_append_sheet(workbook, badSheet, "BAD Priority");

  return workbook;
}

export function writeWorkbook(workbook, fileName) {
  XLSX.writeFile(workbook, fileName);
}