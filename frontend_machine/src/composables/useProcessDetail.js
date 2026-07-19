import { computed, ref } from "vue";
import { getProcessDetail } from "../api/machineApi";

const detailData = ref(null);
const detailLoading = ref(false);
const detailError = ref("");
const detailPage = ref(1);
const detailPageSize = 20;

const currentDetailUuid = ref("");

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function getPendingDetailUuid() {
  return localStorage.getItem("machineDashboardDetailUuid") || "";
}

function clearPendingDetailUuid() {
  localStorage.removeItem("machineDashboardDetailUuid");
  localStorage.removeItem("machineDashboardDetailMachineName");
}

/*
  Fungsi ini dibuat supaya saat masuk dari Location Template,
  UUID yang diklik akan diprioritaskan.

  Contoh:
  - User klik mesin di Location Template
  - App menyimpan UUID ke localStorage: machineDashboardDetailUuid
  - Saat halaman Detail Proses Mesin load, function ini akan pakai UUID tersebut
*/
function resolveDetailUuid(uuid) {
  const pendingUuid = getPendingDetailUuid();

  if (pendingUuid) {
    clearPendingDetailUuid();
    return pendingUuid;
  }

  return uuid || "";
}

async function loadProcessDetail(date, uuid) {
  const targetUuid = resolveDetailUuid(uuid);

  if (!targetUuid) return;

  detailLoading.value = true;
  detailError.value = "";
  currentDetailUuid.value = targetUuid;

  try {
    detailData.value = await getProcessDetail(date, targetUuid);
    detailPage.value = 1;
  } catch (err) {
    detailError.value = `Gagal mengambil detail proses: ${err.message}`;
  } finally {
    detailLoading.value = false;
  }
}

const detailMachine = computed(() => {
  return detailData.value?.machine || {};
});

function eventTimeValue(e) {
  const value = e.endTime || e.startTime || "";
  const safeValue = String(value).replace(" ", "T");
  const time = new Date(safeValue).getTime();

  return Number.isFinite(time) ? time : 0;
}

const detailEvents = computed(() => {
  const events = detailData.value?.events || [];

  return [...events].sort((a, b) => {
    return eventTimeValue(b) - eventTimeValue(a);
  });
});

const detailGroups = computed(() => {
  return detailData.value?.groups || [];
});

const detailHours = computed(() => {
  return detailData.value?.hours || [];
});

const detailAlarms = computed(() => {
  return detailData.value?.alarms || [];
});

const pagedDetailEvents = computed(() => {
  const start = (detailPage.value - 1) * detailPageSize;
  return detailEvents.value.slice(start, start + detailPageSize);
});

const totalDetailPages = computed(() => {
  return Math.max(1, Math.ceil(detailEvents.value.length / detailPageSize));
});

const avgNodeDistance = computed(() => {
  const events = detailEvents.value;
  if (!events.length) return "-";

  const total = events.reduce((sum, e) => {
    return sum + Number(e.nodeDistance || 0);
  }, 0);

  return Math.round(total / events.length);
});

const maxHourProc = computed(() => {
  return Math.max(1, ...detailHours.value.map((h) => Number(h.procSec || 0)));
});

const maxGroupProc = computed(() => {
  return Math.max(1, ...detailGroups.value.map((g) => Number(g.procSec || 0)));
});

function prevDetailPage() {
  if (detailPage.value > 1) {
    detailPage.value--;
  }
}

function nextDetailPage() {
  if (detailPage.value < totalDetailPages.value) {
    detailPage.value++;
  }
}

function goDetailPage(pageNumber) {
  const page = Number(pageNumber);

  if (!Number.isFinite(page)) return;

  if (page < 1) {
    detailPage.value = 1;
    return;
  }

  if (page > totalDetailPages.value) {
    detailPage.value = totalDetailPages.value;
    return;
  }

  detailPage.value = page;
}

function resetProcessDetail() {
  detailData.value = null;
  detailLoading.value = false;
  detailError.value = "";
  detailPage.value = 1;
  currentDetailUuid.value = "";
}

function isCurrentDetailUuid(uuid) {
  return normalizeText(currentDetailUuid.value) === normalizeText(uuid);
}

export function useProcessDetail() {
  return {
    detailData,
    detailLoading,
    detailError,
    detailPage,
    detailPageSize,
    currentDetailUuid,

    detailMachine,
    detailEvents,
    detailGroups,
    detailHours,
    detailAlarms,

    pagedDetailEvents,
    totalDetailPages,

    avgNodeDistance,
    maxHourProc,
    maxGroupProc,

    loadProcessDetail,
    prevDetailPage,
    nextDetailPage,
    goDetailPage,
    resetProcessDetail,

    getPendingDetailUuid,
    clearPendingDetailUuid,
    isCurrentDetailUuid,
  };
}