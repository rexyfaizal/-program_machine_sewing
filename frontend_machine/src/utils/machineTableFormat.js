export function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

export function downTimeSec(machine) {
  const powerOnDuration = toNumber(machine?.runtime);
  const runningTime = toNumber(machine?.procTime);

  return Math.max(0, powerOnDuration - runningTime);
}

export function statusClass(status) {
  const s = String(status || "BAD").toUpperCase();

  if (s === "GOOD") return "good";
  if (s === "NORMAL") return "normal";

  return "bad";
}

export function formatHour(seconds) {
  const totalSeconds = Math.max(0, Math.floor(toNumber(seconds)));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  return `${hours}h ${minutes}m`;
}

export function getDisplayMachineName(machine) {
  return machine?.displayMachineName || machine?.machineName || "-";
}

export function isUsingProcessName(machine) {
  const display = String(machine?.displayMachineName || "").trim();
  const original = String(machine?.machineName || "").trim();

  return Boolean(display && original && display !== original);
}

export function getOperatorNote(machine) {
  return machine?.operatorNote || machine?.spv || "";
}

export function getOperatorName(machine) {
  return machine?.pic || "";
}

export function getOperatorSubText(machine) {
  return machine?.operatorSubText || "";
}

export function getOperatorDisplayRows(machine) {
  if (Array.isArray(machine?.operatorDisplayRows) && machine.operatorDisplayRows.length) {
    return machine.operatorDisplayRows;
  }

  const name = getOperatorName(machine);
  if (!name) return [];

  return [
    {
      label: name,
      subText: getOperatorSubText(machine),
      note: getOperatorNote(machine),
    },
  ];
}