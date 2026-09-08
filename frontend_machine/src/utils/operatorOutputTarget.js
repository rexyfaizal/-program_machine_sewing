import { buildOperatorCtKey } from "./operatorCt";
import { buildOperatorCtStyleKey } from "./operatorCtStyle";

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

export function buildOperatorOutputTargetStyleMap(records = []) {
  const map = new Map();

  (Array.isArray(records) ? records : []).forEach((item) => {
    const key = buildOperatorCtStyleKey(item?.styleName, item?.processName);
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

export function attachOperatorOutputTargetStyleFields(row, styleTargetMap) {
  const key = buildOperatorCtStyleKey(row?.style, row?.mesin);
  const target = styleTargetMap?.get(key);

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
