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
      data?.message ||
      data?.error ||
      data?.Message ||
      text ||
      `HTTP ${res.status}`;

    throw new Error(message);
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

export async function getProductivity(date) {
  const query = date ? `?date=${encodeURIComponent(date)}` : "";

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

export async function getMachineOperatorReport(date) {
  const d = String(date || "").trim();

  if (!d) {
    throw new Error("Tanggal report wajib diisi.");
  }

  const res = await fetch(
    `/api/machine-operator/report?date=${encodeURIComponent(d)}`
  );

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