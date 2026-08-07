<script setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useMechanicDashboard } from "../composables/useMechanicDashboard";

const props = defineProps({
  registerSignal: {
    type: Number,
    default: 0,
  },
});

function useColumnAutoScroll(containerRef, options = {}) {
  const speed = options.speed ?? 0.45;
  const pauseBottomMs = options.pauseBottomMs ?? 2800;
  const pauseTopMs = options.pauseTopMs ?? 1800;

  let rafId = 0;
  let phase = "top-pause";
  let phaseUntil = 0;
  let userPaused = false;
  let wheelTimer = null;
  let listenersAttached = false;

  const onEnter = () => {
    userPaused = true;
  };

  const onLeave = () => {
    userPaused = false;
  };

  const onWheel = () => {
    userPaused = true;
    clearTimeout(wheelTimer);
    wheelTimer = setTimeout(() => {
      userPaused = false;
    }, 4000);
  };

  const attachListeners = (el) => {
    if (!el || listenersAttached) return;
    el.addEventListener("mouseenter", onEnter);
    el.addEventListener("mouseleave", onLeave);
    el.addEventListener("wheel", onWheel, { passive: true });
    listenersAttached = true;
  };

  const detachListeners = (el) => {
    if (!el || !listenersAttached) return;
    el.removeEventListener("mouseenter", onEnter);
    el.removeEventListener("mouseleave", onLeave);
    el.removeEventListener("wheel", onWheel);
    listenersAttached = false;
  };

  const tick = (now) => {
    const el = containerRef.value;
    if (!el) {
      rafId = requestAnimationFrame(tick);
      return;
    }

    attachListeners(el);

    const overflow = el.scrollHeight - el.clientHeight;
    if (overflow <= 4) {
      el.scrollTop = 0;
      rafId = requestAnimationFrame(tick);
      return;
    }

    if (userPaused) {
      rafId = requestAnimationFrame(tick);
      return;
    }

    const ts = now ?? performance.now();

    if (phase === "top-pause") {
      if (ts >= phaseUntil) {
        phase = "scrolling";
      }
      rafId = requestAnimationFrame(tick);
      return;
    }

    if (phase === "bottom-pause") {
      if (ts >= phaseUntil) {
        el.scrollTop = 0;
        phase = "top-pause";
        phaseUntil = ts + pauseTopMs;
      }
      rafId = requestAnimationFrame(tick);
      return;
    }

    el.scrollTop += speed;

    if (el.scrollTop >= overflow - 1) {
      el.scrollTop = overflow;
      phase = "bottom-pause";
      phaseUntil = ts + pauseBottomMs;
    }

    rafId = requestAnimationFrame(tick);
  };

  const start = () => {
    cancelAnimationFrame(rafId);
    phase = "top-pause";
    phaseUntil = performance.now() + pauseTopMs;
    rafId = requestAnimationFrame(tick);
  };

  const stop = () => {
    cancelAnimationFrame(rafId);
    clearTimeout(wheelTimer);
    detachListeners(containerRef.value);
  };

  return { start, stop };
}

const openBodyRef = ref(null);
const busyBodyRef = ref(null);
const doneBodyRef = ref(null);

const openScroll = useColumnAutoScroll(openBodyRef);
const busyScroll = useColumnAutoScroll(busyBodyRef);
const doneScroll = useColumnAutoScroll(doneBodyRef, { speed: 0.5 });

function startAllAutoScroll() {
  nextTick(() => {
    openScroll.start();
    busyScroll.start();
    doneScroll.start();
  });
}

onMounted(() => {
  startAllAutoScroll();
});

onUnmounted(() => {
  openScroll.stop();
  busyScroll.stop();
  doneScroll.stop();
});

const {
  openRows,
  busyRows,
  doneRows,
  loading,
  actingId,
  errorMessage,
  notice,
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
  requestClaim,
  requestDone,
  confirmNikModal,
  closeNikModal,
  openRegisterModal,
  closeRegisterModal,
  focusRegisterRfid,
  submitRegisterRFID,
} = useMechanicDashboard();

