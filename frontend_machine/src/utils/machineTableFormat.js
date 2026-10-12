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

export function formatMinute(seconds) {
  const totalSeconds = Math.max(0, Math.floor(toNumber(seconds)));
  const minutes = Math.floor(totalSeconds / 60);

  return `${minutes}m`;
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

export function getOperatorNoteItems(machine) {
  if (Array.isArray(machine?.operatorNoteItems) && machine.operatorNoteItems.length) {
    return machine.operatorNoteItems;
  }

  const rows = getOperatorDisplayRows(machine);
  const fromRows = rows.flatMap((row) =>
    Array.isArray(row?.noteItems) ? row.noteItems : []
  );

  if (fromRows.length) return fromRows;

  // Fallback: note tersimpan di baris operator (bukan field machine.operatorNotes).
  const fromRowNotes = rows
    .map((row) => String(row?.note || "").trim())
    .filter((text) => text && text !== "-");
  if (fromRowNotes.length) {
    return fromRowNotes.map((text) => ({
      text,
      timeRange: "",
      reasonName: "",
      body: text,
      isActive: false,
    }));
  }

  const raw = String(getOperatorNote(machine) || "").trim();
  if (!raw || raw === "-") return [];

  return raw
    .split(/\s*\|\s*/)
    .map((text) => text.trim())
    .filter(Boolean)
    .map((text) => ({ text, timeRange: "", reasonName: "", body: text, isActive: false }));
}

/** Gabungan keterangan loss/note seperti yang tampil di tabel dashboard. */
export function collectMachineOperatorNotes(machine) {
  const items = getOperatorNoteItems(machine);
  if (items.length) {
    return items
      .map((note) => {
        const line =
          [
            [note.reasonName, note.body].filter(Boolean).join(" "),
            note.timeRange,
          ]
            .filter(Boolean)
            .join(" | ") || String(note.text || "").trim();
        return String(line || "").trim();
      })
      .filter(Boolean)
      .join(" | ");
  }

  return String(machine?.operatorNotes || machine?.operatorNote || "")
    .trim()
    .replace(/^-$/, "");
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

/** Operator belum login, tapi mesin ON / ada power-on / ada process. */
export function isNoLoginRunning(machine) {
  const rows = getOperatorDisplayRows(machine);
  const hasLogin = rows.some((row) => {
    const label = String(row?.label || "").trim();
    if (!label) return false;
    return !/^not\s*logged\s*in$/i.test(label);
  });
  if (hasLogin) return false;

  const mac = String(
    machine?.macState ?? machine?.MacState ?? machine?.mac_state ?? ""
  ).trim();
  if (mac === "1" || mac === "2") return true;

  const runtime = toNumber(machine?.runtime ?? machine?.runtimeSec);
  const proc = toNumber(machine?.procTime ?? machine?.procSec);
  if (proc > 0) return true;
  // Power on ≥ 1 menit sebagai sinyal mesin sempat/masih nyala.
  if (runtime >= 60) return true;
  return false;
}

export function noLoginRunningLabel(machine) {
  const proc = toNumber(machine?.procTime ?? machine?.procSec);
  const mac = String(
    machine?.macState ?? machine?.MacState ?? machine?.mac_state ?? ""
  ).trim();
  if (proc > 0 || mac === "2") return "Tanpa login · mesin jalan";
  return "Tanpa login · mesin ON";
}