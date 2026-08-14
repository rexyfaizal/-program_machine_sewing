import { computed, ref } from "vue";
import { getProcessDetail, getMachineOperatorReport } from "../api/machineApi";
import { formatDurationHHMMSS } from "../utils/format";
import {
  extractRows,
  formatExportTime,
  getVal,
  isAutoLogoutText,
  parseExportDateTime,
  resolveExportNoteDurationSeconds,
  toNumber,
} from "../utils/dashboardExportExcel";

const detailData = ref(null);
const detailLoading = ref(false);
const detailError = ref("");
const detailPage = ref(1);
const detailPageSize = 10;

const currentDetailUuid = ref("");
/** Note losstime operator untuk mesin yang sedang dibuka (sudah dinormalisasi). */
const operatorLossNotes = ref([]);
/** Session operator mesin yang sedang dibuka (untuk nama di export). */
const operatorSessions = ref([]);

function normalizeText(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/\s+/g, " ");
}

function noteReasonLabel(row) {
  return String(
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
}

function collectLossNotesForUuid(reportData, uuid) {
  const target = normalizeText(uuid);
  if (!target) return [];

  const sessions = extractRows(reportData);
  const notes = [];

  for (const session of sessions) {
    const sessionUuid = normalizeText(getVal(session, "uuid", "UUID"));
    if (!sessionUuid || sessionUuid !== target) continue;

    const rawNotes = extractRows(
      getVal(session, "notes", "Notes", "lastNotes", "LastNotes") || []
    );

    const activeLossReasonLabel = String(
      getVal(
        session,
        "activeLossReasonLabel",
        "ActiveLossReasonLabel",
        "active_loss_reason_label"
      ) || ""
    ).trim();

    const list =
      rawNotes.length > 0
        ? rawNotes
        : activeLossReasonLabel
          ? [
              {
                reasonCode: getVal(
                  session,
                  "activeLossReasonCode",
                  "ActiveLossReasonCode"
                ),
                reasonName: activeLossReasonLabel,
                createdAt: getVal(
                  session,
                  "activeLossStartTime",
                  "ActiveLossStartTime"
                ),
                durationSeconds: getVal(
                  session,
                  "activeLossDurationSeconds",
                  "ActiveLossDurationSeconds"
                ),
                status:
                  getVal(session, "activeLossStatus", "ActiveLossStatus") ||
                  "ACTIVE",
              },
            ]
          : [];

    for (const row of list) {
      const reasonName = noteReasonLabel(row);
      const note = String(getVal(row, "note", "Note") || "").trim();
      const reasonCode = String(
        getVal(row, "reasonCode", "ReasonCode", "reason_code") || ""
      ).trim();
      const status = String(getVal(row, "status", "Status") || "").trim();

      if (
        isAutoLogoutText(reasonCode) ||
        isAutoLogoutText(reasonName) ||
        isAutoLogoutText(note) ||
        isAutoLogoutText(status)
      ) {
        continue;
      }

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

      const start = parseExportDateTime(createdAt);
      if (!start) continue;

      const durationSec = resolveExportNoteDurationSeconds(row);
      const end = parseExportDateTime(endTime);
      const statusUpper = status.toUpperCase();
      const isActive = ["ACTIVE", "OPEN"].includes(statusUpper) && !end;
      const endMs = isActive
        ? Date.now()
        : end?.getTime() || start.getTime() + durationSec * 1000;

      if (!endMs || endMs < start.getTime()) continue;

      const label = reasonName || note || "Other";
      const startClock = formatExportTime(createdAt);
      const endClock = formatExportTime(
        endTime || new Date(endMs).toISOString()
      );

      notes.push({
        id: Number(getVal(row, "id", "ID") || 0),
        startMs: start.getTime(),
        endMs,
        durationSec: Math.max(
          0,
          durationSec || Math.floor((endMs - start.getTime()) / 1000)
        ),
        detailText: `${startClock}-${endClock} ${label}`,
        reasonName: label,
      });
    }
  }

  notes.sort((a, b) => a.startMs - b.startMs || a.id - b.id);
  return notes;
}

function collectOperatorSessionsForUuid(reportData, uuid) {
  const target = normalizeText(uuid);
  if (!target) return [];

  const sessions = [];

  for (const row of extractRows(reportData)) {
    const sessionUuid = normalizeText(getVal(row, "uuid", "UUID"));
    if (!sessionUuid || sessionUuid !== target) continue;

    const operatorName = String(
      getVal(row, "operatorName", "OperatorName", "operator_name") || ""
    ).trim();
    const operatorNik = String(
      getVal(row, "operatorNik", "OperatorNik", "operator_nik") || ""
    ).trim();

    if (
      isAutoLogoutText(operatorName) ||
      isAutoLogoutText(operatorNik)
    ) {
      continue;
    }

    const loginTime = String(
      getVal(row, "loginTime", "LoginTime", "login_time") || ""
    ).trim();
    const logoutTime = String(
      getVal(row, "logoutTime", "LogoutTime", "logout_time") || ""
    ).trim();
    const status = String(getVal(row, "status", "Status") || "")
      .trim()
      .toUpperCase();

    const start = parseExportDateTime(loginTime);
    if (!start) continue;

    const end = parseExportDateTime(logoutTime);
    const isActive = ["ACTIVE", "OPEN"].includes(status) && !end;
    const endMs = isActive ? Date.now() : end?.getTime();

    sessions.push({
      id: Number(getVal(row, "id", "ID", "sessionId", "SessionID") || 0),
      startMs: start.getTime(),
      endMs: endMs && endMs >= start.getTime() ? endMs : Number.MAX_SAFE_INTEGER,
      operatorName: operatorName || operatorNik || "",
    });
  }

  sessions.sort((a, b) => a.startMs - b.startMs || a.id - b.id);
  return sessions;
}

function matchOperatorName(sessions, eventStartMs, eventEndMs) {
  if (!Array.isArray(sessions) || !sessions.length) return "";

  const pointMs =
    Number.isFinite(eventStartMs) && eventStartMs > 0
      ? eventStartMs
      : Number.isFinite(eventEndMs) && eventEndMs > 0
        ? eventEndMs
        : null;

  if (pointMs == null) return "";

  const covering = sessions.filter(
    (session) => pointMs >= session.startMs && pointMs <= session.endMs
  );

  if (covering.length) {
    covering.sort((a, b) => b.startMs - a.startMs || b.id - a.id);
    return covering[0].operatorName || "";
  }

  let nearest = null;
  let nearestDiff = Number.POSITIVE_INFINITY;

  for (const session of sessions) {
    const mid = (session.startMs + session.endMs) / 2;
    const diff = Math.abs(pointMs - mid);
    if (diff < nearestDiff) {
      nearestDiff = diff;
      nearest = session;
    }
  }

  return nearest?.operatorName || "";
}

/** Note yang overlap jendela jeda [gapStartMs, gapEndMs]. */
function matchNotesInGap(notes, gapStartMs, gapEndMs) {
  if (
    gapStartMs == null ||
    gapEndMs == null ||
    !Number.isFinite(gapStartMs) ||
    !Number.isFinite(gapEndMs) ||
    gapEndMs <= gapStartMs
  ) {
    return { detailLossTime: "-", lossTimeSec: null, lossTime: "-" };
  }

  const matched = [];

  for (const note of notes) {
    const overlapStart = Math.max(note.startMs, gapStartMs);
    const overlapEnd = Math.min(note.endMs, gapEndMs);
    if (overlapEnd <= overlapStart) continue;

    const overlapSec = Math.max(
      0,
      Math.floor((overlapEnd - overlapStart) / 1000)
    );

    matched.push({
      ...note,
      overlapSec,
      // Tampilkan jam note asli (bukan clip), sesuai contoh HH.mm-HH.mm Alasan.
      detailText: note.detailText,
    });
  }

  if (!matched.length) {
    return { detailLossTime: "-", lossTimeSec: 0, lossTime: "00:00:00" };
  }

  const lossTimeSec = matched.reduce(
    (sum, item) => sum + toNumber(item.overlapSec || item.durationSec),
    0
  );

  return {
    detailLossTime: matched.map((item) => item.detailText).join("\n"),
    lossTimeSec,
    lossTime: formatDurationHHMMSS(lossTimeSec),
  };
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
    const [processData, reportData] = await Promise.all([
      getProcessDetail(date, targetUuid),
      getMachineOperatorReport(date, { forExport: 1 }).catch(() => null),
    ]);

    detailData.value = processData;
    operatorLossNotes.value = collectLossNotesForUuid(reportData, targetUuid);
    operatorSessions.value = collectOperatorSessionsForUuid(
      reportData,
      targetUuid
    );
    detailPage.value = 1;
  } catch (err) {
    detailError.value = `Gagal mengambil detail proses: ${err.message}`;
    operatorLossNotes.value = [];
    operatorSessions.value = [];
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

function parseEventTimeMs(value) {
  if (!value) return null;

  const raw = String(value).trim();
  if (!raw) return null;

  const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
  const time = new Date(normalized).getTime();

  return Number.isFinite(time) ? time : null;
}

function eventStartMs(e) {
  return parseEventTimeMs(e?.startTime);
}

function eventEndMs(e) {
  return parseEventTimeMs(e?.endTime) ?? parseEventTimeMs(e?.startTime);
}

/** Jeda = Start output berikutnya − End output sebelumnya (+ losstime di jendela jeda). */
function withGapBetweenOutputs(events, lossNotes = [], sessions = []) {
  const chronological = [...events].sort((a, b) => {
    const startA = eventStartMs(a) || eventTimeValue(a);
    const startB = eventStartMs(b) || eventTimeValue(b);
    if (startA !== startB) return startA - startB;
    return eventTimeValue(a) - eventTimeValue(b);
  });

  return chronological.map((event, index) => {
    const operatorName = matchOperatorName(
      sessions,
      eventStartMs(event),
      eventEndMs(event)
    );

    if (index === 0) {
      return {
        ...event,
        operatorName,
        gapSec: null,
        gapTime: "-",
        detailLossTime: "-",
        lossTimeSec: null,
        lossTime: "-",
      };
    }

    const prevEnd = eventEndMs(chronological[index - 1]);
    const currStart = eventStartMs(event);

    if (prevEnd == null || currStart == null) {
      return {
        ...event,
        operatorName,
        gapSec: null,
        gapTime: "-",
        detailLossTime: "-",
        lossTimeSec: null,
        lossTime: "-",
      };
    }

    const gapSec = Math.max(0, Math.floor((currStart - prevEnd) / 1000));
    const loss = matchNotesInGap(lossNotes, prevEnd, currStart);

    return {
      ...event,
      operatorName,
      gapSec,
      gapTime: formatDurationHHMMSS(gapSec),
      detailLossTime: loss.detailLossTime,
      lossTimeSec: loss.lossTimeSec,
      lossTime: loss.lossTime,
    };
  });
}

const detailEvents = computed(() => {
  const events = detailData.value?.events || [];
  const withGap = withGapBetweenOutputs(
    events,
    operatorLossNotes.value,
    operatorSessions.value
  );

  return withGap.sort((a, b) => {
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

const visibleDetailPages = computed(() => {
  const pages = [];
  const total = totalDetailPages.value;
  const current = detailPage.value;

  let start = Math.max(1, current - 2);
  let end = Math.min(total, current + 2);

  if (current <= 3) {
    end = Math.min(total, 5);
  }

  if (current >= total - 2) {
    start = Math.max(1, total - 4);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
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
  operatorLossNotes.value = [];
  operatorSessions.value = [];
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
    visibleDetailPages,

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