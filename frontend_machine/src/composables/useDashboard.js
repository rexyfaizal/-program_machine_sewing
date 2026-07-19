import { computed, ref } from "vue";
import {
  getMachineOperatorReport,
  getMachineSettings,
  getProductivity,
} from "../api/machineApi";
import { getStatus } from "../utils/format";

const machines = ref([]);
const loading = ref(false);
const errorMessage = ref("");
const lastUpdate = ref("-");

const machineSettings = ref(new Map());
const activeOperatorMap = ref(new Map());

function setLastUpdate() {
  lastUpdate.value = new Date().toLocaleTimeString("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function getVal(obj, ...keys) {
  for (const key of keys) {
    if (obj && obj[key] !== undefined && obj[key] !== null) {
      return obj[key];
    }
  }

  return undefined;
}

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function extractRows(data) {
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

function normalizeMacState(value) {
  const text = String(value ?? "").trim();

  if (text === "2") return "2";
  if (text === "1") return "1";
  if (text === "0") return "0";

  return "0";
}

function getRowMacState(row) {
  return normalizeMacState(
    getVal(
      row,
      "macState",
      "MacState",
      "mac_state",
      "Macstate",
      "macstate",
      "MACSTATE"
    )
  );
}

function normalizeSetting(item) {
  return {
    uuid: String(getVal(item, "uuid", "UUID") || ""),
    customName: String(getVal(item, "customName", "CustomName") || ""),
    location: String(getVal(item, "location", "Location") || ""),

    pic: String(
      getVal(
        item,
        "pic",
        "PIC",
        "Pic",
        "picName",
        "PICName",
        "personInCharge",
        "PersonInCharge"
      ) || ""
    ),

    spv: String(
      getVal(
        item,
        "spv",
        "SPV",
        "Spv",
        "supervisor",
        "Supervisor",
        "supervisorName",
        "SupervisorName"
      ) || ""
    ),

    updatedAt: String(getVal(item, "updatedAt", "UpdatedAt") || ""),
  };
}

function parseDateTime(value) {
  const text = String(value || "").trim();

  if (!text) return null;

  const safeText = text.includes(" ") ? text.replace(" ", "T") : text;
  const d = new Date(safeText);

  if (Number.isNaN(d.getTime())) {
    return null;
  }

  return d;
}

function formatClock(value) {
  const text = String(value || "").trim();

  if (!text) return "";

  const match = text.match(/\d{4}-\d{2}-\d{2}[ T](\d{2}):(\d{2})/);

  if (match) {
    return `${match[1]}.${match[2]}`;
  }

  const d = parseDateTime(text);

  if (!d) return "";

  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");

  return `${hh}.${mm}`;
}

function formatDurationFromMs(ms) {
  const totalMinutes = Math.max(0, Math.floor(ms / 60000));
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;

  if (hours <= 0) {
    return `${minutes}m`;
  }

  return `${hours}j ${minutes}m`;
}

function formatActiveDuration(loginTime, logoutTime = "") {
  const start = parseDateTime(loginTime);

  if (!start) return "";

  const end = parseDateTime(logoutTime) || new Date();
  const diffMs = end.getTime() - start.getTime();

  return formatDurationFromMs(diffMs);
}

function formatNoteTime(value) {
  const clock = formatClock(value);
  return clock || "";
}

function normalizeOperatorNote(item) {
  const reasonName = String(
    getVal(item, "reasonName", "ReasonName", "reason_name") || ""
  ).trim();

  const reasonCode = String(
    getVal(item, "reasonCode", "ReasonCode", "reason_code") || ""
  ).trim();

  const note = String(getVal(item, "note", "Note") || "").trim();

  const createdAt = String(
    getVal(item, "createdAt", "CreatedAt", "created_at") || ""
  ).trim();

  const time = formatNoteTime(createdAt);

  let reasonText = reasonName || reasonCode || "";

  if (note) {
    reasonText = reasonText ? `${reasonText}: ${note}` : note;
  }

  return {
    id: toNumber(getVal(item, "id", "ID") || 0),
    reasonCode,
    reasonName,
    note,
    createdAt,
    time,
    text: time && reasonText ? `${time}-${reasonText}` : reasonText,
  };
}

function normalizeOperatorSession(item) {
  const operatorNik = String(
    getVal(item, "operatorNik", "OperatorNik", "operator_nik") || ""
  ).trim();

  const operatorName = String(
    getVal(item, "operatorName", "OperatorName", "operator_name") || ""
  ).trim();

  const status = String(getVal(item, "status", "Status") || "").toUpperCase();

  const loginTime = String(
    getVal(item, "loginTime", "LoginTime", "login_time") || ""
  );

  const logoutTime = String(
    getVal(item, "logoutTime", "LogoutTime", "logout_time") || ""
  );

  const uuid = String(getVal(item, "uuid", "UUID") || "").trim();

  const notesRaw =
    getVal(item, "notes", "Notes", "lastNotes", "LastNotes") || [];

  const notes = extractRows(notesRaw)
    .map(normalizeOperatorNote)
    .filter((note) => note.text);

  const notesText = notes.map((note) => note.text).join(" | ");

  const loginClock = formatClock(loginTime);
  const activeDuration = formatActiveDuration(loginTime, logoutTime);

  const operatorLoginText = loginClock ? `Login ${loginClock}` : "";
  const operatorActiveText = activeDuration ? `Aktif ${activeDuration}` : "";

  const operatorSubText = [operatorLoginText, operatorActiveText]
    .filter(Boolean)
    .join(" • ");

  return {
    id: toNumber(getVal(item, "id", "ID", "sessionId", "SessionID") || 0),
    uuid,
    operatorNik,
    operatorName,
    processName: String(
      getVal(item, "processName", "ProcessName", "process_name") || ""
    ),
    styleName: String(
      getVal(item, "styleName", "StyleName", "style_name") || ""
    ),
    loginTime,
    logoutTime,
    loginClock,
    activeDuration,
    operatorLoginText,
    operatorActiveText,
    operatorSubText,
    status,
    notes,
    notesText,
    picText:
      operatorNik && operatorName
        ? `${operatorNik} - ${operatorName}`
        : operatorName || operatorNik || "",
  };
}

async function loadMachineSettings() {
  const list = await getMachineSettings();
  const map = new Map();

  list.forEach((item) => {
    const setting = normalizeSetting(item);

    if (setting.uuid) {
      map.set(normalizeText(setting.uuid), setting);
    }
  });

  machineSettings.value = map;
}

async function loadActiveOperators(date) {
  const map = new Map();

  const data = await getMachineOperatorReport(date);
  const rows = extractRows(data).map(normalizeOperatorSession);

  rows.forEach((session) => {
    if (!session.uuid) return;

    const status = String(session.status || "").toUpperCase();

    if (status !== "ACTIVE" && status !== "OPEN") {
      return;
    }

    const key = normalizeText(session.uuid);
    const existing = map.get(key);

    if (!existing) {
      map.set(key, session);
      return;
    }

    const existingTime = parseDateTime(existing.loginTime)?.getTime() || 0;
    const currentTime = parseDateTime(session.loginTime)?.getTime() || 0;

    if (currentTime >= existingTime) {
      map.set(key, session);
    }
  });

  activeOperatorMap.value = map;
}

function normalizeRows(json) {
  const rawData = extractRows(json);

  machines.value = rawData.map((x) => {
    const uuid = String(getVal(x, "uuid", "UUID") || "-");
    const uuidKey = normalizeText(uuid);

    const procTime = toNumber(
      getVal(
        x,
        "procSec",
        "ProcSec",
        "procTimeSec",
        "ProcTimeSec",
        "procTime",
        "ProcTime",
        "process_time",
        "ProcessTime"
      ) || 0
    );

    const runtime = toNumber(
      getVal(
        x,
        "runtimeSec",
        "RuntimeSec",
        "runtime",
        "Runtime",
        "RunTime",
        "runTime"
      ) || 0
    );

    const productivity =
      runtime > 0 ? Math.min((procTime / runtime) * 100, 100) : 0;

    const status = getStatus(productivity);

    const backendName = String(
      getVal(
        x,
        "nickName",
        "NickName",
        "machineName",
        "MachineName",
        "name",
        "Name"
      ) || uuid
    );

    const originalMachineName = String(
      getVal(
        x,
        "originalNickName",
        "OriginalNickName",
        "originalMachineName",
        "OriginalMachineName"
      ) || backendName
    );

    const setting = machineSettings.value.get(uuidKey);
    const activeOperator = activeOperatorMap.value.get(uuidKey);

    const customName = setting?.customName || "";

    const locationFromApi = String(getVal(x, "location", "Location") || "");
    const location = setting?.location || locationFromApi || "-";

    const picFromApi = String(
      getVal(
        x,
        "pic",
        "PIC",
        "Pic",
        "picName",
        "PICName",
        "personInCharge",
        "PersonInCharge"
      ) || ""
    );

    const pic = activeOperator?.picText || setting?.pic || picFromApi || "";

    const operatorNote = activeOperator?.notesText || "";
    const spv = operatorNote || "";

    const macState = getRowMacState(x);

    return {
      date: String(getVal(x, "date", "Date") || ""),
      machineName: customName || backendName,
      nickName: customName || backendName,
      originalMachineName,
      customName,
      location,

      pic,

      spv,
      operatorNote,
      operatorNotes: activeOperator?.notes || [],

      operatorNik: activeOperator?.operatorNik || "",
      operatorName: activeOperator?.operatorName || "",
      operatorProcessName: activeOperator?.processName || "",
      operatorStyleName: activeOperator?.styleName || "",
      operatorLoginTime: activeOperator?.loginTime || "",
      operatorLogoutTime: activeOperator?.logoutTime || "",
      operatorLoginClock: activeOperator?.loginClock || "",
      operatorActiveDuration: activeOperator?.activeDuration || "",
      operatorLoginText: activeOperator?.operatorLoginText || "",
      operatorActiveText: activeOperator?.operatorActiveText || "",
      operatorSubText: activeOperator?.operatorSubText || "",

      ip: String(getVal(x, "ip", "IP", "lastLoginIP", "LastLoginIP") || "-"),

      uuid,
      tableName: String(getVal(x, "tableName", "TableName") || "-"),

      macState,
      MacState: macState,
      mac_state: macState,

      procTime,
      runtime,

      output: toNumber(
        getVal(x, "output", "Output", "procCounts", "ProcCounts") || 0
      ),

      complete: toNumber(getVal(x, "complete", "Complete") || 0),

      abnormal: toNumber(
        getVal(
          x,
          "incomplete",
          "Incomplete",
          "abnormal",
          "Abnormal",
          "abnormalCount",
          "AbnormalCount"
        ) || 0
      ),

      avgCT: toNumber(
        getVal(x, "avgCycle", "AvgCycle", "avgCT", "AvgCT", "avgCt", "AvgCt") ||
          0
      ),

      maxCT: toNumber(
        getVal(x, "maxCycle", "MaxCycle", "maxCT", "MaxCT") || 0
      ),

      minCT: toNumber(
        getVal(x, "minCycle", "MinCycle", "minCT", "MinCT") || 0
      ),

      alarm: toNumber(
        getVal(x, "alarmCount", "AlarmCount", "alarm", "Alarm") || 0
      ),

      alarmTypes: String(getVal(x, "alarmTypes", "AlarmTypes") || "-"),

      program: String(
        getVal(
          x,
          "topFile",
          "TopFile",
          "program",
          "Program",
          "programName",
          "ProgramName"
        ) || "-"
      ),

      firstProcess: String(getVal(x, "firstProcess", "FirstProcess") || "-"),
      lastProcess: String(getVal(x, "lastProcess", "LastProcess") || "-"),

      mainSource: String(getVal(x, "mainSource", "MainSource") || "process_time"),

      productivity,
      status,
    };
  });

  setLastUpdate();
}

async function loadDashboard(date) {
  loading.value = true;
  errorMessage.value = "";

  try {
    try {
      await loadMachineSettings();
    } catch (settingErr) {
      console.warn("Gagal load machine settings:", settingErr);
    }

    try {
      await loadActiveOperators(date);
    } catch (operatorErr) {
      console.warn("Gagal load active operators:", operatorErr);
      activeOperatorMap.value = new Map();
    }

    const data = await getProductivity(date);
    normalizeRows(data);
  } catch (err) {
    errorMessage.value = `Gagal mengambil data dari backend: ${err.message}`;
  } finally {
    loading.value = false;
  }
}

function makeSummary(list) {
  const totalMachine = list.length;
  const totalProductivity = list.reduce((sum, m) => sum + m.productivity, 0);

  return {
    totalMachine,
    avgProductivity: totalMachine ? totalProductivity / totalMachine : 0,
    good: list.filter((m) => m.status === "GOOD").length,
    normal: list.filter((m) => m.status === "NORMAL").length,
    bad: list.filter((m) => m.status === "BAD").length,
    totalOutput: list.reduce((sum, m) => sum + m.output, 0),
    totalAlarm: list.reduce((sum, m) => sum + m.alarm, 0),
    totalSewingTime: list.reduce((sum, m) => sum + m.procTime, 0),
  };
}

const allSummary = computed(() => makeSummary(machines.value));

export function useDashboard() {
  return {
    machines,
    loading,
    errorMessage,
    lastUpdate,
    machineSettings,
    activeOperatorMap,
    allSummary,
    loadDashboard,
    loadMachineSettings,
    loadActiveOperators,
    normalizeRows,
    makeSummary,
  };
}