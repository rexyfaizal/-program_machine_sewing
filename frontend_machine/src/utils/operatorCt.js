export const KAP_JAM_DIVISOR = 3060;
/** Default jam shift untuk Produktivitas CT (GM1 / GM2 / area lain). */
export const KAP_JAM_SHIFT_HOURS = 8;
/** Jam shift khusus GM3. */
export const KAP_JAM_SHIFT_HOURS_GM3 = 7;

export function getCtShiftHours(area) {
  const text = String(area || "")
    .trim()
    .toUpperCase();
  if (text === "GM3") return KAP_JAM_SHIFT_HOURS_GM3;
  return KAP_JAM_SHIFT_HOURS;
}

export function calcProduktivitasCtPct(numerator, kapPerJam, area) {
  const value = Number(numerator);
  const kap = Number(kapPerJam);
  if (!Number.isFinite(value) || !Number.isFinite(kap) || kap <= 0) {
    return null;
  }

  const hours = getCtShiftHours(area);
  return Number(((value / (kap * hours)) * 100).toFixed(2));
}

export function formatProduktivitasCtPct(value) {
  if (value === null || value === undefined || value === "") return "-";
  const num = Number(value);
  if (!Number.isFinite(num)) return "-";
  return `${num.toFixed(2)}%`;
}

export function attachProduktivitasCtFields(row) {
  const fallbackKap = Number(row?.kapPerJam);
  const kapPerJamCalc =
    row?.kapPerJamCalc ??
    (Number.isFinite(fallbackKap) && fallbackKap > 0 ? fallbackKap : null);
  const hasTarget =
    row?.outputTarget !== null && row?.outputTarget !== undefined;
  const area = row?.area;

  return {
    ...row,
    ctShiftHours: getCtShiftHours(area),
    produktivitasCt: calcProduktivitasCtPct(row?.output, kapPerJamCalc, area),
    produktivitasCtTargetan: hasTarget
      ? calcProduktivitasCtPct(row.outputTarget, kapPerJamCalc, area)
      : null,
  };
}

export function normalizeCtLine(value) {
  let text = String(value || "").trim();
  if (!text) return "";

  const dashIdx = text.indexOf(" - ");
  if (dashIdx > 0) {
    text = text.slice(0, dashIdx).trim();
  }

  return text.replace(/\s+/g, " ");
}

export function normalizeCtKeyPart(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

export function buildOperatorCtKey(uuid, line) {
  const uuidKey = normalizeCtKeyPart(uuid);
  const lineKey = normalizeCtKeyPart(normalizeCtLine(line));
  if (!uuidKey || !lineKey) return "";
  return `${uuidKey}||${lineKey}`;
}

export function calcKapPerJamRaw(ctSum) {
  const value = Number(ctSum);
  if (!Number.isFinite(value) || value <= 0) return null;
  return KAP_JAM_DIVISOR / value;
}

export function calcKapPerJam(ctSum) {
  const raw = calcKapPerJamRaw(ctSum);
  if (raw === null) return "";
  return String(Math.round(raw));
}

export function formatCtNumber(value) {
  if (value === null || value === undefined || value === "") return "-";
  const num = Number(value);
  if (!Number.isFinite(num)) return "-";
  if (num === 0) return "0";
  return Number.isInteger(num) ? String(num) : String(Number(num.toFixed(2)));
}

export function buildOperatorCtMap(records = []) {
  const map = new Map();

  (Array.isArray(records) ? records : []).forEach((item) => {
    const key = buildOperatorCtKey(item?.uuid, item?.line);
    if (!key) return;

    map.set(key, {
      ctSum: Number(item.ctSum || 0),
      ctStd: Number(item.ctStd || 0),
      ctValue: Number(item.ctValue || 0),
    });
  });

  return map;
}

export function attachOperatorCtFields(row, ctMap) {
  const key = buildOperatorCtKey(row?.uuid, row?.location);
  const ct = ctMap?.get(key);

  if (!ct) {
    return {
      ...row,
      ctSum: null,
      ctStd: null,
      ctValue: null,
      kapPerJam: "",
      kapPerJamCalc: null,
    };
  }

  return {
    ...row,
    ctSum: ct.ctSum ?? 0,
    ctStd: ct.ctStd ?? 0,
    ctValue: ct.ctValue ?? 0,
    kapPerJam: calcKapPerJam(ct.ctSum),
    kapPerJamCalc: calcKapPerJamRaw(ct.ctSum),
  };
}
