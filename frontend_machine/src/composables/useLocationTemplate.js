import { computed, onMounted, ref } from "vue";
import {
  getMachineOperatorReport,
  getMachineSettings,
  getProductivity,
  getLineShiftConfig,
  getShiftSettings,
  saveShiftSettings,
  saveMachineSetting,
} from "../api/machineApi";
import { getInitialAdminMode } from "../utils/adminMode";
import { WORK_SECONDS } from "../utils/format";
import { buildLineShiftConfigMap } from "../utils/gm3Shift";

export function useLocationTemplate() {
  const selectedFactory = ref("GM1");
  const selectedDate = ref(todayLocal());
  const isAdmin = ref(false);
  const loading = ref(false);
  const errorMessage = ref("");
  const notice = ref("");

  const machines = ref([]);
  const machineSettings = ref(new Map());
  const activeOperatorMap = ref(new Map());

  const modalOpen = ref(false);
  const modalMode = ref("add");
  const selectedLine = ref("");
  const editingOldUuid = ref("");

  const lineModalOpen = ref(false);
  const lineModalMode = ref("add");
  const oldLineName = ref("");
  const lineFormName = ref("");

  const shiftModalOpen = ref(false);
  const shiftModalSaving = ref(false);
  const shiftConfigs = ref([]);
  const shiftSettings = ref([]);
  const saturdayShiftSettings = ref([]);
  const shiftConfigMap = ref(new Map());
  const shiftDefaults = ref([]);
  const saturdayShiftDefaults = ref([]);

  const draggingLine = ref("");
  const dragOverLine = ref("");

  const form = ref({
    uuid: "",
    customName: "",
    factory: "GM1",
    line: "LINE 1",
  });

  const factoryOptions = [
    { key: "GM1", label: "GM1" },
    { key: "GM2", label: "GM2" },
    { key: "GM3", label: "GM3" },
  ];

  const defaultLineMap = {
    GM1: [
      "LINE 1",
      "LINE 2",
      "LINE 3",
      "LINE 4",
      "LINE 5",
      "LINE 6",
      "LINE 7",
      "LINE 8",
      "LINE 9",
      "LINE 10",
      "LINE 11",
      "LINE 12",
      "LINE 13",
      "LINE 14",
      "LINE 15",
      "LINE 16",
      "LINE 17",
      "LINE 18",
    ],
    GM2: [
      "LINE 1",
      "LINE 2",
      "LINE 3",
      "LINE 4",
      "LINE 5",
      "LINE 6",
      "LINE 7",
      "LINE 8",
      "LINE 9",
      "LINE 10",
      "LINE 11",
      "LINE 12",
      "LINE 13",
      "LINE 14",
      "LINE 15",
      "LINE 16",
      "LINE 17",
      "LINE 18",
    ],
    GM3: [
      "LINE 1",
      "LINE 2",
      "LINE 3",
      "LINE 4",
      "LINE 5",
      "LINE 6",
    ],
  };

  const lineLayout = ref({
    GM1: [...defaultLineMap.GM1],
    GM2: [...defaultLineMap.GM2],
    GM3: [...defaultLineMap.GM3],
  });

  const activeLines = computed(() => {
    return lineLayout.value[selectedFactory.value] || [];
  });

  const assignedCount = computed(() => {
    return machines.value.filter((m) =>
      isMachineInFactory(m, selectedFactory.value)
    ).length;
  });

  const selectedFormMachine = computed(() => {
    return (
      machines.value.find((m) => normalizeText(m.uuid) === normalizeText(form.value.uuid)) ||
      null
    );
  });

  function todayLocal() {
    const d = new Date();
    const offset = d.getTimezoneOffset();
    const local = new Date(d.getTime() - offset * 60000);
    return local.toISOString().slice(0, 10);
  }

  function showNotice(message) {
    notice.value = message;

    setTimeout(() => {
      notice.value = "";
    }, 3000);
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

    return "";
  }

  function normalizeText(value) {
    return String(value || "")
      .trim()
      .toUpperCase()
      .replace(/\s+/g, " ");
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

  function getRowProcTime(row) {
    return toNumber(
      getVal(
        row,
        "procSec",
        "ProcSec",
        "procTime",
        "ProcTime",
        "procTimeSec",
        "ProcTimeSec",
        "process_time",
        "ProcessTime"
      ) || 0
    );
  }

  function getRowRuntime(row) {
    return toNumber(
      getVal(
        row,
        "runtimeSec",
        "RuntimeSec",
        "runtime",
        "Runtime",
        "RunTime",
        "runTime"
      ) || 0
    );
  }

  function getProductivityValue(row, procTime) {
    const fromApi = toNumber(
      getVal(
        row,
        "productivity",
        "Productivity",
        "productivityPct",
        "ProductivityPct",
        "productivityRaw",
        "ProductivityRaw",
        "Produktivitas",
        "productivity_percent"
      ) || 0
    );

    if (fromApi > 0) {
      return fromApi;
    }

    return Math.min((toNumber(procTime) / WORK_SECONDS) * 100, 100);
  }

  function lineLayoutUuid(factory) {
    return `__LOCATION_LINE_LAYOUT_${factory}__`;
  }

  function makeLocation(factory, line) {
    return `${factory} ${line}`;
  }

  function isSameLocation(location, factory, line) {
    const loc = normalizeText(location);
    const target = normalizeText(makeLocation(factory, line));
    const targetDash = normalizeText(`${factory} - ${line}`);

    return loc === target || loc === targetDash;
  }

  function isMachineInFactory(machine, factory) {
    const loc = normalizeText(machine.location);
    const f = normalizeText(factory);

    return loc.startsWith(f + " ");
  }

  function normalizeSetting(item) {
    return {
      uuid: String(getVal(item, "uuid", "UUID") || ""),
      customName: String(getVal(item, "customName", "CustomName") || ""),
      location: String(getVal(item, "location", "Location") || ""),
      pic: String(getVal(item, "pic", "PIC") || ""),
      spv: String(getVal(item, "spv", "SPV") || ""),
    };
  }

  function normalizeMachine(row) {
    const uuid = String(getVal(row, "uuid", "UUID") || "");

    const originalName = String(
      getVal(
        row,
        "originalNickName",
        "OriginalNickName",
        "originalMachineName",
        "OriginalMachineName",
        "nickName",
        "NickName"
      ) || uuid
    );

    const backendName = String(
      getVal(row, "nickName", "NickName", "machineName", "MachineName") ||
        originalName ||
        uuid
    );

    const setting = machineSettings.value.get(normalizeText(uuid));
    const operator = activeOperatorMap.value.get(normalizeText(uuid));

    const customName = String(setting?.customName || "").trim();
    const baseMachineName = customName || backendName;
    const operatorProcessName = String(operator?.processName || "").trim();
    const operatorStyleName = String(operator?.styleName || "").trim();
    const operatorName = String(operator?.operatorName || "").trim();
    const operatorNik = String(operator?.operatorNik || "").trim();
    const operatorLoggedIn = Boolean(operator);

    // Sama dashboard: nama tampil = proses operator (jika login), fallback custom/backend.
    const machineName = operatorProcessName || baseMachineName;

    const output = toNumber(getVal(row, "output", "Output") || 0);
    const procTime = getRowProcTime(row);
    const runtime = getRowRuntime(row);
    const productivity = getProductivityValue(row, procTime);
    const macState = getRowMacState(row);

    return {
      uuid,
      tableName: String(getVal(row, "tableName", "TableName") || ""),
      originalMachineName: originalName,
      machineName,
      customName,
      operatorProcessName,
      operatorStyleName,
      operatorName,
      operatorNik,
      operatorLoggedIn,
      usingProcessName: Boolean(operatorProcessName),

      ip: String(getVal(row, "ip", "IP", "lastLoginIP", "LastLoginIP") || "-"),

      location:
        setting?.location ||
        String(getVal(row, "location", "Location") || "-"),

      // Status mesin mengikuti MacState dari backend:
      // 2 = Working, 1 = Online, 0 = Offline
      macState,
      MacState: macState,
      mac_state: macState,

      output,
      procTime,
      runtime,
      productivity,

      mainSource: String(getVal(row, "mainSource", "MainSource") || "process_time"),

      status: String(getVal(row, "status", "Status") || "").toUpperCase(),
    };
  }

  function buildActiveOperatorMap(reportData) {
    const map = new Map();
    const rows = Array.isArray(reportData)
      ? reportData
      : reportData?.rows ||
        reportData?.Rows ||
        reportData?.data ||
        reportData?.items ||
        [];

    rows.forEach((row) => {
      const uuid = String(getVal(row, "uuid", "UUID") || "").trim();
      if (!uuid) return;

      const status = String(getVal(row, "status", "Status") || "")
        .trim()
        .toUpperCase();

      if (status && !["ACTIVE", "OPEN"].includes(status)) {
        return;
      }

      map.set(normalizeText(uuid), {
        processName: String(
          getVal(row, "processName", "ProcessName", "process_name") || ""
        ).trim(),
        styleName: String(
          getVal(row, "styleName", "StyleName", "style_name") || ""
        ).trim(),
        operatorName: String(
          getVal(row, "operatorName", "OperatorName", "operator_name") || ""
        ).trim(),
        operatorNik: String(
          getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
        ).trim(),
      });
    });

    return map;
  }

  function parseLineLayoutPayload(setting) {
    const candidates = [
      String(setting?.spv || "").trim(),
      String(setting?.customName || "").trim(),
    ];

    for (const raw of candidates) {
      if (!raw || !raw.startsWith("[")) continue;

      try {
        const parsed = JSON.parse(raw);

        if (Array.isArray(parsed) && parsed.length) {
          return parsed.map((x) => String(x).trim()).filter(Boolean);
        }
      } catch {
        // coba kandidat berikutnya
      }
    }

    return null;
  }

  function applyLineLayoutFromSettings(settingsList) {
    const nextLayout = {};

    for (const factory of Object.keys(defaultLineMap)) {
      nextLayout[factory] = [...defaultLineMap[factory]];

      const key = lineLayoutUuid(factory);
      const setting = settingsList.find(
        (x) => normalizeText(x.uuid) === normalizeText(key)
      );

      if (!setting) continue;

      const parsed = parseLineLayoutPayload(setting);

      if (parsed?.length) {
        nextLayout[factory] = parsed;
      }
    }

    lineLayout.value = nextLayout;
  }

  async function saveLineLayout(factory) {
    const lines = lineLayout.value[factory] || [];
    const payload = JSON.stringify(lines);

    // Simpan JSON layout di spv (panjang) + custom_name sebagai cadangan.
    await saveMachineSetting({
      uuid: lineLayoutUuid(factory),
      customName: payload,
      location: "LINE_LAYOUT",
      pic: "",
      spv: payload,
    });
  }

  async function loadMachineSettings() {
    const list = await getMachineSettings();
    const map = new Map();

    const normalized = list.map(normalizeSetting);

    normalized.forEach((setting) => {
      if (setting.uuid) {
        map.set(normalizeText(setting.uuid), setting);
      }
    });

    machineSettings.value = map;
    applyLineLayoutFromSettings(normalized);
  }

  async function loadShiftConfigs(factory = selectedFactory.value) {
    const factoryKey = String(factory || "").trim().toUpperCase();

    try {
      // Map global untuk cek line enabled (dashboard filter).
      const allData = await getLineShiftConfig("");
      const allLines = Array.isArray(allData?.lines) ? allData.lines : [];
      shiftConfigMap.value = buildLineShiftConfigMap(allLines);
      shiftDefaults.value = Array.isArray(allData?.defaults?.schedule)
        ? allData.defaults.schedule
        : [];
    } catch (err) {
      console.warn("Gagal load line shift config:", err);
      shiftConfigMap.value = new Map();
    }

    if (!factoryKey) {
      shiftConfigs.value = [];
      shiftSettings.value = [];
      saturdayShiftSettings.value = [];
      return;
    }

    try {
      // Jadwal area + toggle line dari shift_setting (sumber hitung FINAL).
      const data = await getShiftSettings(factoryKey);
      shiftSettings.value = Array.isArray(data?.shifts) ? data.shifts : [];
      saturdayShiftSettings.value = Array.isArray(data?.saturdayShifts)
        ? data.saturdayShifts
        : [];
      shiftConfigs.value = Array.isArray(data?.lines) ? data.lines : [];
      if (Array.isArray(data?.defaults?.schedule) && data.defaults.schedule.length) {
        shiftDefaults.value = data.defaults.schedule;
      }
      if (
        Array.isArray(data?.defaults?.saturdaySchedule) &&
        data.defaults.saturdaySchedule.length
      ) {
        saturdayShiftDefaults.value = data.defaults.saturdaySchedule;
      }
    } catch (err) {
      console.warn("Gagal load shift settings:", err);
      shiftSettings.value = [];
      saturdayShiftSettings.value = [];
      shiftConfigs.value = [];
      throw err;
    }
  }

  async function loadData() {
    loading.value = true;
    errorMessage.value = "";

    try {
      // Settings wajib; shift config opsional (jangan blokir jika endpoint belum siap).
      const [settingsResult, shiftResult] = await Promise.allSettled([
        loadMachineSettings(),
        loadShiftConfigs(selectedFactory.value),
      ]);

      if (settingsResult.status === "rejected") {
        throw settingsResult.reason;
      }

      if (shiftResult.status === "rejected") {
        console.warn("Gagal load shift settings saat loadData:", shiftResult.reason);
      }

      const useShift =
        selectedFactory.value === "GM3" ||
        [...shiftConfigMap.value.values()].some(
          (cfg) =>
            String(cfg.factory || "").toUpperCase() ===
              String(selectedFactory.value || "").toUpperCase() &&
            cfg.enabled
        );

      const [productivityResult, operatorResult] = await Promise.allSettled([
        getProductivity(
          selectedDate.value,
          useShift ? { shift: "CURRENT" } : {}
        ),
        getMachineOperatorReport(selectedDate.value),
      ]);

      if (operatorResult.status === "fulfilled") {
        activeOperatorMap.value = buildActiveOperatorMap(operatorResult.value);
      } else {
        console.warn("Gagal load operator report:", operatorResult.reason);
        activeOperatorMap.value = new Map();
      }

      if (productivityResult.status === "rejected") {
        throw productivityResult.reason;
      }

      const json = productivityResult.value;
      const rows = Array.isArray(json)
        ? json
        : json.rows || json.Rows || json.data || json.items || [];

      machines.value = rows
        .map(normalizeMachine)
        .filter((m) => m.uuid)
        .sort((a, b) =>
          String(a.machineName).localeCompare(String(b.machineName))
        );
    } catch (err) {
      const msg = String(err?.message || err || "");
      if (/failed to fetch/i.test(msg)) {
        errorMessage.value =
          "Gagal mengambil data: backend tidak aktif atau timeout. Pastikan server :5000 jalan, lalu Refresh.";
      } else {
        errorMessage.value = `Gagal mengambil data: ${msg}`;
      }
    } finally {
      loading.value = false;
    }
  }

  async function openShiftConfigModal() {
    if (!isAdmin.value) {
      showNotice("Akses atur shift hanya untuk admin.");
      return;
    }
    try {
      await loadShiftConfigs(selectedFactory.value);
      shiftModalOpen.value = true;
    } catch (err) {
      const msg = String(err?.message || err || "");
      if (/404|not found/i.test(msg)) {
        showNotice(
          "API /api/shift-settings belum aktif. Restart backend (go run .) lalu coba lagi."
        );
      } else {
        showNotice(`Gagal load shift: ${msg}`);
      }
    }
  }

  function closeShiftConfigModal() {
    shiftModalOpen.value = false;
  }

  async function saveShiftConfig(payload) {
    if (!isAdmin.value) {
      showNotice("Akses simpan shift hanya untuk admin.");
      return;
    }

    shiftModalSaving.value = true;
    try {
      await saveShiftSettings(payload);
      await loadShiftConfigs(selectedFactory.value);
      shiftModalOpen.value = false;
      showNotice("Konfigurasi shift berhasil disimpan ke database.");
      await loadData();
    } catch (err) {
      const msg = String(err?.message || err || "");
      if (/404|not found/i.test(msg)) {
        showNotice(
          "Gagal simpan: API /api/shift-settings belum aktif. Restart backend (go run .) lalu coba lagi."
        );
      } else {
        showNotice(`Gagal simpan shift: ${msg}`);
      }
    } finally {
      shiftModalSaving.value = false;
    }
  }

  function machinesByLine(line) {
    return machines.value.filter((m) =>
      isSameLocation(m.location, selectedFactory.value, line)
    );
  }

  function selectFactory(factory) {
    selectedFactory.value = factory;
  }

  function nextLineName(factory) {
    const lines = lineLayout.value[factory] || [];
    let n = lines.length + 1;
    let name = `LINE ${n}`;

    while (lines.some((x) => normalizeText(x) === normalizeText(name))) {
      n++;
      name = `LINE ${n}`;
    }

    return name;
  }

  function openAddLineModal() {
    if (!isAdmin.value) return;

    lineModalMode.value = "add";
    oldLineName.value = "";
    lineFormName.value = nextLineName(selectedFactory.value);
    lineModalOpen.value = true;
  }

  function openRenameLineModal(line) {
    if (!isAdmin.value) return;

    lineModalMode.value = "rename";
    oldLineName.value = line;
    lineFormName.value = line;
    lineModalOpen.value = true;
  }

  function closeLineModal() {
    lineModalOpen.value = false;
    oldLineName.value = "";
    lineFormName.value = "";
  }

  async function saveLine() {
    if (!isAdmin.value) {
      showNotice("Akses line hanya untuk admin.");
      return;
    }

    const factory = selectedFactory.value;
    const name = lineFormName.value.trim();

    if (!name) {
      showNotice("Nama line tidak boleh kosong.");
      return;
    }

    const currentLines = [...(lineLayout.value[factory] || [])];

    const duplicate = currentLines.some((line) => {
      if (
        lineModalMode.value === "rename" &&
        normalizeText(line) === normalizeText(oldLineName.value)
      ) {
        return false;
      }

      return normalizeText(line) === normalizeText(name);
    });

    if (duplicate) {
      showNotice("Nama line sudah ada.");
      return;
    }

    try {
      if (lineModalMode.value === "add") {
        lineLayout.value[factory] = [...currentLines, name];
        await saveLineLayout(factory);
        showNotice("Line baru berhasil ditambahkan.");
      }

      if (lineModalMode.value === "rename") {
        const oldName = oldLineName.value;

        lineLayout.value[factory] = currentLines.map((line) =>
          normalizeText(line) === normalizeText(oldName) ? name : line
        );

        // Simpan layout dulu supaya gagal lebih awal sebelum ubah mesin.
        await saveLineLayout(factory);

        const affectedMachines = machines.value.filter((m) =>
          isSameLocation(m.location, factory, oldName)
        );

        for (const machine of affectedMachines) {
          await saveMachineSetting({
            uuid: machine.uuid,
            customName: machine.customName || "",
            location: makeLocation(factory, name),
          });
        }

        showNotice("Nama line berhasil diubah.");
      }

      closeLineModal();
      await loadData();
    } catch (err) {
      showNotice(`Gagal simpan line: ${err.message}`);
    }
  }

  async function deleteLine(line) {
    if (!isAdmin.value) return;

    const factory = selectedFactory.value;
    const usedMachines = machines.value.filter((m) =>
      isSameLocation(m.location, factory, line)
    );

    let message = `Hapus ${line}?`;

    if (usedMachines.length) {
      message += `\n\nAda ${usedMachines.length} mesin di line ini. Mesin akan dikosongkan location-nya.`;
    }

    const ok = confirm(message);
    if (!ok) return;

    try {
      for (const machine of usedMachines) {
        await saveMachineSetting({
          uuid: machine.uuid,
          customName: machine.customName || "",
          location: "",
        });
      }

      lineLayout.value[factory] = (lineLayout.value[factory] || []).filter(
        (x) => normalizeText(x) !== normalizeText(line)
      );

      await saveLineLayout(factory);
      showNotice("Line berhasil dihapus.");
      await loadData();
    } catch (err) {
      showNotice(`Gagal hapus line: ${err.message}`);
    }
  }

  function onLineDragStart(event, line) {
    if (!isAdmin.value) return;

    draggingLine.value = line;
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", line);
  }

  function onLineDragEnter(line) {
    if (!isAdmin.value) return;
    if (!draggingLine.value) return;
    if (draggingLine.value === line) return;

    dragOverLine.value = line;
  }

  function onLineDragOver(event) {
    if (!isAdmin.value) return;

    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }

  async function onLineDrop(targetLine) {
    if (!isAdmin.value) return;

    const sourceLine = draggingLine.value;

    if (!sourceLine || sourceLine === targetLine) {
      onLineDragEnd();
      return;
    }

    const factory = selectedFactory.value;
    const lines = [...(lineLayout.value[factory] || [])];

    const sourceIndex = lines.findIndex(
      (line) => normalizeText(line) === normalizeText(sourceLine)
    );

    const targetIndex = lines.findIndex(
      (line) => normalizeText(line) === normalizeText(targetLine)
    );

    if (sourceIndex < 0 || targetIndex < 0) {
      onLineDragEnd();
      return;
    }

    const [movedLine] = lines.splice(sourceIndex, 1);

    let insertIndex = targetIndex;

    if (sourceIndex < targetIndex) {
      insertIndex = targetIndex - 1;
    }

    lines.splice(insertIndex, 0, movedLine);
    lineLayout.value[factory] = lines;

    try {
      await saveLineLayout(factory);
      showNotice("Urutan line berhasil disimpan.");
    } catch (err) {
      showNotice(`Gagal simpan urutan line: ${err.message}`);
      await loadData();
    } finally {
      onLineDragEnd();
    }
  }

  function onLineDragEnd() {
    draggingLine.value = "";
    dragOverLine.value = "";
  }

  function openAddModal(line) {
    if (!isAdmin.value) return;

    modalMode.value = "add";
    selectedLine.value = line;
    editingOldUuid.value = "";

    form.value = {
      uuid: "",
      customName: "",
      factory: selectedFactory.value,
      line,
    };

    modalOpen.value = true;
  }

  function openEditModal(machine) {
    if (!isAdmin.value) return;

    modalMode.value = "edit";
    selectedLine.value = getMachineLine(machine);
    editingOldUuid.value = machine.uuid;

    form.value = {
      uuid: machine.uuid,
      customName: machine.customName || "",
      factory: selectedFactory.value,
      line: getMachineLine(machine),
    };

    modalOpen.value = true;
  }

  function closeModal() {
    modalOpen.value = false;
    editingOldUuid.value = "";
    selectedLine.value = "";
  }

  function getMachineLine(machine) {
    for (const line of activeLines.value) {
      if (isSameLocation(machine.location, selectedFactory.value, line)) {
        return line;
      }
    }

    return selectedLine.value || activeLines.value[0] || "LINE 1";
  }

  function fillNameFromSelectedMachine() {
    const machine = selectedFormMachine.value;

    if (!machine) return;

    form.value.customName = machine.customName || "";
  }

  async function saveLocation() {
    if (!isAdmin.value) {
      showNotice("Akses tambah/edit hanya untuk admin.");
      return;
    }

    if (!form.value.uuid) {
      showNotice("Pilih mesin terlebih dahulu.");
      return;
    }

    const selected = machines.value.find((m) => {
      return normalizeText(m.uuid) === normalizeText(form.value.uuid);
    });

    if (!selected) {
      showNotice("Data mesin tidak ditemukan.");
      return;
    }

    const newLocation = makeLocation(form.value.factory, form.value.line);
    const customName = form.value.customName.trim();

    try {
      if (
        modalMode.value === "edit" &&
        editingOldUuid.value &&
        normalizeText(editingOldUuid.value) !== normalizeText(form.value.uuid)
      ) {
        const oldMachine = machines.value.find((m) => {
          return normalizeText(m.uuid) === normalizeText(editingOldUuid.value);
        });

        if (oldMachine) {
          await saveMachineSetting({
            uuid: oldMachine.uuid,
            customName: oldMachine.customName || "",
            location: "",
          });
        }
      }

      await saveMachineSetting({
        uuid: form.value.uuid,
        customName,
        location: newLocation,
      });

      showNotice("Location mesin berhasil disimpan.");
      closeModal();
      await loadData();
    } catch (err) {
      showNotice(`Gagal simpan: ${err.message}`);
    }
  }

  async function removeMachineFromLine(machine) {
    if (!isAdmin.value) return;

    const ok = confirm(`Hapus ${machine.machineName} dari location ini?`);
    if (!ok) return;

    try {
      await saveMachineSetting({
        uuid: machine.uuid,
        customName: machine.customName || "",
        location: "",
      });

      showNotice("Mesin berhasil dihapus dari line.");
      await loadData();
    } catch (err) {
      showNotice(`Gagal hapus: ${err.message}`);
    }
  }

  onMounted(async () => {
    isAdmin.value = await Promise.resolve(getInitialAdminMode());
    await loadData();
  });

  return {
    selectedFactory,
    selectedDate,
    isAdmin,
    loading,
    errorMessage,
    notice,

    machines,
    modalOpen,
    modalMode,
    selectedLine,
    editingOldUuid,

    lineModalOpen,
    lineModalMode,
    oldLineName,
    lineFormName,

    shiftModalOpen,
    shiftModalSaving,
    shiftConfigs,
    shiftSettings,
    saturdayShiftSettings,
    shiftDefaults,
    saturdayShiftDefaults,

    draggingLine,
    dragOverLine,

    form,
    factoryOptions,
    lineLayout,
    activeLines,
    assignedCount,

    makeLocation,
    loadData,
    machinesByLine,
    selectFactory,

    openAddLineModal,
    openRenameLineModal,
    closeLineModal,
    saveLine,
    deleteLine,

    openShiftConfigModal,
    closeShiftConfigModal,
    saveShiftConfig,

    onLineDragStart,
    onLineDragEnter,
    onLineDragOver,
    onLineDrop,
    onLineDragEnd,

    openAddModal,
    openEditModal,
    closeModal,
    fillNameFromSelectedMachine,
    saveLocation,
    removeMachineFromLine,
  };
}