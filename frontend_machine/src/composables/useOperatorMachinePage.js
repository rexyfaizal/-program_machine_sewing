import { onBeforeUnmount, onMounted, ref, unref } from "vue";
import {
  getActiveMachineOperator,
  getMachineSettings,
  getProductivity,
  loginMachineOperator,
  saveMachineOperatorNote,
  searchEmployees,
  searchStyles,
  searchProcessesByStyle,
} from "../api/machineApi";

export function useOperatorMachinePage(uuidSource) {
  const loading = ref(false);
  const saving = ref(false);
  const noteSaving = ref(false);
  const errorMessage = ref("");
  const successMessage = ref("");

  const pageMode = ref("login");
  const forceReplace = ref(false);

  const operatorNik = ref("");
  const operatorName = ref("");
  const operatorBranch = ref("");
  const employeeOptions = ref([]);
  const employeeSearching = ref(false);
  const showEmployeeOptions = ref(false);

  const processName = ref("");
  const styleName = ref("");

  const styleOptions = ref([]);
  const styleSearching = ref(false);
  const showStyleOptions = ref(false);

  const processOptions = ref([]);
  const processSearching = ref(false);
  const showProcessOptions = ref(false);

  const activeSession = ref(null);
  const activeNotes = ref([]);
  const otherNote = ref("");

  let employeeSearchTimer = null;
  let styleSearchTimer = null;
  let processSearchTimer = null;

  const reasonMenus = [
    { reasonCode: "MACHINE_BROKEN", reasonName: "Mesin Rusak" },
    { reasonCode: "TOILET", reasonName: "Ke Toilet" },
    { reasonCode: "PRAYER", reasonName: "Solat" },
    { reasonCode: "WAIT_HANCA", reasonName: "Tunggu Hanca" },
    { reasonCode: "OTHER", reasonName: "Other" },
  ];

  const machine = ref({
    uuid: "",
    machineName: "",
    originalMachineName: "",
    customName: "",
    location: "",
    ip: "",
  });

  function todayLocal() {
    const d = new Date();
    const offset = d.getTimezoneOffset();
    const local = new Date(d.getTime() - offset * 60000);
    return local.toISOString().slice(0, 10);
  }

  function getUuid() {
    return String(unref(uuidSource) || "").trim();
  }

  function cacheKey(uuid) {
    return `machineOperatorActiveSession:${String(uuid || "").trim()}`;
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

  function unwrapResponse(data) {
    return (
      getVal(data, "data", "Data", "result", "Result", "payload", "Payload") ||
      data ||
      {}
    );
  }

  function normalizeSetting(row) {
    return {
      uuid: String(getVal(row, "uuid", "UUID") || ""),
      customName: String(getVal(row, "customName", "CustomName") || ""),
      location: String(getVal(row, "location", "Location") || ""),
    };
  }

  function normalizeProductivity(row) {
    const machineName = String(
      getVal(
        row,
        "machineName",
        "MachineName",
        "nickName",
        "NickName",
        "name",
        "Name"
      ) || ""
    );

    const originalMachineName = String(
      getVal(
        row,
        "originalMachineName",
        "OriginalMachineName",
        "originalNickName",
        "OriginalNickName"
      ) || machineName
    );

    return {
      uuid: String(getVal(row, "uuid", "UUID") || ""),
      machineName,
      originalMachineName,
      location: String(getVal(row, "location", "Location") || ""),
      ip: String(getVal(row, "ip", "IP", "lastLoginIP", "LastLoginIP") || ""),
    };
  }

  function normalizeEmployee(emp) {
    return {
      nik: String(getVal(emp, "nik", "NIK") || "").trim(),
      name: String(getVal(emp, "name", "Name") || "").trim(),
      branchdetail: String(
        getVal(emp, "branchdetail", "BranchDetail", "branchDetail") || ""
      ).trim(),
    };
  }

  function normalizeStyle(row) {
    return {
      styleName: String(
        getVal(row, "styleName", "StyleName", "style", "STYLE") || ""
      ).trim(),
    };
  }

  function normalizeProcess(row) {
    return {
      id: Number(getVal(row, "id", "ID") || 0),
      styleName: String(
        getVal(row, "styleName", "StyleName", "style", "STYLE") || ""
      ).trim(),
      processName: String(
        getVal(row, "processName", "ProcessName", "proses", "PROSES") || ""
      ).trim(),
    };
  }

  function normalizeSession(row) {
    if (!row) return null;

    return {
      id: Number(
        getVal(row, "id", "ID", "sessionId", "SessionID", "session_id") || 0
      ),
      sessionDate: String(
        getVal(row, "sessionDate", "SessionDate", "session_date") || ""
      ),
      uuid: String(getVal(row, "uuid", "UUID") || ""),
      machineName: String(
        getVal(row, "machineName", "MachineName", "machine_name") || ""
      ),
      location: String(getVal(row, "location", "Location") || ""),
      operatorNik: String(
        getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
      ),
      operatorName: String(
        getVal(row, "operatorName", "OperatorName", "operator_name") || ""
      ),
      branchdetail: String(getVal(row, "branchdetail", "BranchDetail") || ""),
      processName: String(
        getVal(row, "processName", "ProcessName", "process_name") || ""
      ),
      styleName: String(
        getVal(row, "styleName", "StyleName", "style_name") || ""
      ),
      shiftCode: String(
        getVal(row, "shiftCode", "ShiftCode", "shift_code") || ""
      ),
      shiftName: String(
        getVal(row, "shiftName", "ShiftName", "shift_name") || ""
      ),
      loginTime: String(
        getVal(row, "loginTime", "LoginTime", "login_time") || ""
      ),
      logoutTime: String(
        getVal(row, "logoutTime", "LogoutTime", "logout_time") || ""
      ),
      status: String(getVal(row, "status", "Status") || ""),
    };
  }

  function normalizeNote(row) {
    return {
      id: Number(getVal(row, "id", "ID") || 0),
      reasonCode: String(
        getVal(row, "reasonCode", "ReasonCode", "reason_code") || ""
      ),
      reasonName: String(
        getVal(row, "reasonName", "ReasonName", "reason_name") || ""
      ),
      note: String(getVal(row, "note", "Note") || ""),
      createdAt: String(
        getVal(row, "createdAt", "CreatedAt", "created_at") || ""
      ),
    };
  }

  function employeeLabel(emp) {
    const nik = emp?.nik || "";
    const name = emp?.name || "";
    const branch = emp?.branchdetail || "";

    return `${nik} - ${name}${branch ? ` (${branch})` : ""}`;
  }

  function styleLabel(item) {
    return item?.styleName || "";
  }

  function processLabel(item) {
    return item?.processName || "";
  }

  function formatDateTime(value) {
    if (!value) return "-";

    const d = new Date(value);

    if (Number.isNaN(d.getTime())) {
      return String(value).replace("T", " ").slice(0, 16);
    }

    return d.toLocaleString("id-ID", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function formatTime(value) {
    if (!value) return "-";

    const d = new Date(value);

    if (Number.isNaN(d.getTime())) {
      return String(value).replace("T", " ").slice(11, 16);
    }

    return d.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function getCurrentShift() {
    const now = new Date();
    const hour = now.getHours();

    if (hour >= 7 && hour < 15) {
      return { shiftCode: "SHIFT_1", shiftName: "Shift 1" };
    }

    if (hour >= 15 && hour < 23) {
      return { shiftCode: "SHIFT_2", shiftName: "Shift 2" };
    }

    return { shiftCode: "SHIFT_3", shiftName: "Shift 3" };
  }

  function saveSessionCache(session) {
    const uuid = String(session?.uuid || machine.value.uuid || getUuid()).trim();

    if (!uuid || !session?.id) return;

    const payload = {
      savedAt: new Date().toISOString(),
      cacheDate: todayLocal(),
      session,
      notes: activeNotes.value || [],
    };

    localStorage.setItem(cacheKey(uuid), JSON.stringify(payload));
  }

  function loadSessionCache(uuid) {
    try {
      const raw = localStorage.getItem(cacheKey(uuid));

      if (!raw) return null;

      const payload = JSON.parse(raw);
      const session = normalizeSession(payload?.session);
      const notes = extractRows(payload?.notes).map(normalizeNote);

      if (!session?.id) return null;

      if (payload?.cacheDate !== todayLocal()) {
        localStorage.removeItem(cacheKey(uuid));
        return null;
      }

      const sessionUuid = String(session.uuid || uuid || "").trim();

      if (normalizeText(sessionUuid) !== normalizeText(uuid)) {
        return null;
      }

      const status = String(session.status || "ACTIVE").toUpperCase();

      if (["CLOSED", "AUTO_LOGOUT", "LOGOUT"].includes(status)) {
        localStorage.removeItem(cacheKey(uuid));
        return null;
      }

      return {
        session,
        notes,
      };
    } catch {
      localStorage.removeItem(cacheKey(uuid));
      return null;
    }
  }

  function clearSessionCache(uuid) {
    if (!uuid) return;
    localStorage.removeItem(cacheKey(uuid));
  }

  function activateSession(session, notes = [], message = "") {
    activeSession.value = session;
    activeNotes.value = extractRows(notes).map(normalizeNote);

    fillLoginFromSession(session);

    pageMode.value = "menu";
    forceReplace.value = false;
    errorMessage.value = "";

    if (message) {
      successMessage.value = message;
    }

    saveSessionCache(session);
  }

  function resetLoginForm() {
    operatorNik.value = "";
    operatorName.value = "";
    operatorBranch.value = "";
    employeeOptions.value = [];
    showEmployeeOptions.value = false;

    styleName.value = "";
    processName.value = "";
    styleOptions.value = [];
    processOptions.value = [];
    showStyleOptions.value = false;
    showProcessOptions.value = false;
  }

  function fillLoginFromSession(session) {
    if (!session) return;

    operatorNik.value = session.operatorNik || "";
    operatorName.value = session.operatorName || "";
    operatorBranch.value = session.branchdetail || "";
    processName.value = session.processName || processName.value || "";
    styleName.value = session.styleName || styleName.value || "";
  }

  async function loadMachineData() {
    const uuid = getUuid();

    if (!uuid) {
      errorMessage.value = "UUID mesin tidak ditemukan dari QR Code.";
      return;
    }

    let setting = null;
    let prodRow = null;

    try {
      const settingData = await getMachineSettings();
      const settingRows = extractRows(settingData).map(normalizeSetting);

      setting =
        settingRows.find(
          (item) => normalizeText(item.uuid) === normalizeText(uuid)
        ) || null;
    } catch (err) {
      console.warn("Gagal membaca machine settings:", err);
    }

    try {
      const productivityData = await getProductivity(todayLocal());
      const prodRows = extractRows(productivityData).map(normalizeProductivity);

      prodRow =
        prodRows.find(
          (item) => normalizeText(item.uuid) === normalizeText(uuid)
        ) || null;
    } catch (err) {
      console.warn("Gagal membaca productivity:", err);
    }

    const machineName =
      setting?.customName ||
      prodRow?.machineName ||
      prodRow?.originalMachineName ||
      uuid;

    machine.value = {
      uuid,
      machineName,
      originalMachineName: prodRow?.originalMachineName || machineName,
      customName: setting?.customName || "",
      location: setting?.location || prodRow?.location || "",
      ip: prodRow?.ip || "",
    };
  }

  async function loadActiveOperator() {
    const uuid = getUuid();

    if (!uuid) return;

    try {
      const data = await getActiveMachineOperator(uuid);
      const root = unwrapResponse(data);

      const isActive = Boolean(
        getVal(root, "active", "Active") || getVal(data, "active", "Active")
      );

      const sessionRaw =
        getVal(root, "session", "Session") ||
        getVal(data, "session", "Session");

      const notesRaw =
        getVal(root, "lastNotes", "LastNotes", "notes", "Notes") ||
        getVal(data, "lastNotes", "LastNotes", "notes", "Notes") ||
        [];

      const session = normalizeSession(sessionRaw);
      const status = String(session?.status || "").toUpperCase();

      const hasActiveSession =
        isActive ||
        Boolean(session?.id) ||
        status === "ACTIVE" ||
        status === "OPEN";

      if (hasActiveSession && session) {
        activateSession(
          session,
          notesRaw,
          "Operator sudah aktif di mesin ini"
        );
        return;
      }

      const cached = loadSessionCache(uuid);

      if (cached?.session) {
        activateSession(
          cached.session,
          cached.notes,
          "Operator sudah aktif di mesin ini"
        );
        return;
      }

      activeSession.value = null;
      activeNotes.value = [];
      pageMode.value = "login";
      forceReplace.value = false;
      successMessage.value = "";
    } catch (err) {
      console.warn("Gagal cek operator aktif:", err);

      const cached = loadSessionCache(uuid);

      if (cached?.session) {
        activateSession(
          cached.session,
          cached.notes,
          "Operator sudah aktif di mesin ini"
        );
        return;
      }

      activeSession.value = null;
      activeNotes.value = [];
      pageMode.value = "login";
      forceReplace.value = false;
      successMessage.value = "";
    }
  }

  async function loadData() {
    loading.value = true;
    errorMessage.value = "";
    successMessage.value = "";

    try {
      await loadMachineData();
      await loadActiveOperator();
    } catch (err) {
      errorMessage.value = `Gagal membuka halaman operator: ${err.message}`;
    } finally {
      loading.value = false;
    }
  }

  function handleEmployeeInput() {
    successMessage.value = "";
    errorMessage.value = "";
    operatorName.value = "";
    operatorBranch.value = "";
    showEmployeeOptions.value = true;

    if (employeeSearchTimer) {
      clearTimeout(employeeSearchTimer);
    }

    employeeSearchTimer = setTimeout(() => {
      searchEmployeeSuggestion();
    }, 250);
  }

  async function searchEmployeeSuggestion() {
    const q = String(operatorNik.value || "").trim();

    employeeOptions.value = [];

    if (q.length < 1) {
      employeeSearching.value = false;
      return;
    }

    employeeSearching.value = true;

    try {
      const rows = await searchEmployees(q);

      employeeOptions.value = rows
        .map(normalizeEmployee)
        .filter((emp) => emp.nik || emp.name)
        .slice(0, 8);

      showEmployeeOptions.value = true;
    } catch (err) {
      errorMessage.value = `Gagal cari operator: ${err.message}`;
    } finally {
      employeeSearching.value = false;
    }
  }

  function selectEmployee(emp) {
    operatorNik.value = emp.nik;
    operatorName.value = emp.name;
    operatorBranch.value = emp.branchdetail;

    employeeOptions.value = [];
    showEmployeeOptions.value = false;
    errorMessage.value = "";
    successMessage.value = "";
  }

  function hideSuggestionDelay() {
    setTimeout(() => {
      showEmployeeOptions.value = false;
    }, 180);
  }

  function handleStyleInput() {
    successMessage.value = "";
    errorMessage.value = "";

    processName.value = "";
    processOptions.value = [];
    showStyleOptions.value = true;

    if (styleSearchTimer) {
      clearTimeout(styleSearchTimer);
    }

    styleSearchTimer = setTimeout(() => {
      searchStyleSuggestion();
    }, 250);
  }

  async function searchStyleSuggestion() {
    const q = String(styleName.value || "").trim();

    styleOptions.value = [];

    if (q.length < 1) {
      styleSearching.value = false;
      return;
    }

    styleSearching.value = true;

    try {
      const rows = await searchStyles(q);

      styleOptions.value = rows
        .map(normalizeStyle)
        .filter((item) => item.styleName)
        .slice(0, 10);

      showStyleOptions.value = true;
    } catch (err) {
      errorMessage.value = `Gagal cari style: ${err.message}`;
    } finally {
      styleSearching.value = false;
    }
  }

  function openStyleOptions() {
    if (styleOptions.value.length) {
      showStyleOptions.value = true;
      return;
    }

    if (String(styleName.value || "").trim()) {
      searchStyleSuggestion();
    }
  }

  function selectStyle(item) {
    styleName.value = item.styleName;
    styleOptions.value = [];
    showStyleOptions.value = false;

    processName.value = "";
    processOptions.value = [];
    showProcessOptions.value = true;

    setTimeout(() => {
      searchProcessSuggestion();
    }, 100);
  }

  function hideStyleSuggestionDelay() {
    setTimeout(() => {
      showStyleOptions.value = false;
    }, 180);
  }

  function handleProcessInput() {
    successMessage.value = "";
    errorMessage.value = "";
    showProcessOptions.value = true;

    if (processSearchTimer) {
      clearTimeout(processSearchTimer);
    }

    processSearchTimer = setTimeout(() => {
      searchProcessSuggestion();
    }, 250);
  }

  async function searchProcessSuggestion() {
    const selectedStyle = String(styleName.value || "").trim();
    const q = String(processName.value || "").trim();

    processOptions.value = [];

    if (!selectedStyle) {
      processSearching.value = false;
      return;
    }

    processSearching.value = true;

    try {
      const rows = await searchProcessesByStyle(selectedStyle, q);

      processOptions.value = rows
        .map(normalizeProcess)
        .filter((item) => item.processName)
        .slice(0, 20);

      showProcessOptions.value = true;
    } catch (err) {
      errorMessage.value = `Gagal cari proses: ${err.message}`;
    } finally {
      processSearching.value = false;
    }
  }

  function openProcessOptions() {
    if (!String(styleName.value || "").trim()) {
      errorMessage.value = "Pilih style terlebih dahulu.";
      return;
    }

    showProcessOptions.value = true;
    searchProcessSuggestion();
  }

  function selectProcess(item) {
    processName.value = item.processName;
    processOptions.value = [];
    showProcessOptions.value = false;
    errorMessage.value = "";
    successMessage.value = "";
  }

  function hideProcessSuggestionDelay() {
    setTimeout(() => {
      showProcessOptions.value = false;
    }, 180);
  }

  async function submitLogin() {
    const uuid = String(machine.value.uuid || "").trim();

    errorMessage.value = "";
    successMessage.value = "";

    if (!uuid) {
      errorMessage.value = "UUID mesin tidak valid.";
      return;
    }

    if (!operatorNik.value || !operatorName.value) {
      errorMessage.value =
        "Operator wajib dipilih dari suggestion berdasarkan NIK.";
      return;
    }

    if (!styleName.value.trim()) {
      errorMessage.value = "Style wajib diisi.";
      return;
    }

    if (!processName.value.trim()) {
      errorMessage.value = "Proses wajib diisi.";
      return;
    }

    const shift = getCurrentShift();

    saving.value = true;

    try {
      const data = await loginMachineOperator({
        uuid,
        machineName: machine.value.machineName || "",
        location: machine.value.location || "",
        operatorNik: operatorNik.value,
        operatorName: operatorName.value,
        branchdetail: operatorBranch.value,
        processName: processName.value,
        styleName: styleName.value,
        shiftCode: shift.shiftCode,
        shiftName: shift.shiftName,
        forceReplace: forceReplace.value,
      });

      const root = unwrapResponse(data);
      const mode = String(getVal(root, "mode", "Mode") || "");
      const sessionRaw =
        getVal(root, "session", "Session") ||
        getVal(data, "session", "Session");

      const currentSessionRaw =
        getVal(root, "currentSession", "CurrentSession") ||
        getVal(data, "currentSession", "CurrentSession");

      const message = String(
        getVal(root, "message", "Message") ||
          getVal(data, "message", "Message") ||
          ""
      );

      if (mode === "DIFFERENT_OPERATOR" || currentSessionRaw) {
        activeSession.value = normalizeSession(currentSessionRaw);
        pageMode.value = "conflict";
        errorMessage.value =
          message || "Mesin ini sedang digunakan operator lain.";
        return;
      }

      const session = normalizeSession(sessionRaw);

      if (!session?.id) {
        errorMessage.value =
          "Login berhasil, tapi response sessionId tidak terbaca dari backend.";
        return;
      }

      activeNotes.value = [];
      activateSession(session, [], message || "Operator berhasil login.");
    } catch (err) {
      errorMessage.value = `Gagal login operator: ${err.message}`;
    } finally {
      saving.value = false;
    }
  }

  function loginOperatorBaru() {
    clearSessionCache(machine.value.uuid || getUuid());

    forceReplace.value = true;
    pageMode.value = "login";
    successMessage.value = "";
    errorMessage.value = "";
    resetLoginForm();
  }

  function batalGantiOperator() {
    forceReplace.value = false;
    pageMode.value = activeSession.value ? "menu" : "login";
    errorMessage.value = "";
  }

  async function submitNote(reason) {
    const session = activeSession.value;
    const sessionId = Number(session?.id || 0);
    const uuid = String(machine.value.uuid || "").trim();

    errorMessage.value = "";
    successMessage.value = "";

    if (!sessionId) {
      errorMessage.value =
        "Session operator tidak ditemukan. Silakan login ulang.";
      pageMode.value = "login";
      return;
    }

    if (!uuid) {
      errorMessage.value = "UUID mesin tidak valid.";
      return;
    }

    const noteText = String(otherNote.value || "").trim();

    if (reason.reasonCode === "OTHER" && !noteText) {
      errorMessage.value = "Untuk Other, keterangan tambahan wajib diisi.";
      return;
    }

    noteSaving.value = true;

    try {
      const data = await saveMachineOperatorNote({
        sessionId,
        uuid,
        reasonCode: reason.reasonCode,
        reasonName: reason.reasonName,
        note: noteText,
      });

      const root = unwrapResponse(data);
      const noteRaw =
        getVal(root, "note", "Note") || getVal(data, "note", "Note");

      const savedNote = normalizeNote(noteRaw);

      if (savedNote.id || savedNote.reasonCode) {
        activeNotes.value = [savedNote, ...activeNotes.value].slice(0, 5);
      } else {
        activeNotes.value = [
          {
            id: Date.now(),
            reasonCode: reason.reasonCode,
            reasonName: reason.reasonName,
            note: noteText,
            createdAt: new Date().toISOString(),
          },
          ...activeNotes.value,
        ].slice(0, 5);
      }

      otherNote.value = "";
      successMessage.value = `${reason.reasonName} berhasil disimpan.`;

      if (activeSession.value) {
        saveSessionCache(activeSession.value);
      }
    } catch (err) {
      errorMessage.value = `Gagal simpan keterangan: ${err.message}`;
    } finally {
      noteSaving.value = false;
    }
  }

  onMounted(() => {
    loadData();
  });

  onBeforeUnmount(() => {
    if (employeeSearchTimer) clearTimeout(employeeSearchTimer);
    if (styleSearchTimer) clearTimeout(styleSearchTimer);
    if (processSearchTimer) clearTimeout(processSearchTimer);
  });

  return {
    loading,
    saving,
    noteSaving,
    errorMessage,
    successMessage,
    pageMode,
    forceReplace,

    operatorNik,
    operatorName,
    operatorBranch,
    employeeOptions,
    employeeSearching,
    showEmployeeOptions,

    processName,
    styleName,

    styleOptions,
    styleSearching,
    showStyleOptions,

    processOptions,
    processSearching,
    showProcessOptions,

    activeSession,
    activeNotes,
    otherNote,
    reasonMenus,
    machine,

    employeeLabel,
    styleLabel,
    processLabel,

    formatDateTime,
    formatTime,

    handleEmployeeInput,
    selectEmployee,
    hideSuggestionDelay,

    handleStyleInput,
    openStyleOptions,
    selectStyle,
    hideStyleSuggestionDelay,

    handleProcessInput,
    openProcessOptions,
    selectProcess,
    hideProcessSuggestionDelay,

    submitLogin,
    loginOperatorBaru,
    batalGantiOperator,
    submitNote,
    loadData,
  };
}