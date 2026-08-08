async function readResponse(res) {
  const text = await res.text();

  let data = null;

  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }

  if (!res.ok) {
    const message =
      (typeof data === "string" && data.trim()) ||
      data?.message ||
      data?.error ||
      data?.Message ||
      text ||
      `HTTP ${res.status}`;

    throw new Error(String(message).trim());
  }

  return data;
}

function normalizeList(data) {
  if (Array.isArray(data)) return data;

  return (
    data?.rows ||
    data?.Rows ||
    data?.data ||
    data?.Data ||
    data?.items ||
    data?.Items ||
    data?.machines ||
    data?.Machines ||
    []
  );
}

export async function getProductivity(date, options = {}) {
  const params = new URLSearchParams();

  if (date) {
    params.set("date", date);
  }

  const shift = String(options.shift || "").trim();
  if (shift) {
    params.set("shift", shift);
  }

  const query = params.toString() ? `?${params.toString()}` : "";

  const res = await fetch(`/api/productivity${query}`);

  return await readResponse(res);
}

export async function getProcessDetail(arg1, arg2) {
  let date = "";
  let uuid = "";

  const first = String(arg1 || "");
  const second = String(arg2 || "");

  if (/^\d{4}-\d{2}-\d{2}$/.test(first)) {
    date = first;
    uuid = second;
  } else {
    uuid = first;
    date = second;
  }

  const params = new URLSearchParams();

  if (date) params.set("date", date);
  if (uuid) params.set("uuid", uuid);

  const res = await fetch(`/api/process/detail?${params.toString()}`);

  return await readResponse(res);
}

export async function getMachineSettings() {
  const res = await fetch("/api/machine-settings");
  const data = await readResponse(res);

  return normalizeList(data);
}

export async function saveMachineSetting(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch("/api/machine-settings", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      uuid: String(payload.uuid || "").trim(),
      customName: String(payload.customName || "").trim(),
      location: String(payload.location || "").trim(),
      pic: String(payload.pic || "").trim(),
      spv: String(payload.spv || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function updateMachineSetting(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch("/api/machine-settings", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      uuid: String(payload.uuid || "").trim(),
      customName: String(payload.customName || "").trim(),
      location: String(payload.location || "").trim(),
      pic: String(payload.pic || "").trim(),
      spv: String(payload.spv || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function deleteMachineSetting(uuid) {
  if (!uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch(
    `/api/machine-settings?uuid=${encodeURIComponent(uuid)}`,
    {
      method: "DELETE",
    }
  );

  return await readResponse(res);
}

export async function getLineShiftConfig(factory = "") {
  const params = new URLSearchParams();
  const factoryText = String(factory || "").trim();
  if (factoryText) {
    params.set("factory", factoryText);
  }

  const query = params.toString() ? `?${params.toString()}` : "";
  const res = await fetch(`/api/line-shift-config${query}`);
  return await readResponse(res);
}

export async function saveLineShiftConfig(payload) {
  const factory = String(payload?.factory || "").trim();
  if (!factory) {
    throw new Error("Factory wajib diisi.");
  }

  const res = await fetch("/api/line-shift-config", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      factory,
      lines: Array.isArray(payload?.lines) ? payload.lines : [],
    }),
  });

  return await readResponse(res);
}

export async function getShiftSettings(area = "") {
  const areaText = String(area || "").trim();
  if (!areaText) {
    throw new Error("Area wajib diisi.");
  }

  const params = new URLSearchParams();
  params.set("area", areaText);

  const res = await fetch(`/api/shift-settings?${params.toString()}`);
  return await readResponse(res);
}

export async function saveShiftSettings(payload) {
  const area = String(payload?.area || payload?.factory || "").trim();
  if (!area) {
    throw new Error("Area wajib diisi.");
  }

  const res = await fetch("/api/shift-settings", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      area,
      shifts: Array.isArray(payload?.shifts) ? payload.shifts : [],
      lines: Array.isArray(payload?.lines)
        ? payload.lines.map((line) => ({
            lineName: line.lineName,
            enabled: Boolean(line.enabled),
            custom: Boolean(line.custom),
            shifts: Array.isArray(line.shifts) ? line.shifts : [],
          }))
        : [],
    }),
  });

  return await readResponse(res);
}

export async function searchEmployees(query) {
  const q = String(query || "").trim();

  if (!q) return [];

  const res = await fetch(`/api/employees/search?q=${encodeURIComponent(q)}`);
  const data = await readResponse(res);

  return normalizeList(data);
}

export async function getActiveMachineOperator(uuid) {
  const id = String(uuid || "").trim();

  if (!id) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch(
    `/api/machine-operator/active?uuid=${encodeURIComponent(id)}`
  );

  return await readResponse(res);
}

export async function loginMachineOperator(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  if (!payload?.operatorNik) {
    throw new Error("NIK operator wajib diisi.");
  }

  const res = await fetch("/api/machine-operator/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      uuid: String(payload.uuid || "").trim(),
      machineName: String(payload.machineName || "").trim(),
      location: String(payload.location || "").trim(),

      operatorNik: String(payload.operatorNik || "").trim(),
      operatorName: String(payload.operatorName || "").trim(),
      branchdetail: String(payload.branchdetail || "").trim(),

      processName: String(payload.processName || "").trim(),
      styleName: String(payload.styleName || "").trim(),

      shiftCode: String(payload.shiftCode || "").trim(),
      shiftName: String(payload.shiftName || "").trim(),

      forceReplace: Boolean(payload.forceReplace || false),
    }),
  });

  return await readResponse(res);
}

