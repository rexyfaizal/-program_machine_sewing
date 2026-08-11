export const WORK_SECONDS = 28800;

export function todayLocal() {
  const d = new Date();
  const offset = d.getTimezoneOffset();
  const local = new Date(d.getTime() - offset * 60000);
  return local.toISOString().slice(0, 10);
}

function toLocalIsoDate(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function startOfLocalDay(date = new Date()) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

/** Senin–Sabtu minggu berjalan (Minggu tidak termasuk). */
export function getThisWeekRange(now = new Date()) {
  const today = startOfLocalDay(now);
  const day = today.getDay();
  const monday = new Date(today);
  monday.setDate(today.getDate() + (day === 0 ? -6 : 1 - day));
  const saturday = new Date(monday);
  saturday.setDate(monday.getDate() + 5);
  return { start: toLocalIsoDate(monday), end: toLocalIsoDate(saturday) };
}

export function getThisMonthRange(now = new Date()) {
  const start = new Date(now.getFullYear(), now.getMonth(), 1);
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return { start: toLocalIsoDate(start), end: toLocalIsoDate(end) };
}

export function getTodayRange(now = new Date()) {
  const day = toLocalIsoDate(startOfLocalDay(now));
  return { start: day, end: day };
}

export const MONTH_LABELS_ID = [
  "Januari",
  "Februari",
  "Maret",
  "April",
  "Mei",
  "Juni",
  "Juli",
  "Agustus",
  "September",
  "Oktober",
  "November",
  "Desember",
];

export function getMonthRangeByIndex(year, monthIndex) {
  const start = new Date(year, monthIndex, 1);
  const end = new Date(year, monthIndex + 1, 0);
  return { start: toLocalIsoDate(start), end: toLocalIsoDate(end) };
}

function formatDayMonthId(start, end) {
  const startMonth = MONTH_LABELS_ID[start.getMonth()];
  const endMonth = MONTH_LABELS_ID[end.getMonth()];
  if (start.getMonth() === end.getMonth()) {
    return `${start.getDate()}-${end.getDate()} ${startMonth}`;
  }
  return `${start.getDate()} ${startMonth}-${end.getDate()} ${endMonth}`;
}

/** Minggu 1..n = Senin–Sabtu dalam bulan itu (hari Minggu dilewati). */
export function getMonthWeekOptions(year, monthIndex) {
  const last = new Date(year, monthIndex + 1, 0);
  const cursor = new Date(year, monthIndex, 1);

  while (cursor.getDay() !== 1 && cursor <= last) {
    cursor.setDate(cursor.getDate() + 1);
  }

  const options = [];
  let weekNo = 1;

  while (cursor <= last && cursor.getMonth() === monthIndex) {
    const saturday = new Date(cursor);
    saturday.setDate(cursor.getDate() + 5);
    const end = saturday > last ? new Date(last) : saturday;
    if (end.getDay() === 0) {
      end.setDate(end.getDate() - 1);
    }

    if (end >= cursor) {
      options.push({
        weekNo,
        start: toLocalIsoDate(cursor),
        end: toLocalIsoDate(end),
        label: `Minggu ${weekNo}`,
        rangeLabel: formatDayMonthId(cursor, end),
      });
      weekNo += 1;
    }

    cursor.setDate(cursor.getDate() + 7);
  }

  return options;
}

export function formatDateRangeLabel(startText, endText) {
  const start = new Date(`${startText}T00:00:00`);
  const end = new Date(`${String(endText || startText)}T00:00:00`);
  if (Number.isNaN(start.getTime())) return "-";
  if (Number.isNaN(end.getTime()) || startText === endText) {
    return `${start.getDate()} ${MONTH_LABELS_ID[start.getMonth()]} ${start.getFullYear()}`;
  }
  const year = end.getFullYear();
  return `${formatDayMonthId(start, end)} ${year}`;
}

export function getStatus(productivity) {
  const value = Number(productivity || 0);

  if (value >= 90) return "GOOD";
  if (value >= 80) return "NORMAL";
  return "BAD";
}

export function statusClass(status) {
  if (status === "GOOD") return "good";
  if (status === "NORMAL") return "normal";
  return "bad";
}

export function formatHour(seconds) {
  const hour = Number(seconds || 0) / 3600;
  return `${hour.toFixed(2)}h`;
}

export function formatSecond(seconds) {
  return `${Number(seconds || 0).toFixed(0)}s`;
}

export function formatDurationHHMMSS(sec) {
  sec = Math.max(0, Number(sec || 0));
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = Math.floor(sec % 60);

  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export function formatTimeOnly(value) {
  if (!value) return "-";
  const text = String(value);
  return text.length >= 16 ? text.slice(11, 16) : text;
}

export function csvCell(value) {
  const text = String(value ?? "");
  return `"${text.replace(/"/g, '""')}"`;
}

export function moveDate(date, delta) {
  const d = new Date(date + "T00:00:00");
  d.setDate(d.getDate() + delta);
  return d.toISOString().slice(0, 10);
}
