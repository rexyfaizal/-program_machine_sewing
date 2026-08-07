import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import {
  claimMechanicBrokenMachine,
  doneMechanicBrokenMachine,
  getMechanicBrokenMachines,
  identifyMechanic,
  registerMechanicRFID,
} from "../api/machineApi";

const RECENT_NIK_KEY = "mechanicDashboardRecentNik";
const RECENT_NIK_TTL_MS = 10 * 60 * 1000;

function loadRecentNik() {
  try {
    const raw = sessionStorage.getItem(RECENT_NIK_KEY);
    if (!raw) return "";
    const parsed = JSON.parse(raw);
    const nik = String(parsed?.nik || "").trim();
    const savedAt = Number(parsed?.savedAt || 0);
    if (!nik || !savedAt) return "";
    if (Date.now() - savedAt > RECENT_NIK_TTL_MS) {
      sessionStorage.removeItem(RECENT_NIK_KEY);
      return "";
    }
    return nik;
  } catch {
    return "";
  }
}

function saveRecentNik(nik) {
  const value = String(nik || "").trim();
  if (!value) {
    sessionStorage.removeItem(RECENT_NIK_KEY);
    return;
  }

  sessionStorage.setItem(
    RECENT_NIK_KEY,
    JSON.stringify({
      nik: value,
      savedAt: Date.now(),
    })
  );
}

function playAlertBeep() {
  try {
    const AudioCtx = window.AudioContext || window.webkitAudioContext;
    if (!AudioCtx) return;

    const ctx = new AudioCtx();
    const now = ctx.currentTime;

    [0, 0.22, 0.44].forEach((offset, index) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = "square";
      osc.frequency.value = index === 1 ? 880 : 660;

      gain.gain.setValueAtTime(0.0001, now + offset);
      gain.gain.exponentialRampToValueAtTime(0.18, now + offset + 0.02);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + offset + 0.16);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now + offset);
      osc.stop(now + offset + 0.18);
    });

    setTimeout(() => {
      ctx.close().catch(() => {});
    }, 900);
  } catch {
    // ignore
  }
}

function formatDurationHm(seconds) {
  const total = Math.max(0, Math.floor(Number(seconds || 0)));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);

  if (h > 0) return `${h}j ${m}m`;
  return `${m}m`;
}

function formatClock(value) {
  const text = String(value || "").trim();
  if (!text) return "-";

  const normalized = text.includes("T") ? text : text.replace(" ", "T");
  const d = new Date(normalized);

  if (Number.isNaN(d.getTime())) {
    return text.length >= 16 ? text.slice(11, 16).replace(":", ".") : text;
  }

  return d
    .toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
    })
    .replace(":", ".");
}

function mapTicketRows(rows) {
  return (Array.isArray(rows) ? rows : []).map((item) => {
    const waitSec = Number(item.waitMechanicSeconds || 0);
    const workSec = Number(item.mechanicWorkSeconds || 0);
    const lossSec = Number(item.operatorLossSeconds || item.durationSeconds || 0);

    return {
      id: Number(item.id || 0),
      uuid: String(item.uuid || "").trim(),
      machineName: String(item.machineName || item.uuid || "-").trim(),
      location: String(item.location || "-").trim(),
      operatorNik: String(item.operatorNik || "").trim(),
      operatorName: String(item.operatorName || "").trim(),
      note: String(item.note || "").trim(),
      startTime: String(item.startTime || "").trim(),
      endTime: String(item.endTime || "").trim(),
      durationSeconds: lossSec,
      ticketStatus: String(item.ticketStatus || "OPEN").toUpperCase(),
      claimedByNik: String(item.claimedByNik || "").trim(),
      claimedByName: String(item.claimedByName || "").trim(),
      claimedAt: String(item.claimedAt || "").trim(),
      mechanicDoneAt: String(item.mechanicDoneAt || "").trim(),
      mechanicDoneByNik: String(item.mechanicDoneByNik || "").trim(),
      mechanicDoneByName: String(item.mechanicDoneByName || "").trim(),
      waitMechanicSeconds: waitSec,
      mechanicWorkSeconds: workSec,
      operatorLossSeconds: lossSec,
      operatorStillActive: Boolean(item.operatorStillActive),
      closedByMechanic: Boolean(item.closedByMechanic),
      closedByOperator: Boolean(item.closedByOperator),
      startClock: formatClock(item.startTime),
      endClock: formatClock(item.endTime),
      claimedClock: formatClock(item.claimedAt),
      mechanicDoneClock: formatClock(item.mechanicDoneAt),
      durationText: formatDurationHm(lossSec),
      waitText: formatDurationHm(waitSec),
      workText: formatDurationHm(workSec),
      lossText: formatDurationHm(lossSec),
    };
  });
}

