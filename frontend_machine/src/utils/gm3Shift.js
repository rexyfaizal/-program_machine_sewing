/** Jadwal shift default (legacy GM3) + helper window dinamis. */

export const DEFAULT_SHIFT_SCHEDULE = [
  {
    code: "SHIFT_1",
    start: "06:00",
    end: "13:30",
    breakStart: "09:30",
    breakEnd: "10:00",
  },
  {
    code: "SHIFT_2",
    start: "13:30",
    end: "21:00",
    breakStart: "17:00",
    breakEnd: "17:30",
  },
  {
    code: "SHIFT_3",
    start: "21:00",
    end: "04:30",
    breakStart: "01:00",
    breakEnd: "01:30",
  },
];

export const DEFAULT_SATURDAY_SHIFT_SCHEDULE = [
  {
    code: "SHIFT_1",
    start: "06:00",
    end: "11:30",
    breakStart: "10:00",
    breakEnd: "10:30",
  },
  {
    code: "SHIFT_2",
    start: "11:30",
    end: "17:00",
    breakStart: "15:30",
    breakEnd: "16:00",
  },
  {
    code: "SHIFT_3",
    start: "17:00",
    end: "22:30",
    breakStart: "21:00",
    breakEnd: "21:30",
  },
];

export const GM3_SHIFT_OPTIONS = [
  { value: "CURRENT", label: "Current Shift" },
  { value: "SHIFT_1", label: "Shift 1 (06:00-13:30)" },
  { value: "SHIFT_2", label: "Shift 2 (13:30-21:00)" },
  { value: "SHIFT_3", label: "Shift 3 (21:00-04:30)" },
  { value: "ALL", label: "All Shifts" },
];

export function resolveGM3WorkDate(now = new Date()) {
  const d = new Date(now);
  const minutes = d.getHours() * 60 + d.getMinutes();

  // Sen–Sabtu sebelum 04:30 masih work_date kemarin (SHIFT_3 weekday).
  // Minggu dini hari tidak wrap — Sabtu sudah selesai 22:30.
  if (minutes < 270 && d.getDay() !== 0) {
    d.setDate(d.getDate() - 1);
  }

  d.setHours(0, 0, 0, 0);
  return d;
}

