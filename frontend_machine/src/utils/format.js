export const WORK_SECONDS = 28800;

export function todayLocal() {
  const d = new Date();
  const offset = d.getTimezoneOffset();
  const local = new Date(d.getTime() - offset * 60000);
  return local.toISOString().slice(0, 10);
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
