import { parseLocationParts } from "./gm3Shift";

export function normalizeSpvLocationKey(location) {
  return String(location || "")
    .trim()
    .replace(/\s+/g, " ")
    .toUpperCase();
}

function addMapEntry(map, location, value) {
  const key = normalizeSpvLocationKey(location);
  if (!key || !value?.spvNik) return;
  map.set(key, value);

  const { factory, lineName } = parseLocationParts(location);
  if (factory && lineName) {
    map.set(normalizeSpvLocationKey(`${factory} ${lineName}`), value);
    map.set(normalizeSpvLocationKey(`${factory} - ${lineName}`), value);
  }
}

export function buildSpvAssignmentMap(assignments = []) {
  const map = new Map();

  (Array.isArray(assignments) ? assignments : []).forEach((row) => {
    const location = String(row?.location || "").trim();
    const spvNik = String(row?.spvNik || "").trim();
    if (!location || !spvNik) return;

    const value = {
      location,
      spvNik,
      spvName: String(row?.spvName || "").trim(),
      branchdetail: String(row?.branchdetail || "").trim(),
    };

    addMapEntry(map, location, value);
  });

  return map;
}

export function resolveSpvAssignment(location, assignmentMap) {
  if (!(assignmentMap instanceof Map) || !assignmentMap.size) return null;

  const raw = String(location || "").trim();
  if (!raw || raw === "-") return null;

  const direct = assignmentMap.get(normalizeSpvLocationKey(raw));
  if (direct) return direct;

  const { factory, lineName } = parseLocationParts(raw);
  if (!factory || !lineName) return null;

  return (
    assignmentMap.get(normalizeSpvLocationKey(`${factory} ${lineName}`)) ||
    assignmentMap.get(normalizeSpvLocationKey(`${factory} - ${lineName}`)) ||
    null
  );
}

export function formatSpvId(assignment) {
  return String(assignment?.spvNik || "").trim();
}
