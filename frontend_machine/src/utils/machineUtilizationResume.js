import * as XLSX from "xlsx";
import {
  applyProcRuntimeTolerance,
  extractRows,
  getLocationGroup,
  getVal,
  normalizeText,
  toNumber,
} from "./dashboardExportExcel";

function text(row, ...keys) {
  return String(getVal(row, ...keys) ?? "").trim();
}

function identity(row) {
  return JSON.stringify([
    normalizeText(row.uuid),
    normalizeText(row.location),
    normalizeText(row.operatorNik || row.operatorName),
    normalizeText(row.processName),
  ]);
}

function rowInfo(row, fallback = {}) {
  const location = text(row, "location", "Location") || fallback.location || "";
  return {
    uuid: text(row, "uuid", "UUID") || fallback.uuid || "",
    area: getLocationGroup(location),
    location,
    operatorName: text(row, "operatorName", "OperatorName"),
    operatorNik: text(row, "operatorNik", "OperatorNik"),
    processName: text(row, "processName", "ProcessName"),
    machineName: text(row, "machineName", "nickName", "MachineName") || fallback.machineName || "",
  };
}

function durations(row) {
  return applyProcRuntimeTolerance(
    toNumber(getVal(row, "runtimeSec", "RuntimeSec")),
    toNumber(getVal(row, "procSec", "ProcSec", "productive_seconds"))
  );
}

export function average(values) {
  const numbers = values.filter((value) => typeof value === "number" && Number.isFinite(value));
  return numbers.length ? numbers.reduce((sum, value) => sum + value, 0) / numbers.length : null;
}

// One row per machine, location, operator, and process.
// Combine sessions within each day before averaging daily percentages by week.
export function buildMachineUtilizationResume(days, weeks) {
  const rows = new Map();
  const inventory = new Map();
  const representedMachines = new Set();

  function addDay(info, date, runtime, procTime) {
    if (!info.uuid) return;
    const key = identity(info);
    if (!rows.has(key)) rows.set(key, { ...info, key, daily: new Map() });
    const current = rows.get(key);
    if (info.operatorName) current.operatorName = info.operatorName;
    const daily = current.daily.get(date) || { runtime: 0, procTime: 0 };
    daily.runtime += runtime;
    daily.procTime += procTime;
    current.daily.set(date, daily);
    representedMachines.add(normalizeText(info.uuid));
  }

  [...days].sort((a, b) => a.date.localeCompare(b.date)).forEach((day) => {
    const machines = new Map();
    extractRows(day.productivity).forEach((machine) => {
      const info = rowInfo(machine);
      const uuid = normalizeText(info.uuid);
      if (!uuid) return;
      machines.set(uuid, machine);
      inventory.set(uuid, info);
    });

    const assignedMachines = new Set();
    const sessionIds = new Set();
    extractRows(day.operators).forEach((session) => {
      const uuid = normalizeText(text(session, "uuid", "UUID"));
      const info = rowInfo(session, inventory.get(uuid));
      if (!uuid || (!info.operatorNik && !info.operatorName)) return;
      const sessionId = text(session, "id", "ID") || JSON.stringify([
        identity(info), text(session, "loginTime"), text(session, "logoutTime"),
      ]);
      if (sessionIds.has(sessionId)) return;
      sessionIds.add(sessionId);
      assignedMachines.add(uuid);
      const metrics = durations(session);
      addDay(info, day.date, metrics.runtime, metrics.procTime);
    });

    machines.forEach((machine, uuid) => {
      if (assignedMachines.has(uuid)) return;
      const metrics = durations(machine);
      if (metrics.runtime > 0 || metrics.procTime > 0) {
        addDay(inventory.get(uuid), day.date, metrics.runtime, metrics.procTime);
      }
    });
  });

  // Keep machines with no measured activity visible, with empty weekly cells.
  inventory.forEach((info, uuid) => {
    if (!representedMachines.has(uuid)) rows.set(identity(info), { ...info, key: identity(info), daily: new Map() });
  });

  return [...rows.values()].map((row) => ({
    ...row,
    weeks: Object.fromEntries(weeks.map((week) => {
      const measurements = [...row.daily.entries()]
        .filter(([date, metric]) => date >= week.start && date <= week.end && metric.runtime > 0)
        .map(([, metric]) => Math.min(metric.procTime / metric.runtime * 100, 100));
      return [week.weekNo, { value: average(measurements), days: measurements.length }];
    })),
  })).sort((a, b) => {
    for (const key of ["area", "location", "operatorName", "uuid", "processName"]) {
      const comparison = a[key].localeCompare(b[key], "id", { numeric: true, sensitivity: "base" });
      if (comparison) return comparison;
    }
    return 0;
  });
}

