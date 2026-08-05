import { ref } from "vue";
import {
  getMachineOperatorReport,
  getMachineSettings,
  getLineShiftConfig,
  getProductivity,
} from "../api/machineApi";
import { isAutoLogoutText } from "../utils/dashboardExportExcel";
import {
  buildLineShiftConfigMap,
  formatLocalDate,
  getGM3ShiftWindow,
  resolveLineShiftForLocation,
} from "../utils/gm3Shift";

const WORK_SECONDS_PER_DAY = 28800;

export function useDashboard() {
  const machines = ref([]);
  const loading = ref(false);
  const errorMessage = ref("");
  const lastUpdate = ref("-");
  const machineSettings = ref(new Map());
  const activeOperatorMap = ref(new Map());
  const shiftConfigMap = ref(new Map());
  const loadedDashboardDate = ref("");
  const loadedDashboardShift = ref("");

  let dashboardRequestSeq = 0;
  let inFlightRequest = null;
  let inFlightDate = "";

  function isLoadedDateToday() {
    const loaded = String(loadedDashboardDate.value || "").trim();
    if (!loaded) return true;
    return loaded === formatLocalDate(new Date());
  }

  // Window jam shift terpilih untuk sebuah mesin. null = tanpa filter shift.
  function getShiftWindowForLocation(location) {
    const shiftCode = String(loadedDashboardShift.value || "")
      .trim()
      .toUpperCase();

    if (!["SHIFT_1", "SHIFT_2", "SHIFT_3", "CURRENT"].includes(shiftCode)) {
      return null;
    }

    const lineShift = resolveLineShiftForLocation(location, shiftConfigMap.value);
    if (!lineShift.useShift) return null;

    return getGM3ShiftWindow(
      loadedDashboardDate.value,
      shiftCode,
      lineShift.schedule
    );
  }

  // True jika rentang [start, end) overlap dengan window shift.
  function overlapsShiftWindow(startValue, endValue, isActive, window) {
    const start = parseDateTime(startValue);
    if (!start) return false;

    const end = isActive ? new Date() : parseDateTime(endValue);

    return (
      start.getTime() < window.end.getTime() &&
      (!end || end.getTime() > window.start.getTime())
    );
  }

  function getSessionBounds(item) {
    const start = parseDateTime(item?.loginTime);
    if (!start) return null;

    const status = String(item?.status || "").toUpperCase();
    const isActive =
      ["ACTIVE", "OPEN"].includes(status) &&
      !String(item?.logoutTime || "").trim();
    const end = isActive ? new Date() : parseDateTime(item?.logoutTime);

    if (!end || end.getTime() <= start.getTime()) return null;

    return { start, end, isActive };
  }

  // Durasi overlap session vs window dalam detik. 0 = tidak overlap / terlalu singkat.
  function getOverlapSeconds(bounds, window) {
    if (!bounds || !window) return 0;

    const startMs = Math.max(bounds.start.getTime(), window.start.getTime());
    const endMs = Math.min(bounds.end.getTime(), window.end.getTime());
    if (endMs <= startMs) return 0;

    return Math.floor((endMs - startMs) / 1000);
  }

  // Session dimiliki 1 shift saja: yang overlap terbesar. Minimal 1 menit.
  function resolveSessionOwnerShiftCode(item, location) {
    const bounds = getSessionBounds(item);
    if (!bounds) return null;

    const durationSec = Math.floor(
      (bounds.end.getTime() - bounds.start.getTime()) / 1000
    );
    // Session sangat singkat (0 menit / < 1 menit) tidak dimasukkan.
    if (durationSec < 60) return null;

    const lineShift = resolveLineShiftForLocation(
      location,
      shiftConfigMap.value
    );
    if (!lineShift.useShift) return null;

    const workDate = String(loadedDashboardDate.value || "").trim();
    if (!workDate) return null;

    let bestCode = null;
    let bestOverlap = 0;

    for (const code of ["SHIFT_1", "SHIFT_2", "SHIFT_3"]) {
      const window = getGM3ShiftWindow(workDate, code, lineShift.schedule);
      const overlap = getOverlapSeconds(bounds, window);
      if (overlap > bestOverlap) {
        bestOverlap = overlap;
        bestCode = code;
      }
    }

    return bestOverlap > 0 ? bestCode : null;
  }

  // Label shift mengikuti filter dashboard / jam sistem (bukan jam login).
  function resolveCurrentShiftCodeFromSchedule(schedule, now = new Date()) {
    const today = formatLocalDate(now);
    const yesterdayDate = new Date(now.getTime());
    yesterdayDate.setDate(yesterdayDate.getDate() - 1);
    const yesterday = formatLocalDate(yesterdayDate);

    for (const dateText of [today, yesterday]) {
      for (const code of ["SHIFT_1", "SHIFT_2", "SHIFT_3"]) {
        const window = getGM3ShiftWindow(dateText, code, schedule);
        if (
          window &&
          now.getTime() >= window.start.getTime() &&
          now.getTime() < window.end.getTime()
        ) {
          return code;
        }
      }
    }

    return null;
  }

  function resolveOperatorShiftTag(location) {
    const lineShift = resolveLineShiftForLocation(
      location,
      shiftConfigMap.value
    );

    if (!lineShift.useShift) {
      return { label: "Normal", code: "NORMAL" };
    }

    const selected = String(loadedDashboardShift.value || "")
      .trim()
      .toUpperCase();

    let code = selected;

    if (!code || code === "CURRENT") {
      code =
        resolveCurrentShiftCodeFromSchedule(lineShift.schedule) || "ALL";
    }

    if (code === "SHIFT_1") return { label: "Shift 1", code: "SHIFT_1" };
    if (code === "SHIFT_2") return { label: "Shift 2", code: "SHIFT_2" };
    if (code === "SHIFT_3") return { label: "Shift 3", code: "SHIFT_3" };
    if (code === "ALL") return { label: "All Shifts", code: "ALL" };
    if (code === "NORMAL") return { label: "Normal", code: "NORMAL" };

    return { label: "Normal", code: "NORMAL" };
  }

  function setLastUpdate() {
    lastUpdate.value = new Date().toLocaleString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
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

  function parseDateTime(value) {
    if (!value) return null;

    const raw = String(value || "").trim();

    if (!raw) return null;

    const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
    const d = new Date(normalized);

    if (Number.isNaN(d.getTime())) return null;

    return d;
  }

  function formatClock(value) {
    if (!value) return "";

    const d = parseDateTime(value);

    if (!d) {
      const text = String(value || "").replace("T", " ");
      const timePart = text.slice(11, 16);

      if (timePart) {
        return timePart.replace(":", ".");
      }

      return text.slice(0, 16);
    }

    return d
      .toLocaleTimeString("id-ID", {
        hour: "2-digit",
        minute: "2-digit",
      })
      .replace(":", ".");
  }

  function formatDurationSeconds(seconds) {
    const totalSeconds = Math.max(0, Math.floor(Number(seconds || 0)));
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const secs = totalSeconds % 60;

    return [hours, minutes, secs]
      .map((item) => String(item).padStart(2, "0"))
      .join(":");
  }

  function formatDurationFromMs(ms) {
    const totalSeconds = Math.max(0, Math.floor(Number(ms || 0) / 1000));
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}j ${minutes}m`;
    }

    return `${minutes}m`;
  }

  function formatActiveDuration(loginTime) {
    const start = parseDateTime(loginTime);

    if (!start) return "";

    return formatDurationFromMs(Date.now() - start.getTime());
  }

  function formatSessionUsageDuration(loginTime, logoutTime, status) {
    const start = parseDateTime(loginTime);

    if (!start) return "";

    const statusUpper = String(status || "").toUpperCase();
    const end = parseDateTime(logoutTime);
    const isActive = ["ACTIVE", "OPEN"].includes(statusUpper) && !end;

    if (isActive) {
      return formatDurationFromMs(Date.now() - start.getTime());
    }

    if (!end) return "";

    return formatDurationFromMs(end.getTime() - start.getTime());
  }

  function buildOperatorSessionSubText(item) {
    const loginClock = String(item?.operatorLoginClock || "").trim();
    const logoutClock = String(item?.operatorLogoutClock || "").trim();
    const usageDuration = String(item?.operatorUsageDuration || "").trim();
    const statusUpper = String(item?.status || "").toUpperCase();
    const isActive =
      ["ACTIVE", "OPEN"].includes(statusUpper) && !item?.logoutTime;

    if (loginClock && logoutClock && usageDuration) {
      return `Login ${loginClock}–${logoutClock} · Pakai ${usageDuration}`;
    }

    if (loginClock && isActive && usageDuration) {
      return `Login ${loginClock} · Aktif ${usageDuration}`;
    }

    if (loginClock && usageDuration) {
      return `Login ${loginClock} · Pakai ${usageDuration}`;
    }

    if (loginClock) {
      return `Login ${loginClock}`;
    }

    return "";
  }

  function normalizeMacState(row) {
    const value = String(
      getVal(row, "macState", "MacState", "mac_state", "Macstate", "macstate") ??
        ""
    ).trim();

    if (value === "2") return "2";
    if (value === "1") return "1";

    return "0";
  }

  function normalizeSetting(row) {
    return {
      uuid: String(getVal(row, "uuid", "UUID") || "").trim(),
      customName: String(
        getVal(row, "customName", "CustomName", "custom_name") || ""
      ).trim(),
      location: String(getVal(row, "location", "Location") || "").trim(),
      pic: String(getVal(row, "pic", "PIC") || "").trim(),
      spv: String(getVal(row, "spv", "SPV") || "").trim(),
    };
  }

  function buildMachineSettingsMap(data) {
    const map = new Map();

    extractRows(data)
      .map(normalizeSetting)
      .filter((item) => item.uuid)
      .forEach((item) => {
        map.set(normalizeText(item.uuid), item);
      });

    return map;
  }

  function getProductivityStatus(productivity) {
    const value = Number(productivity || 0);

    if (value >= 90) return "GOOD";
    if (value >= 80) return "NORMAL";

    return "BAD";
  }

  function normalizeOperatorNote(row) {
    const reasonCode = String(
      getVal(row, "reasonCode", "ReasonCode", "reason_code") || ""
    ).trim();

    const reasonName = String(
      getVal(
        row,
        "reasonName",
        "ReasonName",
        "reasonLabel",
        "ReasonLabel",
        "reason_name",
        "reason_label"
      ) || ""
    ).trim();

    const note = String(getVal(row, "note", "Note") || "").trim();

    const createdAt = String(
      getVal(
        row,
        "createdAt",
        "CreatedAt",
        "created_at",
        "startTime",
        "StartTime",
        "start_time"
      ) || ""
    ).trim();

    const endTime = String(
      getVal(row, "endTime", "EndTime", "end_time") || ""
    ).trim();

    const rawStatus = String(getVal(row, "status", "Status") || "").trim();
    const status = rawStatus.toUpperCase();

    // Sembunyikan semua note Auto Logout di dashboard.
    if (
      isAutoLogoutText(reasonCode) ||
      isAutoLogoutText(reasonName) ||
      isAutoLogoutText(note) ||
      isAutoLogoutText(rawStatus) ||
      String(status || "").includes("AUTO_LOGOUT")
    ) {
      return null;
    }

    let durationSeconds = toNumber(
      getVal(
        row,
        "durationSeconds",
        "DurationSeconds",
        "duration_sec",
        "duration"
      ) || 0
    );

    const startDate = parseDateTime(createdAt);

    if ((status === "ACTIVE" || status === "OPEN") && startDate) {
      durationSeconds = Math.max(
        0,
        Math.floor((Date.now() - startDate.getTime()) / 1000)
      );
    }

    const durationText =
      String(
        getVal(row, "durationText", "DurationText", "duration_text") || ""
      ).trim() || formatDurationSeconds(durationSeconds);

    const timeText = formatClock(createdAt);
    const endTimeText = formatClock(endTime);

    let noteText = note;

    if (status === "ACTIVE" || status === "OPEN") {
      if (!noteText || !noteText.toLowerCase().includes("sedang")) {
        noteText = `Sedang berjalan ${durationText}`;
      }
    } else if (status === "CLOSED" || endTime) {
      if (!noteText || !noteText.toLowerCase().includes("selesai")) {
        noteText = `Selesai ${durationText}`;
      }
    }

    let text = "";

    if (reasonName) {
      if (endTimeText && timeText) {
        text = `${timeText}-${endTimeText} ${reasonName}: ${noteText}`;
      } else if (timeText) {
        text = `${timeText}-${reasonName}: ${noteText}`;
      } else {
        text = `${reasonName}: ${noteText}`;
      }
    } else {
      text = noteText || "-";
    }

    return {
      id: Number(getVal(row, "id", "ID") || Date.now()),
      reasonCode,
      reasonName,
      note: noteText,
      createdAt,
      endTime,
      durationSeconds,
      durationText,
      status,
      text,
    };
  }

  function normalizeOperatorSession(row) {
    if (!row) return null;

    const uuid = String(getVal(row, "uuid", "UUID") || "").trim();

    if (!uuid) return null;

    const operatorNik = String(
      getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
    ).trim();

    const operatorName = String(
      getVal(row, "operatorName", "OperatorName", "operator_name") || ""
    ).trim();

    const processName = String(
      getVal(row, "processName", "ProcessName", "process_name") || ""
    ).trim();

    const styleName = String(
      getVal(row, "styleName", "StyleName", "style_name") || ""
    ).trim();

    const loginTime = String(
      getVal(row, "loginTime", "LoginTime", "login_time") || ""
    ).trim();

    const logoutTime = String(
      getVal(row, "logoutTime", "LogoutTime", "logout_time") || ""
    ).trim();

    const status = String(getVal(row, "status", "Status") || "").trim();

    let notes = extractRows(
      getVal(row, "notes", "Notes", "lastNotes", "LastNotes") || []
    )
      .map(normalizeOperatorNote)
      .filter((note) => note && (note.reasonName || note.note || note.text));

    const activeLossReasonCode = String(
      getVal(
        row,
        "activeLossReasonCode",
        "ActiveLossReasonCode",
        "active_loss_reason_code"
      ) || ""
    ).trim();

    const activeLossReasonLabel = String(
      getVal(
        row,
        "activeLossReasonLabel",
        "ActiveLossReasonLabel",
        "active_loss_reason_label"
      ) || ""
    ).trim();

    const activeLossStartTime = String(
      getVal(
        row,
        "activeLossStartTime",
        "ActiveLossStartTime",
        "active_loss_start_time"
      ) || ""
    ).trim();

    const activeLossDurationSeconds = toNumber(
      getVal(
        row,
        "activeLossDurationSeconds",
        "ActiveLossDurationSeconds",
        "active_loss_duration_seconds"
      ) || 0
    );

    const activeLossStatus = String(
      getVal(
        row,
        "activeLossStatus",
        "ActiveLossStatus",
        "active_loss_status"
      ) || ""
    ).trim();

    if (activeLossReasonLabel && !notes.length) {
      notes = [
        normalizeOperatorNote({
          reasonCode: activeLossReasonCode,
          reasonName: activeLossReasonLabel,
          note: `Sedang berjalan ${formatDurationSeconds(
            activeLossDurationSeconds
          )}`,
          createdAt: activeLossStartTime,
          durationSeconds: activeLossDurationSeconds,
          status: activeLossStatus || "ACTIVE",
        }),
      ];
    }

    const operatorLabel =
      operatorNik && operatorName
        ? `${operatorNik} - ${operatorName}`
        : operatorName || operatorNik || "";

    const loginClock = formatClock(loginTime);
    const logoutClock = formatClock(logoutTime);
    const usageDuration = formatSessionUsageDuration(
      loginTime,
      logoutTime,
      status
    );
    const activeDuration = ["ACTIVE", "OPEN"].includes(
      String(status || "").toUpperCase()
    )
      ? formatActiveDuration(loginTime)
      : usageDuration;

    const sessionItem = {
      id: Number(getVal(row, "id", "ID", "sessionId", "SessionID") || 0),
      uuid,
      operatorNik,
      operatorName,
      operatorLabel,
      branchdetail: String(
        getVal(row, "branchdetail", "BranchDetail", "branch_detail") || ""
      ).trim(),
      processName,
      styleName,
      loginTime,
      logoutTime,
      status,
      notes,
      operatorNote: notes.map((note) => note.text).join(" | "),
      operatorNotes: notes.map((note) => note.text).join(" | "),
      operatorLoginClock: loginClock,
      operatorLogoutClock: logoutClock,
      operatorUsageDuration: usageDuration,
      operatorActiveDuration: activeDuration,
    };

    sessionItem.operatorSubText = buildOperatorSessionSubText(sessionItem);
    sessionItem.operatorLoginText = sessionItem.operatorSubText;
    sessionItem.operatorActiveText = usageDuration || activeDuration;

    return sessionItem;
  }

  function pickPrimaryOperatorSession(sessions) {
    if (!Array.isArray(sessions) || !sessions.length) return null;

    const active = sessions.find((item) =>
      ["ACTIVE", "OPEN"].includes(String(item.status || "").toUpperCase())
    );

    if (active) return active;

    return sessions[sessions.length - 1];
  }

  // Filter session operator sesuai window shift (jika ada).
  function filterOperatorSessionsForShift(sessions, location) {
    let list = (Array.isArray(sessions) ? sessions : []).filter(
      (item) => item?.operatorLabel
    );

    const selectedShift = String(loadedDashboardShift.value || "")
      .trim()
      .toUpperCase();
    const isCurrentShift = selectedShift === "CURRENT";

    // Current Shift: hanya operator yang masih ACTIVE (hindari session selesai yang rancu).
    if (isCurrentShift) {
      list = list.filter((item) => {
        const status = String(item.status || "").toUpperCase();
        return (
          ["ACTIVE", "OPEN"].includes(status) &&
          !String(item.logoutTime || "").trim()
        );
      });
    }

    const window = getShiftWindowForLocation(location);

    if (window) {
      // Shift spesifik / Current: session hanya masuk ke 1 shift (overlap terbesar).
      // Session < 1 menit tidak ditampilkan.
      const targetCode = String(window.shiftCode || "").toUpperCase();

      return list.filter((item) => {
        const ownerCode = resolveSessionOwnerShiftCode(item, location);
        return ownerCode && ownerCode === targetCode;
      });
    }

    // Tanpa window: hari ini operator aktif saja, tanggal lampau semua session.
    if (isLoadedDateToday()) {
      return list.filter((item) =>
        ["ACTIVE", "OPEN"].includes(String(item.status || "").toUpperCase())
      );
    }

    return list;
  }

  // Note yang ditampilkan dibatasi window shift terpilih.
  function resolveNoteTextForShift(item, window) {
    if (!window) {
      return item.operatorNote || "";
    }

    const notes = Array.isArray(item.notes) ? item.notes : [];
    const filtered = notes.filter((note) => {
      const status = String(note.status || "").toUpperCase();
      const isActive = ["ACTIVE", "OPEN"].includes(status);

      return overlapsShiftWindow(
        note.createdAt,
        note.endTime,
        isActive,
        window
      );
    });

    return filtered.map((note) => note.text).join(" | ");
  }

  function buildOperatorDisplayRows(sessions, location) {
    const list = filterOperatorSessionsForShift(sessions, location);
    const window = getShiftWindowForLocation(location);
    const shiftTag = resolveOperatorShiftTag(location);

    return list.map((item) => ({
      label: item.operatorLabel,
      shiftTag: shiftTag.label,
      shiftTagCode: shiftTag.code,
      subText: item.operatorSubText || "",
      processName: item.processName || "",
      styleName: item.styleName || "",
      status: item.status || "",
      note: resolveNoteTextForShift(item, window),
    }));
  }

  // Map uuid -> daftar session hari itu (semua operator / shift).
  function buildActiveOperatorMap(data) {
    const map = new Map();

    extractRows(data)
      .map(normalizeOperatorSession)
      .filter((item) => item?.uuid && item?.operatorLabel)
      .forEach((item) => {
        const key = normalizeText(item.uuid);

        if (!map.has(key)) {
          map.set(key, []);
        }

        map.get(key).push(item);
      });

    map.forEach((sessions, key) => {
      sessions.sort((a, b) => {
        const timeA = parseDateTime(a.loginTime)?.getTime() || 0;
        const timeB = parseDateTime(b.loginTime)?.getTime() || 0;
        return timeA - timeB || Number(a.id || 0) - Number(b.id || 0);
      });
      map.set(key, sessions);
    });

    return map;
  }

  async function loadMachineSettings() {
    const data = await getMachineSettings();
    machineSettings.value = buildMachineSettingsMap(data);

    return machineSettings.value;
  }

  async function loadShiftConfigs() {
    try {
      const data = await getLineShiftConfig("");
      const lines = Array.isArray(data?.lines) ? data.lines : [];
      shiftConfigMap.value = buildLineShiftConfigMap(lines);
    } catch (err) {
      console.warn("Gagal load line shift config:", err);
      shiftConfigMap.value = new Map();
    }

    return shiftConfigMap.value;
  }

  async function loadActiveOperators(date) {
    const data = await getMachineOperatorReport(date);
    activeOperatorMap.value = buildActiveOperatorMap(data);

    return activeOperatorMap.value;
  }

  function normalizeRows(data) {
    const rows = extractRows(data);

    const normalized = rows.map((row) => {
      const uuid = String(getVal(row, "uuid", "UUID") || "").trim();
      const setting = machineSettings.value.get(normalizeText(uuid)) || null;
      const locationFromApi = String(getVal(row, "location", "Location") || "");
      const location = setting?.location || locationFromApi || "-";
      const operatorSessions =
        activeOperatorMap.value.get(normalizeText(uuid)) || [];
      const displaySessions = filterOperatorSessionsForShift(
        operatorSessions,
        location
      );
      const operator = pickPrimaryOperatorSession(displaySessions);
      const operatorDisplayRows = buildOperatorDisplayRows(
        operatorSessions,
        location
      );

      const backendName = String(
        getVal(
          row,
          "machineName",
          "MachineName",
          "nickName",
          "NickName",
          "name",
          "Name"
        ) || uuid
      ).trim();

      const originalMachineName = String(
        getVal(
          row,
          "originalMachineName",
          "OriginalMachineName",
          "originalNickName",
          "OriginalNickName",
          "original_name",
          "OriginalName"
        ) || backendName
      ).trim();

      const machineName =
        setting?.customName || backendName || originalMachineName || uuid;

      const operatorProcessName = operator?.processName || "";
      const displayMachineName = operatorProcessName || machineName;

      let runtime = toNumber(
        getVal(
          row,
          "runtimeSec",
          "RuntimeSec",
          "runtime",
          "Runtime",
          "runTime",
          "RunTime",
          "powerOnSeconds",
          "PowerOnSeconds",
          "power_on_seconds"
        ) || 0
      );

      const procTime = toNumber(
        getVal(
          row,
          "procSec",
          "ProcSec",
          "procTimeSec",
          "ProcTimeSec",
          "procTime",
          "ProcTime",
          "productive_seconds",
          "productiveSeconds",
          "process_time",
          "ProcessTime"
        ) || 0
      );

      const okProcTime = toNumber(
        getVal(
          row,
          "okProcSec",
          "OkProcSec",
          "ok_proc_sec",
          "okProcessTime",
          "OkProcessTime"
        ) || 0
      );

      if (procTime > runtime) {
        runtime = procTime;
      }

      const lossTime = Math.max(0, runtime - procTime);

      const productivity =
        runtime > 0 ? Math.min((procTime / runtime) * 100, 100) : 0;

      const output = toNumber(getVal(row, "output", "Output") || 0);
      const cycles = toNumber(getVal(row, "cycles", "Cycles") || 0);
      const complete = toNumber(getVal(row, "complete", "Complete") || 0);
      const incomplete = toNumber(getVal(row, "incomplete", "Incomplete") || 0);

      const avgCT = toNumber(
        getVal(row, "avgCycle", "AvgCycle", "avgCT", "AvgCT") || 0
      );

      const minCT = toNumber(
        getVal(row, "minCycle", "MinCycle", "minCT", "MinCT") || 0
      );

      const maxCT = toNumber(
        getVal(row, "maxCycle", "MaxCycle", "maxCT", "MaxCT") || 0
      );

      const slowCycles = toNumber(
        getVal(row, "slowCycles", "SlowCycles") || 0
      );

      const alarm = toNumber(
        getVal(row, "alarmCount", "AlarmCount", "alarm", "Alarm") || 0
      );

      return {
        date: String(getVal(row, "date", "Date") || ""),
        uuid,
        tableName: String(getVal(row, "tableName", "TableName") || ""),
        machineName,
        displayMachineName,
        originalMachineName,
        customName: setting?.customName || "",
        location,
        manualPic: setting?.pic || "",
        manualSpv: setting?.spv || "",
        ip: String(getVal(row, "ip", "IP") || ""),

        macType: String(getVal(row, "macType", "MacType") || ""),
        macState: normalizeMacState(row),

        runtime,
        procTime,
        okProcTime,
        lossTime,

        runtimeHours: Number((runtime / 3600).toFixed(2)),
        procHours: Number((procTime / 3600).toFixed(2)),
        lossTimeHours: Number((lossTime / 3600).toFixed(2)),

        productivity,
        productivityPct: productivity,
        status: getProductivityStatus(productivity),

        output,
        cycles,
        complete,
        incomplete,
        abnormal: incomplete,

        avgCT,
        minCT,
        maxCT,
        slowCycles,

        uniqueFiles: toNumber(
          getVal(row, "uniqueFiles", "UniqueFiles", "unique_files") || 0
        ),
        program: String(
          getVal(row, "topFile", "TopFile", "program", "Program") || ""
        ),
        firstProcess: String(
          getVal(row, "firstProcess", "FirstProcess", "first_process") || ""
        ),
        lastProcess: String(
          getVal(row, "lastProcess", "LastProcess", "last_process") || ""
        ),
        alarm,
        alarmTypes: String(
          getVal(row, "alarmTypes", "AlarmTypes", "alarm_types") || ""
        ),

        workSecondsPerDay: WORK_SECONDS_PER_DAY,

        pic: operator?.operatorLabel || "",
        spv: setting?.spv || "",

        operatorNik: operator?.operatorNik || "",
        operatorName: operator?.operatorName || "",
        operatorLabel: operator?.operatorLabel || "",
        operatorProcessName,
        operatorStyleName: operator?.styleName || "",
        operatorLoginTime: operator?.loginTime || "",
        operatorLogoutTime: operator?.logoutTime || "",
        operatorLoginClock: operator?.operatorLoginClock || "",
        operatorActiveDuration: operator?.operatorActiveDuration || "",
        operatorLoginText: operator?.operatorLoginText || "",
        operatorActiveText: operator?.operatorActiveText || "",
        operatorSubText: operator?.operatorSubText || "",
        operatorNote:
          displaySessions
            .map((item) => resolveNoteTextForShift(item, getShiftWindowForLocation(location)))
            .filter(Boolean)
            .join(" | ") || "",
        operatorNotes:
          displaySessions
            .map((item) => resolveNoteTextForShift(item, getShiftWindowForLocation(location)))
            .filter(Boolean)
            .join(" | ") || "",
        operatorSessions: displaySessions,
        operatorDisplayRows,
        operatorCount: operatorDisplayRows.length,
      };
    });

    machines.value = normalized;
    setLastUpdate();

    return normalized;
  }

  async function loadDashboard(date, options = {}) {
    const requestDate = String(date || "").trim();
    const requestShift = String(options.shift || "").trim();
    const requestKey = `${requestDate}|${requestShift}`;

    if (loading.value && inFlightRequest && inFlightDate === requestKey) {
      return inFlightRequest;
    }

    const requestSeq = ++dashboardRequestSeq;

    loading.value = true;
    errorMessage.value = "";
    inFlightDate = requestKey;
    loadedDashboardDate.value = requestDate;
    loadedDashboardShift.value = requestShift;

    const request = (async () => {
      const [settingsResult, operatorResult, productivityResult, shiftConfigResult] =
        await Promise.allSettled([
          getMachineSettings(),
          getMachineOperatorReport(requestDate),
          getProductivity(requestDate, { shift: requestShift }),
          getLineShiftConfig(""),
        ]);

      if (requestSeq !== dashboardRequestSeq) return;

      if (settingsResult.status === "fulfilled") {
        machineSettings.value = buildMachineSettingsMap(settingsResult.value);
      } else {
        console.warn("Gagal load machine settings:", settingsResult.reason);
      }

      if (shiftConfigResult.status === "fulfilled") {
        const lines = Array.isArray(shiftConfigResult.value?.lines)
          ? shiftConfigResult.value.lines
          : [];
        shiftConfigMap.value = buildLineShiftConfigMap(lines);
      } else {
        console.warn("Gagal load line shift config:", shiftConfigResult.reason);
      }

      if (operatorResult.status === "fulfilled") {
        activeOperatorMap.value = buildActiveOperatorMap(operatorResult.value);
      } else {
        console.warn("Gagal load active operators:", operatorResult.reason);
        activeOperatorMap.value = new Map();
      }

      if (productivityResult.status === "rejected") {
        throw productivityResult.reason;
      }

      normalizeRows(productivityResult.value);
    })();

    inFlightRequest = request;

    try {
      await request;
    } catch (err) {
      if (requestSeq === dashboardRequestSeq) {
        errorMessage.value = `Gagal mengambil data dari backend: ${
          err?.message || err
        }`;
      }
    } finally {
      if (requestSeq === dashboardRequestSeq) {
        loading.value = false;
      }

      if (inFlightRequest === request) {
        inFlightRequest = null;
        inFlightDate = "";
      }
    }

    return request;
  }

  return {
    machines,
    loading,
    errorMessage,
    lastUpdate,
    machineSettings,
    activeOperatorMap,
    shiftConfigMap,

    loadMachineSettings,
    loadShiftConfigs,
    loadActiveOperators,
    loadDashboard,
    normalizeRows,
  };
}