export async function saveMachineOperatorNote(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  if (!payload?.sessionId) {
    throw new Error("Session operator wajib diisi.");
  }

  if (!payload?.reasonCode) {
    throw new Error("Keterangan wajib dipilih.");
  }

  const res = await fetch("/api/machine-operator/note", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      sessionId: Number(payload.sessionId || 0),
      uuid: String(payload.uuid || "").trim(),
      reasonCode: String(payload.reasonCode || "").trim(),
      reasonName: String(payload.reasonName || "").trim(),
      note: String(payload.note || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function startMachineOperatorLossEvent(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  if (!payload?.reasonCode) {
    throw new Error("Keterangan loss time wajib dipilih.");
  }

  const res = await fetch("/api/machine-operator/loss-event/start", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      uuid: String(payload.uuid || "").trim(),
      reasonCode: String(payload.reasonCode || "").trim(),
      reasonLabel: String(payload.reasonLabel || "").trim(),
      note: String(payload.note || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function finishMachineOperatorLossEvent(payload) {
  if (!payload?.uuid) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch("/api/machine-operator/loss-event/finish", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      uuid: String(payload.uuid || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function getActiveMachineOperatorLossEvent(uuid) {
  const id = String(uuid || "").trim();

  if (!id) {
    throw new Error("UUID mesin wajib diisi.");
  }

  const res = await fetch(
    `/api/machine-operator/loss-event/active?uuid=${encodeURIComponent(id)}`
  );

  return await readResponse(res);
}

export async function getMachineOperatorReport(date, options = {}) {
  const d = String(date || "").trim();

  if (!d) {
    throw new Error("Tanggal report wajib diisi.");
  }

  const withStats =
    options?.withStats === true ||
    options?.withStats === 1 ||
    String(options?.withStats || "").trim() === "1";

  const params = new URLSearchParams({ date: d });
  if (withStats) {
    params.set("withStats", "1");
  }

  const res = await fetch(`/api/machine-operator/report?${params.toString()}`);

  return await readResponse(res);
}

export async function getProcessStyleList(query = "") {
  const q = String(query || "").trim();
  const res = await fetch(`/api/process-style/list?q=${encodeURIComponent(q)}`);
  const data = await readResponse(res);

  return normalizeList(data);
}

export async function createProcessStyle(payload) {
  const styleName = String(payload?.styleName || "").trim();
  const processName = String(payload?.processName || "").trim();

  if (!styleName) throw new Error("Style wajib diisi.");
  if (!processName) throw new Error("Proses wajib diisi.");

  const res = await fetch("/api/process-style", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      styleName,
      processName,
    }),
  });

  return await readResponse(res);
}

export async function updateProcessStyle(id, payload) {
  const rowId = Number(id || 0);
  const styleName = String(payload?.styleName || "").trim();
  const processName = String(payload?.processName || "").trim();

  if (!rowId) throw new Error("ID data wajib diisi.");
  if (!styleName) throw new Error("Style wajib diisi.");
  if (!processName) throw new Error("Proses wajib diisi.");

  const res = await fetch(`/api/process-style/${encodeURIComponent(rowId)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      styleName,
      processName,
    }),
  });

  return await readResponse(res);
}

export async function deleteProcessStyle(id) {
  const rowId = Number(id || 0);

  if (!rowId) throw new Error("ID data wajib diisi.");

  const res = await fetch(`/api/process-style/${encodeURIComponent(rowId)}`, {
    method: "DELETE",
  });

  return await readResponse(res);
}

export async function searchStyles(query) {
  const q = String(query || "").trim();

  if (!q) return [];

  const res = await fetch(`/api/process-style/styles?q=${encodeURIComponent(q)}`);
  const data = await readResponse(res);

  return normalizeList(data);
}

export async function searchProcessesByStyle(style, query = "") {
  const s = String(style || "").trim();
  const q = String(query || "").trim();

  if (!s) return [];

  const params = new URLSearchParams();
  params.set("style", s);

  if (q) {
    params.set("q", q);
  }

  const res = await fetch(`/api/process-style/processes?${params.toString()}`);
  const data = await readResponse(res);

  return normalizeList(data);
}

export async function importProcessStyles(rows) {
  if (!Array.isArray(rows)) {
    throw new Error("Data import harus berupa array.");
  }

  const cleanRows = rows
    .map((row) => {
      return {
        style: String(row?.style || "").trim(),
        processName: String(row?.processName || "").trim(),
      };
    })
    .filter((row) => row.style && row.processName);

  if (!cleanRows.length) {
    throw new Error("Tidak ada data valid untuk diimport.");
  }

  const res = await fetch("/api/process-style/import", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      rows: cleanRows,
    }),
  });

  return await readResponse(res);
}

export async function identifyMechanic(code) {
  const value = String(code || "").trim();
  if (!value) {
    throw new Error("NIK / kartu RFID wajib diisi.");
  }

  const res = await fetch("/api/mechanic/identify", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ code: value, nik: value }),
  });

  return await readResponse(res);
}

export async function registerMechanicRFID(payload) {
  const nik = String(payload?.nik || "").trim();
  const rfidNo = String(payload?.rfidNo || "").trim();

  if (!nik) {
    throw new Error("NIK mekanik wajib diisi.");
  }
  if (!rfidNo) {
    throw new Error("Nomor kartu RFID wajib diisi.");
  }

  const res = await fetch("/api/mechanic/rfid/register", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ nik, rfidNo }),
  });

  return await readResponse(res);
}

export async function getMechanicBrokenMachines(options = {}) {
  const params = new URLSearchParams();
  const status = String(options.status || "ALL").trim();
  const location = String(options.location || "").trim();

  if (status) params.set("status", status);
  if (location) params.set("location", location);

  const query = params.toString();
  const res = await fetch(
    `/api/mechanic/broken-machines${query ? `?${query}` : ""}`
  );

  return await readResponse(res);
}

export async function claimMechanicBrokenMachine(payload) {
  if (!payload?.id) {
    throw new Error("ID tiket wajib diisi.");
  }
  if (!payload?.mechanicNik) {
    throw new Error("NIK mekanik wajib diisi.");
  }

  const res = await fetch("/api/mechanic/claim", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: Number(payload.id || 0),
      mechanicNik: String(payload.mechanicNik || "").trim(),
      mechanicName: String(payload.mechanicName || "").trim(),
    }),
  });

  return await readResponse(res);
}

export async function doneMechanicBrokenMachine(payload) {
  if (!payload?.id) {
    throw new Error("ID tiket wajib diisi.");
  }
  if (!payload?.mechanicNik) {
    throw new Error("NIK mekanik wajib diisi.");
  }

  const res = await fetch("/api/mechanic/done", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: Number(payload.id || 0),
      mechanicNik: String(payload.mechanicNik || "").trim(),
    }),
  });

  return await readResponse(res);
}