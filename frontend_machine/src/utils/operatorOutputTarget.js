import { buildOperatorCtKey } from "./operatorCt";

export function formatOutputTarget(value) {
  if (value === null || value === undefined || value === "") return "-";
  const num = Number(value);
  if (!Number.isFinite(num)) return "-";
  return String(Math.round(num));
}

export function buildOperatorOutputTargetMap(records = []) {
  const map = new Map();

  (Array.isArray(records) ? records : []).forEach((item) => {
    const key = buildOperatorCtKey(item?.uuid, item?.line);
    if (!key) return;

    map.set(key, {
      outputTarget: Number(item.outputTarget || 0),
    });
  });

  return map;
}

export function attachOperatorOutputTargetFields(row, targetMap) {
  const key = buildOperatorCtKey(row?.uuid, row?.location);
  const target = targetMap?.get(key);

  if (!target) {
    return {
      ...row,
      outputTarget: null,
    };
  }

  return {
    ...row,
    outputTarget: target.outputTarget ?? 0,
  };
}