export function formatLocalDate(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function parseClockToMinutes(clock) {
  const text = String(clock || "").trim();
  if (!text) return null;

  const parts = text.split(":");
  if (parts.length < 2) return null;

  const hour = Number(parts[0]);
  const minute = Number(parts[1]);

  if (
    !Number.isFinite(hour) ||
    !Number.isFinite(minute) ||
    hour < 0 ||
    hour > 23 ||
    minute < 0 ||
    minute > 59
  ) {
    return null;
  }

  return hour * 60 + minute;
}

export function getGM3CurrentShift(now = new Date()) {
  const workDate = resolveGM3WorkDate(now);
  const minutesFromWorkDate = Math.floor(
    (now.getTime() - workDate.getTime()) / 60000
  );

  if (workDate.getDay() === 6) {
    if (minutesFromWorkDate >= 360 && minutesFromWorkDate < 690) {
      return { shiftCode: "SHIFT_1", shiftName: "Shift 1", workDate };
    }
    if (minutesFromWorkDate >= 690 && minutesFromWorkDate < 1020) {
      return { shiftCode: "SHIFT_2", shiftName: "Shift 2", workDate };
    }
    if (minutesFromWorkDate >= 1020 && minutesFromWorkDate < 1350) {
      return { shiftCode: "SHIFT_3", shiftName: "Shift 3", workDate };
    }
    return { shiftCode: "ALL", shiftName: "All Shifts", workDate };
  }

  if (minutesFromWorkDate >= 360 && minutesFromWorkDate < 810) {
    return { shiftCode: "SHIFT_1", shiftName: "Shift 1", workDate };
  }

  if (minutesFromWorkDate >= 810 && minutesFromWorkDate < 1260) {
    return { shiftCode: "SHIFT_2", shiftName: "Shift 2", workDate };
  }

  if (minutesFromWorkDate >= 1260 && minutesFromWorkDate < 1710) {
    return { shiftCode: "SHIFT_3", shiftName: "Shift 3", workDate };
  }

  return { shiftCode: "ALL", shiftName: "All Shifts", workDate };
}

export function getShiftLabel(shiftCode, schedule = null) {
  const code = String(shiftCode || "ALL").trim().toUpperCase() || "ALL";

  if (code === "CURRENT") {
    const current = getGM3CurrentShift();
    return getShiftLabel(current.shiftCode, schedule);
  }

  if (Array.isArray(schedule)) {
    const item = schedule.find(
      (row) => String(row.code || "").toUpperCase() === code
    );
    if (item?.start && item?.end) {
      const no = code.replace("SHIFT_", "");
      return `Shift ${no} (${item.start}-${item.end})`;
    }
  }

  const found = GM3_SHIFT_OPTIONS.find((item) => item.value === code);
  return found?.label || "All Shifts";
}

/** Label shift untuk export Excel per area mesin. */
export function getExportShiftLabel(locationGroup, shiftCode, options = {}) {
  const group = String(locationGroup || "").trim().toUpperCase();
  const useShift = Boolean(options.useShift);

  if (!useShift && group !== "GM3") {
    return "Full Day";
  }

  if (!useShift && group === "GM3" && options.explicitDisabled) {
    return "Full Day";
  }

  return getShiftLabel(shiftCode, options.schedule || null);
}

/**
 * Window jam kerja efektif per shift.
 * schedule opsional: [{code,start,end}] — jika ada dipakai; else default GM3.
 */
export function getGM3ShiftWindow(workDateText, shiftCode, schedule = null) {
  const code = String(shiftCode || "ALL").trim().toUpperCase();
  const dateText = String(workDateText || "").trim();

  if (!dateText || !/^\d{4}-\d{2}-\d{2}$/.test(dateText)) {
    return null;
  }

  if (!code || code === "ALL") {
    return null;
  }

  let resolved = code;

  if (code === "CURRENT") {
    const current = getGM3CurrentShift();
    const todayWork = formatLocalDate(current.workDate);

    if (dateText !== todayWork || current.shiftCode === "ALL") {
      return null;
    }

    resolved = current.shiftCode;
  }

  const base = new Date(`${dateText}T00:00:00`);
  if (Number.isNaN(base.getTime())) {
    return null;
  }

  const addMinutes = (mins) => {
    const d = new Date(base.getTime());
    d.setMinutes(d.getMinutes() + mins);
    return d;
  };

  const scheduleList = Array.isArray(schedule) && schedule.length
    ? schedule
    : DEFAULT_SHIFT_SCHEDULE;

  const item = scheduleList.find(
    (row) => String(row.code || "").toUpperCase() === resolved
  );

  if (item) {
    const startMin = parseClockToMinutes(item.start);
    let endMin = parseClockToMinutes(item.end);

    if (startMin == null || endMin == null) {
      return null;
    }

    if (endMin <= startMin) {
      endMin += 24 * 60;
    }

    return {
      shiftCode: resolved,
      start: addMinutes(startMin),
      end: addMinutes(endMin),
    };
  }

  switch (resolved) {
    case "SHIFT_1":
      return { shiftCode: "SHIFT_1", start: addMinutes(360), end: addMinutes(810) };
    case "SHIFT_2":
      return { shiftCode: "SHIFT_2", start: addMinutes(810), end: addMinutes(1260) };
    case "SHIFT_3":
      return { shiftCode: "SHIFT_3", start: addMinutes(1260), end: addMinutes(1710) };
    default:
      return null;
  }
}

/** True jika waktu jatuh di dalam window shift [start, end). */
export function isDateInShiftWindow(
  dateValue,
  workDateText,
  shiftCode,
  schedule = null
) {
  const window = getGM3ShiftWindow(workDateText, shiftCode, schedule);

  if (!window) return true;

  if (!dateValue) return false;

  const raw = String(dateValue || "").trim();
  if (!raw) return false;

  const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
  const d = new Date(normalized);

  if (Number.isNaN(d.getTime())) return false;

  return d.getTime() >= window.start.getTime() && d.getTime() < window.end.getTime();
}

export function lineShiftConfigKey(factory, lineName) {
  return `${String(factory || "").trim().toUpperCase()}||${String(lineName || "")
    .trim()
    .toUpperCase()}`;
}

export function parseLocationParts(location) {
  const text = String(location || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");

  if (!text) {
    return { factory: "", lineName: "" };
  }

  const match = text.match(/\bGM\s*([0-9]+)\b/);
  const factory = match ? `GM${match[1]}` : "";

  if (!factory) {
    return { factory: "", lineName: text };
  }

  let lineName = text;
  const prefixes = [`${factory} - `, `${factory} `, `${factory}-`];
  for (const prefix of prefixes) {
    if (lineName.startsWith(prefix)) {
      lineName = lineName.slice(prefix.length).trim();
      break;
    }
  }

  if (lineName === factory) {
    lineName = "";
  }

  return { factory, lineName };
}

export function buildLineShiftConfigMap(list) {
  const map = new Map();
  (Array.isArray(list) ? list : []).forEach((item) => {
    const factory = String(item.factory || "").trim().toUpperCase();
    const lineName = String(item.lineName || "").trim();
    if (!factory || !lineName) return;
    map.set(lineShiftConfigKey(factory, lineName), {
      factory,
      lineName,
      enabled: Boolean(item.enabled),
      schedule: Array.isArray(item.schedule) ? item.schedule : [],
    });
  });
  return map;
}

export function resolveLineShiftForLocation(location, configMap) {
  const { factory, lineName } = parseLocationParts(location);
  const key = lineShiftConfigKey(factory, lineName);

  if (configMap instanceof Map && factory && lineName && configMap.has(key)) {
    const cfg = configMap.get(key);
    if (!cfg.enabled) {
      return { useShift: false, schedule: null, factory, lineName };
    }
    const schedule =
      Array.isArray(cfg.schedule) && cfg.schedule.length
        ? cfg.schedule
        : DEFAULT_SHIFT_SCHEDULE;
    return { useShift: true, schedule, factory, lineName };
  }

  // Legacy: GM3 tanpa config tetap multi-shift.
  if (factory === "GM3") {
    return {
      useShift: true,
      schedule: DEFAULT_SHIFT_SCHEDULE,
      factory,
      lineName,
    };
  }

  return { useShift: false, schedule: null, factory, lineName };
}

export function factoryHasEnabledShift(factory, configMap) {
  const f = String(factory || "").trim().toUpperCase();
  if (!f) return false;

  if (configMap instanceof Map) {
    let found = false;
    for (const cfg of configMap.values()) {
      if (String(cfg.factory || "").toUpperCase() !== f) continue;
      found = true;
      if (cfg.enabled && Array.isArray(cfg.schedule) && cfg.schedule.length) {
        return true;
      }
    }
    if (found) return false;
  }

  return f === "GM3";
}

/** True jika ada minimal satu line di factory yang disetel Hari Penuh. */
export function factoryHasFullDayLine(factory, configMap) {
  const f = String(factory || "").trim().toUpperCase();
  if (!f || !(configMap instanceof Map)) return false;

  for (const cfg of configMap.values()) {
    if (String(cfg.factory || "").toUpperCase() !== f) continue;
    if (!cfg.enabled || !Array.isArray(cfg.schedule) || !cfg.schedule.length) {
      return true;
    }
  }

  return false;
}

export function buildShiftOptionsFromConfigMap(factory, configMap) {
  const f = String(factory || "").trim().toUpperCase();
  const labels = new Map();

  if (configMap instanceof Map) {
    for (const cfg of configMap.values()) {
      if (String(cfg.factory || "").toUpperCase() !== f) continue;
      if (!cfg.enabled) continue;
      (cfg.schedule || []).forEach((item) => {
        const code = String(item.code || "").toUpperCase();
        if (!code.startsWith("SHIFT_")) return;
        if (!labels.has(code)) {
          labels.set(code, getShiftLabel(code, [item]));
        }
      });
    }
  }

  const options = labels.size
    ? [{ value: "CURRENT", label: "Current Shift" }]
    : [];

  if (labels.size) {
    ["SHIFT_1", "SHIFT_2", "SHIFT_3"].forEach((code) => {
      if (labels.has(code)) {
        options.push({ value: code, label: labels.get(code) });
      }
    });
    options.push({ value: "ALL", label: "All Shifts" });
  } else if (f === "GM3") {
    options.push(...GM3_SHIFT_OPTIONS);
  }

  if (!options.length || factoryHasFullDayLine(f, configMap)) {
    options.push({ value: "NORMAL", label: "Normal (Hari Penuh)" });
  }

  return options;
}