export function filterMachineUtilizationRows(rows, building, keyword = "") {
  const search = normalizeText(keyword);
  return rows.filter((row) =>
    (building === "ALL" || row.area === building) &&
    (!search || [row.area, row.location, row.uuid, row.operatorName,
      row.operatorNik, row.processName, row.machineName].some((value) => normalizeText(value).includes(search)))
  );
}

export function weeklyAverages(rows, weeks) {
  return Object.fromEntries(weeks.map((week) => [week.weekNo, average(rows.map((row) => row.weeks[week.weekNo]?.value))]));
}

export function createMachineUtilizationWorkbook({ rows, weeks, monthLabel }) {
  const headings = ["Area", "Location", "UUID", "NAMA", "NIK", "PROSES",
    ...weeks.map((week) => `AVG W${week.weekNo}\n${monthLabel}\n${week.rangeLabel}`)];
  const means = weeklyAverages(rows, weeks);
  const values = rows.map((row) => [row.area, row.location, row.uuid,
    row.operatorName, row.operatorNik, row.processName,
    ...weeks.map((week) => {
      const value = row.weeks[week.weekNo]?.value;
      return typeof value === "number" ? value / 100 : null;
    })]);
  const footer = ["", "", "", "", "", "Average",
    ...weeks.map((week) => means[week.weekNo] === null ? null : means[week.weekNo] / 100)];
  const sheet = XLSX.utils.aoa_to_sheet([headings, ...values, footer]);
  const footerIndex = rows.length + 1;
  weeks.forEach((week, index) => {
    const column = index + 6;
    for (let rowIndex = 1; rowIndex <= footerIndex; rowIndex++) {
      const address = XLSX.utils.encode_cell({ r: rowIndex, c: column });
      if (sheet[address]) sheet[address].z = "0.00%";
    }
    if (rows.length && means[week.weekNo] !== null) {
      const address = XLSX.utils.encode_cell({ r: footerIndex, c: column });
      const letter = XLSX.utils.encode_col(column);
      sheet[address].f = `AVERAGE(${letter}2:${letter}${rows.length + 1})`;
    }
  });
  sheet["!cols"] = [10, 30, 26, 32, 16, 48, ...weeks.map(() => 23)].map((wch) => ({ wch }));
  sheet["!rows"] = [{ hpt: 45 }];
  sheet["!autofilter"] = { ref: XLSX.utils.encode_range({ r: 0, c: 0 }, { r: rows.length, c: headings.length - 1 }) };
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, sheet, "Resume Utilisasi Mesin");
  workbook.Props = { Title: `Resume Utilisasi Mesin - ${monthLabel}` };
  return workbook;
}

export function exportMachineUtilizationExcel(options) {
  const workbook = createMachineUtilizationWorkbook(options);
  const weekSuffix = options.weeks.map((week) => `W${week.weekNo}`).join("-");
  XLSX.writeFile(workbook, `resume-utilisasi-mesin-${options.building.toLowerCase()}-${options.period}-${weekSuffix}.xlsx`);
}
