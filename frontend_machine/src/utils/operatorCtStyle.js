import {
  calcKapPerJam,
  calcKapPerJamRaw,
  normalizeCtKeyPart,
} from "./operatorCt";

export function normalizeCtStylePart(value) {
  return normalizeCtKeyPart(value);
}

export function buildOperatorCtStyleKey(style, processName) {
  const styleKey = normalizeCtStylePart(style);
  const processKey = normalizeCtStylePart(processName);
  if (!styleKey || !processKey || styleKey === "-" || processKey === "-") {
    return "";
  }
  return `${styleKey}||${processKey}`;
}

export function buildOperatorCtStyleMap(records = []) {
  const map = new Map();

  (Array.isArray(records) ? records : []).forEach((item) => {
    const key = buildOperatorCtStyleKey(item?.styleName, item?.processName);
    if (!key) return;

    map.set(key, {
      ctTotal: Number(item.ctTotal || 0),
    });
  });

  return map;
}

export function attachOperatorCtStyleFields(row, styleCtMap) {
  const key = buildOperatorCtStyleKey(row?.style, row?.mesin);
  const ct = styleCtMap?.get(key);

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

  const ctTotal = ct.ctTotal ?? 0;

  return {
    ...row,
    ctSum: ctTotal,
    ctStd: null,
    ctValue: ctTotal,
    kapPerJam: calcKapPerJam(ctTotal),
    kapPerJamCalc: calcKapPerJamRaw(ctTotal),
  };
}

export function isGm3Area(area) {
  return String(area || "").trim().toUpperCase() === "GM3";
}