watch(
  () => [
    openRows.value.map((row) => row.id).join(","),
    busyRows.value.map((row) => row.id).join(","),
    doneRows.value.map((row) => row.id).join(","),
  ],
  () => {
    startAllAutoScroll();
  }
);

watch(
  () => props.registerSignal,
  (next, prev) => {
    if (next === prev) return;
    openRegisterModal();
  }
);
</script>

<template>
  <section class="mechanic-page">
    <p v-if="notice" class="notice">{{ notice }}</p>
    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div class="board">
      <!-- Kolom 1 -->
      <section class="column open" :class="{ 'has-alert': openCount > 0 }">
        <header class="column-head">
          <div>
            <strong>Belum Ditangani</strong>
            <small>Menunggu mekanik</small>
          </div>
          <span class="count">{{ openCount }}</span>
        </header>

        <div ref="openBodyRef" class="column-body">
          <p v-if="loading && !openRows.length" class="empty">Memuat...</p>
          <p v-else-if="!openRows.length" class="empty">Kosong</p>

          <article
            v-for="item in openRows"
            :key="item.id"
            class="row-card"
            :class="{ blink: openCount > 0 }"
          >
            <div class="row-main">
              <h3>{{ item.machineName }}</h3>
              <p class="location">{{ item.location || "-" }}</p>
              <p class="meta-line">
                {{ item.operatorNik }}
                <template v-if="item.operatorNik && item.operatorName"> - </template>
                {{ item.operatorName || "-" }}
              </p>
              <p class="meta-line muted">
                Mulai {{ item.startClock }} · {{ item.durationText }}
              </p>
              <p v-if="item.closedByOperator" class="meta-line warn">
                Operator sudah close · Loss {{ item.lossText }}
              </p>
            </div>
            <button
              type="button"
              class="btn primary"
              :disabled="actingId === item.id"
              @click="requestClaim(item)"
            >
              {{ actingId === item.id ? "..." : "Ambil" }}
            </button>
          </article>
        </div>
      </section>

      <!-- Kolom 2 -->
      <section class="column busy">
        <header class="column-head">
          <div>
            <strong>Dikerjakan</strong>
            <small>Sedang diperbaiki</small>
          </div>
          <span class="count">{{ busyCount }}</span>
        </header>

        <div ref="busyBodyRef" class="column-body">
          <p v-if="!busyRows.length" class="empty">Kosong</p>

          <article v-for="item in busyRows" :key="item.id" class="row-card">
            <div class="row-main">
              <h3>{{ item.machineName }}</h3>
              <p class="location">{{ item.location || "-" }}</p>
              <p class="meta-line">
                Mekanik:
                {{ item.claimedByNik }}
                <template v-if="item.claimedByName"> - {{ item.claimedByName }}</template>
              </p>
              <p class="meta-line muted">
                Ambil {{ item.claimedClock }} · Kerja {{ item.workText }}
              </p>
              <p
                v-if="item.closedByOperator"
                class="meta-line warn"
              >
                Operator sudah close · Loss {{ item.lossText }}
              </p>
              <p v-else class="meta-line muted">
                Operator masih aktif · Loss {{ item.lossText }}
              </p>
            </div>
            <button
              type="button"
              class="btn success"
              :disabled="actingId === item.id"
              @click="requestDone(item)"
            >
              {{ actingId === item.id ? "..." : "Selesai" }}
            </button>
          </article>
        </div>
      </section>

      <!-- Kolom 3 -->
      <section class="column done">
        <header class="column-head">
          <div>
            <strong>Selesai Hari Ini</strong>
            <small>Histori hari ini</small>
          </div>
          <span class="count">{{ doneToday }}</span>
        </header>

        <div ref="doneBodyRef" class="column-body">
          <p v-if="!doneRows.length" class="empty">Belum ada</p>

          <article v-for="item in doneRows" :key="item.id" class="row-card done-card">
            <div class="row-main">
              <h3>{{ item.machineName }}</h3>
              <p class="location">{{ item.location || "-" }}</p>

              <p class="meta-line">
                <template v-if="item.closedByMechanic">
                  Mekanik selesai:
                  {{ item.mechanicDoneByNik || item.claimedByNik }}
                  <template v-if="item.mechanicDoneByName || item.claimedByName">
                    -
                    {{ item.mechanicDoneByName || item.claimedByName }}
                  </template>
                  · {{ item.mechanicDoneClock }}
                </template>
                <template v-else>Mekanik: belum Selesai</template>
              </p>

              <p class="meta-line muted">
                Tunggu {{ item.waitText }}
                · Kerja mekanik {{ item.workText }}
              </p>
            </div>
          </article>
        </div>
      </section>
    </div>

    <div
      v-if="nikModalOpen"
      class="modal-backdrop"
      @click.self="closeNikModal"
    >
      <form class="nik-modal" @submit.prevent="confirmNikModal">
        <p class="modal-kicker">NIK / KARTU RFID</p>
        <h3>{{ nikModalTitle }}</h3>
        <p class="modal-hint">{{ nikModalHint }}</p>

        <label>
          NIK atau Tap Kartu
          <input
            ref="nikInputRef"
            v-model="nikModalInput"
            type="text"
            inputmode="numeric"
            placeholder="Ketik NIK / tap kartu"
            autocomplete="off"
            :disabled="nikModalLoading"
          />
        </label>

        <p v-if="nikModalError" class="modal-error">{{ nikModalError }}</p>

        <div class="modal-actions">
          <button
            type="button"
            class="btn ghost"
            :disabled="nikModalLoading"
            @click="closeNikModal"
          >
            Batal
          </button>
          <button
            type="submit"
            class="btn primary"
            :disabled="nikModalLoading"
          >
            {{ nikModalLoading ? "Memproses..." : "Konfirmasi" }}
          </button>
        </div>
      </form>
    </div>

    <div
      v-if="registerOpen"
      class="modal-backdrop"
      @click.self="closeRegisterModal"
    >
      <form class="nik-modal" @submit.prevent="submitRegisterRFID">
        <p class="modal-kicker">DAFTAR KARTU</p>
        <h3>Hubungkan RFID ke NIK</h3>
        <p class="modal-hint">
          Satu kartu = satu NIK. Hanya untuk bagian Mekanik.
        </p>

        <label>
          NIK Mekanik
          <input
            v-model="registerNik"
            type="text"
            inputmode="numeric"
            placeholder="Masukkan NIK"
            autocomplete="off"
            :disabled="registerLoading"
            @keydown.enter.prevent="focusRegisterRfid"
          />
        </label>

        <label>
          Tap Kartu RFID
          <input
            ref="registerRfidRef"
            v-model="registerRfid"
            type="text"
            inputmode="numeric"
            placeholder="Fokus di sini lalu tap kartu"
            autocomplete="off"
            :disabled="registerLoading"
          />
        </label>

        <p v-if="registerError" class="modal-error">{{ registerError }}</p>

        <div class="modal-actions">
          <button
            type="button"
            class="btn ghost"
            :disabled="registerLoading"
            @click="closeRegisterModal"
          >
            Batal
          </button>
          <button
            type="submit"
            class="btn primary"
            :disabled="registerLoading"
          >
            {{ registerLoading ? "Menyimpan..." : "Simpan" }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.mechanic-page {
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: calc(100vh - 200px);
  min-height: 520px;
  max-height: calc(100vh - 200px);
  overflow: hidden;
}

.btn {
  height: 36px;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: 0 14px;
  font-weight: 800;
  cursor: pointer;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn.primary {
  background: #2563eb;
  color: #fff;
}

.btn.success {
  background: #16a34a;
  color: #fff;
}

.btn.ghost {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

.notice,
.error {
  margin: 0;
  padding: 10px 12px;
  border-radius: 10px;
  font-size: 13px;
}

.notice {
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}

.error {
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fecaca;
}

.board {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  align-items: stretch;
}

.column {
  display: grid;
  grid-template-rows: auto 1fr;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  min-height: 0;
  overflow: hidden;
}

.column.open {
  border-color: #fecaca;
  background: #fff5f5;
}

.column.open.has-alert {
  animation: open-column-blink 1.1s ease-in-out infinite;
}

.column.open.has-alert .count {
  background: #ef4444;
  color: #fff;
  border-color: #dc2626;
  animation: open-count-blink 1.1s ease-in-out infinite;
}

.column.open.has-alert .row-card.blink {
  border-color: #f87171;
  animation: open-card-blink 1.1s ease-in-out infinite;
}

@keyframes open-column-blink {
  0%,
  100% {
    background: #fff5f5;
    border-color: #fecaca;
    box-shadow: none;
  }
  50% {
    background: #fecaca;
    border-color: #ef4444;
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.25);
  }
}

@keyframes open-count-blink {
  0%,
  100% {
    transform: scale(1);
    background: #ef4444;
  }
  50% {
    transform: scale(1.12);
    background: #b91c1c;
  }
}

@keyframes open-card-blink {
  0%,
  100% {
    background: #ffffff;
    border-color: #fca5a5;
  }
  50% {
    background: #fee2e2;
    border-color: #ef4444;
  }
}

.column.busy {
  border-color: #fde68a;
  background: #fffbeb;
}

.column.done {
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.column-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 14px 14px 12px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
}

.column-head strong {
  display: block;
  font-size: 15px;
  color: #0f172a;
}

.column-head small {
  display: block;
  margin-top: 2px;
  color: #64748b;
  font-size: 12px;
}

.count {
  min-width: 36px;
  height: 36px;
  padding: 0 6px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 15px;
  line-height: 1;
  font-variant-numeric: tabular-nums;
  background: #fff;
  border: 1px solid rgba(15, 23, 42, 0.08);
  color: #0f172a;
  box-sizing: border-box;
  text-align: center;
}

.column.open .count {
  color: #b91c1c;
  border-color: #fecaca;
}

.column.busy .count {
  color: #b45309;
  border-color: #fde68a;
}

.column.done .count {
  color: #15803d;
  border-color: #bbf7d0;
}

.column-body {
  min-height: 0;
  padding: 10px;
  overflow-x: hidden;
  overflow-y: auto;
  display: grid;
  gap: 8px;
  align-content: start;
  scrollbar-width: thin;
  scrollbar-color: rgba(15, 23, 42, 0.22) transparent;
}

.column-body::-webkit-scrollbar {
  width: 6px;
}

.column-body::-webkit-scrollbar-thumb {
  background: rgba(15, 23, 42, 0.18);
  border-radius: 999px;
}

.empty {
  margin: 0;
  padding: 18px 10px;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 700;
}

.row-card {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px;
}

.row-card.done-card {
  grid-template-columns: 1fr;
}

.row-main h3 {
  margin: 0;
  font-size: 15px;
  line-height: 1.25;
  color: #0f172a;
}

.location {
  margin: 4px 0 0;
  font-size: 12px;
  font-weight: 800;
  color: #64748b;
}

.meta-line {
  margin: 4px 0 0;
  font-size: 12px;
  font-weight: 700;
  color: #334155;
}

.meta-line.muted {
  color: #94a3b8;
  font-weight: 600;
}

.meta-line.warn {
  color: #b45309;
  font-weight: 800;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  display: grid;
  place-items: center;
  z-index: 80;
  padding: 16px;
}

.nik-modal {
  width: min(420px, 100%);
  background: #fff;
  border-radius: 18px;
  border: 1px solid #e2e8f0;
  padding: 18px;
  display: grid;
  gap: 10px;
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.25);
}

.modal-kicker {
  margin: 0;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  color: #c2410c;
}

.nik-modal h3 {
  margin: 0;
  font-size: 22px;
  color: #0f172a;
}

.modal-hint {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.nik-modal label {
  display: grid;
  gap: 6px;
  font-size: 12px;
  font-weight: 800;
  color: #475569;
}

.nik-modal input {
  height: 46px;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  padding: 0 14px;
  font-size: 18px;
  font-weight: 700;
}

.modal-error {
  margin: 0;
  color: #b91c1c;
  font-size: 13px;
  font-weight: 700;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

@media (max-width: 980px) {
  .mechanic-page {
    height: auto;
    max-height: none;
    overflow: visible;
  }

  .board {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .column {
    min-height: 280px;
    max-height: 360px;
  }
}
</style>