export function useMechanicDashboard() {
  const activeRows = ref([]);
  const historyRows = ref([]);
  const loading = ref(false);
  const actingId = ref(0);
  const errorMessage = ref("");
  const notice = ref("");
  const soundEnabled = ref(true);
  const doneToday = ref(0);

  const knownOpenIds = ref(new Set());
  const primed = ref(false);

  const nikModalOpen = ref(false);
  const nikModalAction = ref("");
  const nikModalTicket = ref(null);
  const nikModalInput = ref("");
  const nikModalError = ref("");
  const nikModalLoading = ref(false);
  const nikInputRef = ref(null);

  const registerOpen = ref(false);
  const registerNik = ref("");
  const registerRfid = ref("");
  const registerError = ref("");
  const registerLoading = ref(false);
  const registerRfidRef = ref(null);

  let pollTimer = null;
  let tickTimer = null;
  let noticeTimer = null;

  const openCount = computed(
    () => activeRows.value.filter((item) => item.ticketStatus === "OPEN").length
  );
  const busyCount = computed(
    () =>
      activeRows.value.filter((item) => item.ticketStatus === "IN_PROGRESS")
        .length
  );

  const openRows = computed(() =>
    activeRows.value.filter((item) => item.ticketStatus === "OPEN")
  );
  const busyRows = computed(() =>
    activeRows.value.filter((item) => item.ticketStatus === "IN_PROGRESS")
  );
  const doneRows = computed(() => historyRows.value);

  const nikModalTitle = computed(() => {
    if (nikModalAction.value === "done") return "Selesai Perbaikan";
    return "Ambil Tiket";
  });

  const nikModalHint = computed(() => {
    const ticket = nikModalTicket.value;
    if (!ticket) return "Ketik NIK atau tap kartu RFID.";

    if (nikModalAction.value === "done") {
      return `Selesai: ${ticket.machineName}. Ketik NIK / tap kartu mekanik yang Ambil.`;
    }

    return `Ambil: ${ticket.machineName}. Ketik NIK atau tap kartu RFID.`;
  });

  function showNotice(message, isError = false) {
    notice.value = message;
    errorMessage.value = isError ? message : "";

    if (noticeTimer) clearTimeout(noticeTimer);
    noticeTimer = setTimeout(() => {
      notice.value = "";
    }, 3500);
  }

  async function loadList({ silent = false } = {}) {
    if (!silent) loading.value = true;

    try {
      const [activeData, historyData] = await Promise.all([
        getMechanicBrokenMachines({ status: "ALL" }),
        getMechanicBrokenMachines({ status: "DONE" }),
      ]);

      const nextActive = mapTicketRows(activeData?.rows);
      const nextHistory = mapTicketRows(historyData?.rows);

      const nextOpenIds = new Set(
        nextActive
          .filter((item) => item.ticketStatus === "OPEN")
          .map((item) => item.id)
      );

      if (primed.value && soundEnabled.value) {
        let hasNew = false;
        nextOpenIds.forEach((id) => {
          if (!knownOpenIds.value.has(id)) hasNew = true;
        });
        if (hasNew) playAlertBeep();
      }

      knownOpenIds.value = nextOpenIds;
      primed.value = true;
      activeRows.value = nextActive;
      historyRows.value = nextHistory;
      doneToday.value = Number(
        historyData?.doneToday ?? activeData?.doneToday ?? nextHistory.length
      );

      if (!silent) errorMessage.value = "";
    } catch (err) {
      if (!silent) {
        errorMessage.value = err?.message || "Gagal memuat daftar mesin rusak";
      }
    } finally {
      if (!silent) loading.value = false;
    }
  }

  async function openNikModal(action, item) {
    nikModalAction.value = action;
    nikModalTicket.value = item;
    nikModalInput.value = loadRecentNik();
    nikModalError.value = "";
    nikModalOpen.value = true;

    await nextTick();
    nikInputRef.value?.focus?.();
    nikInputRef.value?.select?.();
  }

  function closeNikModal() {
    if (nikModalLoading.value) return;
    nikModalOpen.value = false;
    nikModalAction.value = "";
    nikModalTicket.value = null;
    nikModalError.value = "";
  }

  function requestClaim(item) {
    if (!item?.id || item.ticketStatus !== "OPEN") return;
    openNikModal("claim", item);
  }

  function requestDone(item) {
    if (!item?.id || item.ticketStatus !== "IN_PROGRESS") return;
    openNikModal("done", item);
  }

  async function confirmNikModal() {
    const code = String(nikModalInput.value || "").trim();
    const ticket = nikModalTicket.value;
    const action = nikModalAction.value;

    if (!code) {
      nikModalError.value = "NIK / kartu RFID wajib diisi";
      return;
    }
    if (!ticket?.id || !action) {
      nikModalError.value = "Tiket tidak valid";
      return;
    }

    nikModalLoading.value = true;
    nikModalError.value = "";
    actingId.value = ticket.id;

    try {
      const identity = await identifyMechanic(code);
      if (!identity?.isValid) {
        nikModalError.value = identity?.message || "NIK/kartu bukan mekanik";
        return;
      }

      if (action === "claim") {
        await claimMechanicBrokenMachine({
          id: ticket.id,
          mechanicNik: identity.nik,
          mechanicName: identity.name,
        });
        showNotice(`Diambil oleh ${identity.name}`);
      } else {
        await doneMechanicBrokenMachine({
          id: ticket.id,
          mechanicNik: identity.nik,
        });
        showNotice(`Selesai oleh ${identity.name}`);
      }

      saveRecentNik(identity.nik);
      nikModalOpen.value = false;
      nikModalAction.value = "";
      nikModalTicket.value = null;
      await loadList({ silent: true });
    } catch (err) {
      nikModalError.value = err?.message || "Gagal memproses tiket";
      await loadList({ silent: true });
    } finally {
      nikModalLoading.value = false;
      actingId.value = 0;
    }
  }

  async function openRegisterModal() {
    registerOpen.value = true;
    registerNik.value = "";
    registerRfid.value = "";
    registerError.value = "";
    await nextTick();
    // Fokus ke NIK dulu; setelah isi NIK user/tap fokus ke RFID.
  }

  function closeRegisterModal() {
    if (registerLoading.value) return;
    registerOpen.value = false;
    registerError.value = "";
  }

  async function focusRegisterRfid() {
    await nextTick();
    registerRfidRef.value?.focus?.();
  }

  async function submitRegisterRFID() {
    const nik = String(registerNik.value || "").trim();
    const rfidNo = String(registerRfid.value || "").trim();

    if (!nik) {
      registerError.value = "NIK wajib diisi";
      return;
    }
    if (!rfidNo) {
      registerError.value = "Tap / isi nomor kartu RFID";
      return;
    }

    registerLoading.value = true;
    registerError.value = "";

    try {
      const data = await registerMechanicRFID({ nik, rfidNo });
      showNotice(`Kartu terdaftar: ${data?.name || nik}`);
      registerOpen.value = false;
      registerNik.value = "";
      registerRfid.value = "";
    } catch (err) {
      registerError.value = err?.message || "Gagal daftar kartu";
      registerRfid.value = "";
      await focusRegisterRfid();
    } finally {
      registerLoading.value = false;
    }
  }

  function startPolling() {
    stopPolling();
    pollTimer = setInterval(() => {
      loadList({ silent: true });
    }, 8000);

    tickTimer = setInterval(() => {
      activeRows.value = activeRows.value.map((item) => {
        if (item.ticketStatus === "DONE") return item;
        const nextSec = Number(item.durationSeconds || 0) + 1;
        return {
          ...item,
          durationSeconds: nextSec,
          durationText: formatDurationHm(nextSec),
        };
      });
    }, 1000);
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
    if (tickTimer) {
      clearInterval(tickTimer);
      tickTimer = null;
    }
  }

  onMounted(async () => {
    localStorage.removeItem("mechanicDashboardIdentity");
    await loadList({ silent: false });
    startPolling();
  });

  onUnmounted(() => {
    stopPolling();
    if (noticeTimer) clearTimeout(noticeTimer);
  });

  return {
    openRows,
    busyRows,
    doneRows,
    loading,
    actingId,
    errorMessage,
    notice,
    soundEnabled,
    openCount,
    busyCount,
    doneToday,
    nikModalOpen,
    nikModalInput,
    nikModalError,
    nikModalLoading,
    nikModalTitle,
    nikModalHint,
    nikInputRef,
    registerOpen,
    registerNik,
    registerRfid,
    registerError,
    registerLoading,
    registerRfidRef,
    loadList,
    requestClaim,
    requestDone,
    confirmNikModal,
    closeNikModal,
    openRegisterModal,
    closeRegisterModal,
    focusRegisterRfid,
    submitRegisterRFID,
  };
}
