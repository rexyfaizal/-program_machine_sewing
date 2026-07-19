function getVal(obj, ...keys) {
  for (const key of keys) {
    if (obj && obj[key] !== undefined && obj[key] !== null) {
      return obj[key];
    }
  }

  return undefined;
}

export function getMacStateValue(machine) {
  const value = String(
    getVal(
      machine,
      "macState",
      "MacState",
      "mac_state",
      "Macstate",
      "macstate",
      "MACSTATE"
    ) ?? "0"
  ).trim();

  if (value === "2") return "2";
  if (value === "1") return "1";
  return "0";
}

export function getMachineStatusClass(machine) {
  const macState = getMacStateValue(machine);

  if (macState === "2") return "working";
  if (macState === "1") return "online";
  return "offline";
}

export function getMachineStatusText(machine) {
  const statusClass = getMachineStatusClass(machine);

  if (statusClass === "working") return "Working";
  if (statusClass === "online") return "Online";
  return "Offline";
}

export function getMachineStatusLabel(machine) {
  return getMachineStatusText(machine);
}

export function isMachineWorking(machine) {
  return getMachineStatusClass(machine) === "working";
}

export function isMachineOnline(machine) {
  const statusClass = getMachineStatusClass(machine);

  return statusClass === "working" || statusClass === "online";
}

export function isMachineOffline(machine) {
  return getMachineStatusClass(machine) === "offline";
}