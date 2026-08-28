import { buildOperatorCtKey, normalizeCtLine } from "./operatorCt";

export function formatTemplateOperatorNik(row) {
  return String(row?.operatorNik || "").trim() || "-";
}

export function formatTemplateOperatorName(row) {
  const name = String(row?.operatorName || "").trim();
  if (name) return name;
  return row?.loggedIn ? "-" : "Not logged in";
}

export function hasTemplateOperatorInfo(row) {
  if (!row?.loggedIn) return false;
  return Boolean(
    String(row?.operatorNik || "").trim() || String(row?.operatorName || "").trim()
  );
}

export function upsertOperatorTemplateRow(uniqueMap, row, buildEntry) {
  const uuid = String(row?.uuid || "").trim();
  const line = normalizeCtLine(row?.location || row?.locationLabel || "");
  const area = String(row?.area || "").trim();

  if (!uuid || !line || line === "-") return;

  const key = buildOperatorCtKey(uuid, line);
  if (!key) return;

  const existing = uniqueMap.get(key);
  if (!existing) {
    uniqueMap.set(key, buildEntry(row, { area, line, uuid }));
    return;
  }

  if (hasTemplateOperatorInfo(row) && !existing._hasOperatorInfo) {
    uniqueMap.set(key, buildEntry(row, { area, line, uuid }));
  }
}
