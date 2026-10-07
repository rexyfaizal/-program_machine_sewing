<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import AiInsightPanel from "./AiInsightPanel.vue";

const props = defineProps({
  open: { type: Boolean, default: false },
  sending: { type: Boolean, default: false },
  messages: { type: Array, default: () => [] },
  errorMessage: { type: String, default: "" },
  modelValue: { type: String, default: "" },
  insight: { type: Object, default: null },
  insightLoading: { type: Boolean, default: false },
  insightError: { type: String, default: "" },
  insightUpdatedAt: { type: String, default: "" },
  initialTab: { type: String, default: "chat" },
  executive: { type: Object, default: null },
  executiveLoading: { type: Boolean, default: false },
  chatFocusTick: { type: Number, default: 0 },
});

const emit = defineEmits([
  "update:open",
  "update:modelValue",
  "send",
  "apply-action",
  "quick",
  "refresh-insight",
  "open-insight",
  "briefing",
  "compare-yesterday",
  "open-presentation",
  "refresh-executive",
  "change-since",
  "insight-filter",
  "insight-chart",
  "insight-ask",
  "insight-why-bad",
]);

const health = computed(() => props.executive?.health || null);

const listRef = ref(null);
const tab = ref(props.initialTab === "insight" ? "insight" : "chat");
const chatExpanded = ref(false);

const highCount = computed(() => Number(props.insight?.highCount || 0));
const isInsightFullscreen = computed(() => props.open && tab.value === "insight");
const statusText = computed(() =>
  props.sending ? "Sedang menganalisa data mesin..." : "Siap membantu"
);

const hasPriorityAlert = computed(() => highCount.value > 0);

const robotCalloutTitle = computed(() => {
  if (!hasPriorityAlert.value) return "Butuh bantuan?";
  const n = highCount.value;
  if (n >= 10) return "Ada banyak isu";
  if (n >= 5) return "Perlu dicek";
  return "Ada update";
});

const robotCalloutSub = computed(() => {
  if (!hasPriorityAlert.value) return "Klik, saya bantu";
  return "Klik saya untuk Insight";
});

function lockScroll(lock) {
  document.body.style.overflow = lock ? "hidden" : "";
}

async function scrollMessages() {
  await nextTick();
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight;
  }
}

watch(() => props.messages.length, scrollMessages);
watch(() => props.sending, scrollMessages);

function onWindowKeydown(event) {
  if (event.key === "Escape" && props.open) {
    close();
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && tab.value === "insight") {
      emit("open-insight");
    }
    if (!isOpen) {
      chatExpanded.value = false;
    }
    lockScroll(isOpen && tab.value === "insight");
    if (isOpen) {
      window.addEventListener("keydown", onWindowKeydown);
    } else {
      window.removeEventListener("keydown", onWindowKeydown);
    }
  }
);

watch(tab, (value) => {
  if (value === "insight" && props.open) {
    emit("open-insight");
  }
  if (value !== "chat") {
    chatExpanded.value = false;
  }
  lockScroll(props.open && value === "insight");
});

watch(
  () => props.chatFocusTick,
  (tick) => {
    if (tick > 0) {
      tab.value = "chat";
      emit("update:open", true);
      lockScroll(false);
    }
  }
);

onBeforeUnmount(() => {
  lockScroll(false);
  window.removeEventListener("keydown", onWindowKeydown);
});

function onSubmit() {
  emit("send");
}

function close() {
  chatExpanded.value = false;
  emit("update:open", false);
  lockScroll(false);
}

function openPanel(nextTab = "chat") {
  tab.value = nextTab === "insight" ? "insight" : "chat";
  emit("update:open", true);
  if (tab.value === "insight") {
    emit("open-insight");
    lockScroll(true);
  }
}

function openFromRobot() {
  openPanel(highCount.value > 0 ? "insight" : "chat");
}
</script>

<template>
  <div class="copilot-root" :class="{ 'is-insight': isInsightFullscreen, 'is-chat': open && !isInsightFullscreen }">
    <button
      v-if="!open"
      type="button"
      class="robot-fab"
      :class="{ 'has-alert': hasPriorityAlert }"
      title="Asisten Produksi"
      aria-label="Buka asisten produksi"
      @click="openFromRobot"
    >
      <span
        class="robot-bubble"
        :class="{ 'is-alert': hasPriorityAlert }"
        aria-hidden="true"
      >
        <strong>{{ robotCalloutTitle }}</strong>
        <small>{{ robotCalloutSub }}</small>
      </span>
      <span v-if="highCount" class="robot-badge">
        <i class="robot-badge__ring" aria-hidden="true" />
        {{ highCount > 99 ? "99+" : highCount }}
      </span>
      <span class="robot-aura" aria-hidden="true" />
      <span class="robot" aria-hidden="true">
        <span class="robot-antenna"><i /></span>
        <span class="robot-head">
          <span class="robot-face">
            <span class="robot-eye left"><i /></span>
            <span class="robot-eye right"><i /></span>
            <span class="robot-mouth" />
          </span>
        </span>
        <span class="robot-body">
          <span class="robot-chest" />
          <span class="robot-arm left" />
          <span class="robot-arm right" />
        </span>
      </span>
      <span class="robot-label">Asisten</span>
    </button>

    <!-- Insight: layar penuh solid, di atas semua elemen dashboard -->
    <div v-else-if="isInsightFullscreen" class="insight-overlay" role="dialog" aria-modal="true">
      <div class="insight-shell">
        <header class="insight-head">
          <div class="insight-head__left">
            <h3>AI Insight</h3>
            <p>Keputusan 1 jam ke depan · menyesuaikan layar</p>
          </div>

          <div class="insight-head__center">
            <div class="doors doors-inline">
              <button type="button" class="door" @click="tab = 'chat'">Tanya AI</button>
              <button type="button" class="door active">
                Insight
                <span v-if="highCount" class="tab-badge">{{ highCount }}</span>
              </button>
              <button
                type="button"
                class="door door-accent"
                :disabled="executiveLoading"
                @click="emit('open-presentation')"
              >
                Presentasi
              </button>
            </div>
          </div>

          <div class="insight-head__right">
            <button
              type="button"
              class="btn-refresh"
              :disabled="insightLoading"
              @click="emit('refresh-insight')"
            >
              {{ insightLoading ? "Analisa..." : "Refresh" }}
            </button>
            <button type="button" class="btn-ghost" @click="close">Tutup</button>
          </div>
        </header>

        <div class="insight-body">
          <AiInsightPanel
            embedded
            fullscreen
            hide-toolbar
            :insight="insight"
            :loading="insightLoading"
            :error-message="insightError"
            :last-updated-at="insightUpdatedAt"
            @refresh="emit('refresh-insight')"
            @filter-item="emit('insight-filter', $event)"
            @open-chart="emit('insight-chart', $event)"
            @ask-ai="emit('insight-ask', $event)"
            @why-bad="emit('insight-why-bad', $event)"
          />
        </div>
      </div>
    </div>

    <!-- Chat: panel mengambang (dashboard tetap terlihat) -->
    <section
      v-else
      class="panel panel-chat"
      :class="{ 'is-expanded': chatExpanded }"
      role="dialog"
      aria-label="Tanya AI"
    >
      <header class="chat-head">
        <div class="chat-head__brand">
          <div class="mini-robot" aria-hidden="true">
            <span class="mini-robot__antenna" />
            <span class="mini-robot__head">
              <i class="mini-robot__eye" />
              <i class="mini-robot__eye" />
            </span>
          </div>
          <div>
            <div class="chat-head__title-row">
              <h3>Asisten Produksi</h3>
            </div>
            <p class="chat-status" :class="{ busy: sending }">
              <span class="live-dot" />
              {{ statusText }}
            </p>
          </div>
        </div>
        <div class="chat-head__actions">
          <button
            type="button"
            class="btn-ghost"
            :title="chatExpanded ? 'Kecilkan panel' : 'Perbesar panel'"
            @click="chatExpanded = !chatExpanded"
          >
            {{ chatExpanded ? "Kecilkan" : "Perbesar" }}
          </button>
          <button type="button" class="btn-ghost" @click="close">Tutup</button>
        </div>
      </header>

      <div v-if="health" class="health-strip" :class="health.tone || 'warn'">
        <div>
          <span>Skor utilitas mesin</span>
          <strong>{{ health.score }}/100 · {{ health.label }}</strong>
        </div>
        <span v-if="highCount" class="health-badge">{{ highCount }} prioritas</span>
      </div>

      <div class="doors">
        <button type="button" class="door active">
          Tanya AI
        </button>
        <button type="button" class="door" @click="tab = 'insight'">
          Insight
          <span v-if="highCount" class="tab-badge">{{ highCount }}</span>
        </button>
        <button
          type="button"
          class="door door-accent"
          :disabled="executiveLoading"
          @click="emit('open-presentation')"
        >
          Presentasi
        </button>
      </div>

      <p class="hint-line">Pilih salah satu: tanya AI, lihat prioritas, atau presentasi untuk atasan.</p>

      <div class="quick">
        <button
          type="button"
          @click="emit('quick', 'Apa yang harus dilakukan 1 jam ke depan berdasarkan data saat ini?')"
        >
          Prioritas 1 jam
        </button>
        <button type="button" @click="emit('change-since')">
          Berubah sejak tadi?
        </button>
        <button
          type="button"
          @click="emit('quick', 'Buka chart produktivitas mesin paling produktif sekarang')"
        >
          Buka chart mesin
        </button>
      </div>

      <div ref="listRef" class="messages">
        <div
          v-for="(msg, idx) in messages"
          :key="`${msg.role}-${idx}`"
          class="msg"
          :class="msg.role"
        >
          <div class="bubble" :class="{ ai: msg.role === 'assistant' }">
            <div v-if="msg.role === 'assistant'" class="ai-label">
              <span class="ai-label__dot" />
              AI Analisa
            </div>
            <pre>{{ msg.content }}</pre>

            <div v-if="msg.actions?.length" class="actions">
              <button
                v-for="(action, aIdx) in msg.actions"
                :key="`${idx}-${aIdx}-${action.type}`"
                type="button"
                class="btn-action"
                @click="emit('apply-action', action)"
              >
                Terapkan: {{ action.label || action.type }}
              </button>
            </div>
          </div>
        </div>

        <div v-if="sending" class="msg assistant">
          <div class="bubble ai typing">
            <div class="ai-label">
              <span class="ai-label__dot" />
              AI berpikir
            </div>
            <div class="typing-dots" aria-hidden="true">
              <i /><i /><i />
            </div>
            <p class="typing-text">Mengolah data mesin, loss, dan prioritas utilisasi mesin...</p>
          </div>
        </div>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <form class="composer" @submit.prevent="onSubmit">
        <input
          :value="modelValue"
          type="text"
          placeholder="Contoh: prioritas 1 jam ke depan..."
          :disabled="sending"
          @input="emit('update:modelValue', $event.target.value)"
        />
        <button type="submit" :disabled="sending || !String(modelValue || '').trim()">
          {{ sending ? "Analisa" : "Kirim" }}
        </button>
      </form>
    </section>
  </div>
</template>

<style scoped>
.copilot-root {
  position: fixed;
  right: 20px;
  bottom: 20px;
  z-index: 12050;
}

.copilot-root.is-insight {
  inset: 0;
  right: 0;
  bottom: 0;
}

.tab-badge {
  display: inline-flex;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  align-items: center;
  justify-content: center;
  background: #fecaca;
  color: #991b1b;
  font-size: 10px;
  font-weight: 900;
}

/* ===== Robot FAB ===== */
.robot-fab {
  position: relative;
  width: 168px;
  height: 196px;
  border: 0;
  padding: 0;
  background: transparent;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  filter: drop-shadow(0 18px 30px rgba(37, 99, 235, 0.28));
  animation: robot-enter 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.robot-fab.has-alert {
  filter: drop-shadow(0 18px 34px rgba(37, 99, 235, 0.38));
}

.robot-bubble {
  position: absolute;
  right: 138px;
  bottom: 96px;
  z-index: 4;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 3px;
  min-width: 128px;
  max-width: 160px;
  padding: 11px 14px;
  border-radius: 18px 18px 6px 18px;
  background: linear-gradient(180deg, #ffffff 0%, #eff6ff 100%);
  color: #1e3a8a;
  border: 1.5px solid #60a5fa;
  box-shadow:
    0 12px 28px rgba(37, 99, 235, 0.22),
    0 0 0 4px rgba(59, 130, 246, 0.12);
  animation:
    callout-pop 0.55s ease 0.25s both,
    callout-nudge 2.4s ease-in-out 0.9s infinite;
  pointer-events: none;
  text-align: left;
}

.robot-bubble.is-alert {
  border-color: #3b82f6;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 55%, #fff 100%);
  box-shadow:
    0 14px 32px rgba(37, 99, 235, 0.28),
    0 0 0 5px rgba(59, 130, 246, 0.16);
  animation:
    callout-pop 0.55s ease 0.25s both,
    callout-nudge 2.2s ease-in-out 0.8s infinite,
    bubble-glow 2.2s ease-in-out 0.8s infinite;
}

.robot-bubble::after {
  content: "";
  position: absolute;
  right: -6px;
  bottom: 16px;
  width: 12px;
  height: 12px;
  background: #eff6ff;
  border-right: 1.5px solid #60a5fa;
  border-bottom: 1.5px solid #60a5fa;
  transform: rotate(-45deg);
}

.robot-bubble.is-alert::after {
  background: #dbeafe;
  border-right-color: #3b82f6;
  border-bottom-color: #3b82f6;
}

.robot-bubble strong {
  font-size: 13.5px;
  font-weight: 800;
  line-height: 1.2;
  color: #1d4ed8;
}

.robot-bubble.is-alert strong {
  color: #1e40af;
}

.robot-bubble small {
  font-size: 11.5px;
  font-weight: 600;
  line-height: 1.25;
  color: #64748b;
}

.robot-bubble.is-alert small {
  color: #334155;
}

.robot-fab:hover .robot {
  transform: translateY(-5px) scale(1.04);
}

.robot-fab:active .robot {
  transform: translateY(0) scale(0.98);
}

.robot-badge {
  position: absolute;
  top: 0;
  right: 10px;
  z-index: 3;
  min-width: 36px;
  height: 36px;
  padding: 0 8px;
  border-radius: 999px;
  background: linear-gradient(160deg, #f87171, #dc2626);
  color: #fff;
  font-size: 15px;
  font-weight: 900;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 16px rgba(220, 38, 38, 0.5);
  animation: badge-pop 1.5s ease-in-out infinite;
}

.robot-badge__ring {
  position: absolute;
  inset: -5px;
  border-radius: 999px;
  border: 2px solid rgba(248, 113, 113, 0.7);
  animation: badge-ring 1.6s ease-out infinite;
  pointer-events: none;
}

.robot-fab.has-alert .robot-aura {
  background: radial-gradient(circle, rgba(59, 130, 246, 0.55), transparent 72%);
  animation: aura-pulse 1.8s ease-in-out infinite;
}

.robot-aura {
  position: absolute;
  bottom: 36px;
  width: 118px;
  height: 118px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(96, 165, 250, 0.4), transparent 70%);
  animation: aura-pulse 2.4s ease-in-out infinite;
  pointer-events: none;
}

.robot {
  position: relative;
  z-index: 1;
  width: 128px;
  height: 144px;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: robot-float 2.8s ease-in-out infinite;
  transition: transform 0.18s ease;
}

.robot-antenna {
  width: 5px;
  height: 18px;
  background: #1d4ed8;
  border-radius: 2px;
  position: relative;
  margin-bottom: 3px;
}

.robot-antenna i {
  position: absolute;
  top: -8px;
  left: 50%;
  width: 14px;
  height: 14px;
  margin-left: -7px;
  border-radius: 50%;
  background: #60a5fa;
  box-shadow: 0 0 14px rgba(96, 165, 250, 0.95);
  animation: antenna-blink 1.6s ease-in-out infinite;
}

.robot-head {
  width: 92px;
  height: 58px;
  border-radius: 22px 22px 16px 16px;
  background: linear-gradient(160deg, #3b82f6, #1d4ed8);
  border: 3px solid #1e40af;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.28);
}

.robot-face {
  width: 68px;
  height: 30px;
  border-radius: 12px;
  background: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 11px;
  position: relative;
  padding: 0 8px;
}

.robot-eye {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #bfdbfe;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  animation: eye-blink 3.6s infinite;
}

.robot-eye i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #0f172a;
}

.robot-mouth {
  position: absolute;
  bottom: 4px;
  width: 18px;
  height: 5px;
  border-radius: 0 0 8px 8px;
  background: #93c5fd;
}

.robot-body {
  position: relative;
  width: 78px;
  height: 50px;
  margin-top: 5px;
  border-radius: 14px 14px 18px 18px;
  background: linear-gradient(180deg, #2563eb, #1d4ed8);
  border: 3px solid #1e3a8a;
}

.robot-chest {
  position: absolute;
  top: 12px;
  left: 50%;
  width: 26px;
  height: 18px;
  margin-left: -13px;
  border-radius: 6px;
  background: #93c5fd;
  box-shadow: 0 0 12px rgba(147, 197, 253, 0.8);
  animation: chest-glow 2s ease-in-out infinite;
}

.robot-arm {
  position: absolute;
  top: 8px;
  width: 15px;
  height: 30px;
  border-radius: 10px;
  background: #2563eb;
  border: 2px solid #1e3a8a;
}

.robot-arm.left {
  left: -16px;
  transform-origin: top center;
  animation: wave-left 2.8s ease-in-out infinite;
}

.robot-arm.right {
  right: -16px;
  transform-origin: top center;
  animation: wave-right 2.8s ease-in-out infinite;
}

.robot-label {
  margin-top: 8px;
  font-size: 15px;
  font-weight: 900;
  color: #1d4ed8;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  padding: 5px 14px;
  line-height: 1.1;
}

@keyframes robot-float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-9px);
  }
}

@keyframes eye-blink {
  0%,
  42%,
  48%,
  100% {
    transform: scaleY(1);
  }
  45% {
    transform: scaleY(0.12);
  }
}

@keyframes antenna-blink {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.45;
    transform: scale(0.85);
  }
}

@keyframes chest-glow {
  0%,
  100% {
    opacity: 0.75;
  }
  50% {
    opacity: 1;
  }
}

@keyframes wave-left {
  0%,
  100% {
    transform: rotate(8deg);
  }
  50% {
    transform: rotate(-14deg);
  }
}

@keyframes wave-right {
  0%,
  100% {
    transform: rotate(-6deg);
  }
  50% {
    transform: rotate(12deg);
  }
}

@keyframes aura-pulse {
  0%,
  100% {
    transform: scale(0.9);
    opacity: 0.55;
  }
  50% {
    transform: scale(1.15);
    opacity: 0.95;
  }
}

@keyframes badge-pop {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.08);
  }
}

@keyframes robot-enter {
  0% {
    opacity: 0;
    transform: translate(28px, 18px) scale(0.86);
  }
  100% {
    opacity: 1;
    transform: translate(0, 0) scale(1);
  }
}

@keyframes callout-pop {
  0% {
    opacity: 0;
    transform: translateX(16px) scale(0.92);
  }
  100% {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
}

@keyframes callout-nudge {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes bubble-glow {
  0%,
  100% {
    box-shadow:
      0 14px 32px rgba(37, 99, 235, 0.24),
      0 0 0 5px rgba(59, 130, 246, 0.12);
  }
  50% {
    box-shadow:
      0 16px 36px rgba(37, 99, 235, 0.34),
      0 0 0 8px rgba(59, 130, 246, 0.22);
  }
}

@keyframes count-pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.06);
  }
}

@keyframes badge-ring {
  0% {
    transform: scale(0.92);
    opacity: 0.85;
  }
  70% {
    transform: scale(1.35);
    opacity: 0;
  }
  100% {
    transform: scale(1.35);
    opacity: 0;
  }
}

@media (max-width: 640px) {
  .robot-bubble {
    right: 114px;
    bottom: 90px;
    min-width: 118px;
    max-width: 148px;
    padding: 9px 11px;
  }

  .robot-bubble strong {
    font-size: 12.5px;
  }

  .robot-bubble small {
    font-size: 11px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .robot-fab,
  .robot-bubble {
    animation: none !important;
  }
}

/* ===== Insight fullscreen ===== */
.insight-overlay {
  position: fixed;
  inset: 0;
  z-index: 12050;
  background: #f1f5f9;
  display: flex;
  flex-direction: column;
}

.insight-shell {
  flex: 1;
  min-height: 0;
  width: min(1400px, 100%);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  padding: clamp(12px, 2vw, 24px);
  box-sizing: border-box;
  background: #fff;
  box-shadow: 0 0 0 1px #e2e8f0;
}

.insight-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 12px 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.insight-head__left h3 {
  margin: 0;
  font-size: clamp(18px, 2.2vw, 24px);
  color: #0f172a;
  line-height: 1.2;
}

.insight-head__left p {
  margin: 4px 0 0;
  font-size: clamp(11px, 1.2vw, 13px);
  color: #64748b;
  font-weight: 700;
}

.insight-head__center {
  justify-self: center;
}

.insight-head__right {
  justify-self: end;
  display: flex;
  align-items: center;
  gap: 8px;
}

.insight-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding-top: 14px;
  overflow: hidden;
}

.tabs {
  display: inline-grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  background: #f8fafc;
  border-radius: 12px;
  padding: 4px;
  min-width: 220px;
}

.tabs button {
  border: 0;
  border-radius: 10px;
  padding: 9px 14px;
  background: transparent;
  color: #64748b;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.tabs button.active {
  background: #fff;
  color: #0f172a;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
}

.btn-ghost,
.btn-refresh {
  border: 0;
  border-radius: 10px;
  padding: 9px 14px;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.btn-ghost {
  background: #f1f5f9;
  color: #334155;
}

.btn-refresh {
  background: #1d4ed8;
  color: #fff;
}

.btn-refresh:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* ===== Chat panel (mengambang, dashboard tetap terlihat) ===== */
.panel-chat {
  width: min(720px, calc(100vw - 24px));
  height: min(900px, calc(100vh - 24px));
  display: flex;
  flex-direction: column;
  gap: 10px;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.14), transparent 42%),
    linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  border: 1px solid #bfdbfe;
  border-radius: 24px;
  padding: 16px;
  box-shadow: 0 28px 60px rgba(37, 99, 235, 0.18);
  animation: rise 0.18s ease;
  overflow: hidden;
  transition: width 0.22s ease, height 0.22s ease;
}

.panel-chat.is-expanded {
  width: min(980px, calc(100vw - 28px));
  height: min(94vh, calc(100vh - 20px));
  border-color: #93c5fd;
  box-shadow:
    0 30px 70px rgba(37, 99, 235, 0.22),
    0 0 0 4px rgba(59, 130, 246, 0.08);
}

@media (max-width: 720px) {
  .panel-chat.is-expanded {
    width: calc(100vw - 16px);
    height: calc(100vh - 16px);
  }

  .chat-head__actions {
    flex-wrap: wrap;
    justify-content: flex-end;
  }
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.chat-head {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: flex-start;
}

.chat-head__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.chat-head__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.chat-head__title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.chat-head h3 {
  margin: 0;
  font-size: 16px;
  color: #0f172a;
}

.chat-status {
  margin: 4px 0 0;
  font-size: 11px;
  color: #15803d;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.chat-status.busy {
  color: #b45309;
}

.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.55);
  animation: live-pulse 1.6s ease-out infinite;
}

.chat-status.busy .live-dot {
  background: #f59e0b;
  animation: live-pulse 0.9s ease-out infinite;
}

.mini-robot {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(160deg, #3b82f6, #1d4ed8);
  border: 2px solid #1e40af;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.28);
}

.mini-robot__antenna {
  width: 3px;
  height: 6px;
  background: #93c5fd;
  border-radius: 2px;
  margin-bottom: 2px;
  position: relative;
}

.mini-robot__antenna::after {
  content: "";
  position: absolute;
  top: -4px;
  left: 50%;
  width: 6px;
  height: 6px;
  margin-left: -3px;
  border-radius: 50%;
  background: #60a5fa;
  animation: antenna-blink 1.6s ease-in-out infinite;
}

.mini-robot__head {
  width: 26px;
  height: 14px;
  border-radius: 6px;
  background: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.mini-robot__eye {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #bfdbfe;
  animation: eye-blink 3.6s infinite;
}

.health-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 14px;
  background: #eff6ff;
  border: 1px solid #93c5fd;
}

.health-strip.warn {
  background: #fff7ed;
  border-color: #fdba74;
}

.health-strip.critical {
  background: #fef2f2;
  border-color: #fca5a5;
}

.health-strip > div span {
  display: block;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.health-strip strong {
  display: block;
  margin-top: 2px;
  font-size: 13px;
  color: #0f172a;
}

.health-badge {
  border-radius: 999px;
  padding: 5px 9px;
  background: #fee2e2;
  color: #b91c1c;
  font-size: 11px;
  font-weight: 900;
  flex-shrink: 0;
}

.doors {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  background: #f8fafc;
  border-radius: 14px;
  padding: 5px;
}

.doors-inline {
  min-width: min(420px, 100%);
}

.door {
  border: 0;
  border-radius: 10px;
  padding: 10px 8px;
  background: transparent;
  color: #64748b;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.door.active {
  background: #fff;
  color: #0f172a;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
}

.door-accent {
  background: #eff6ff;
  color: #1d4ed8;
}

.door-accent:hover:not(:disabled) {
  background: #dbeafe;
}

.door:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.hint-line {
  margin: 0;
  font-size: 11px;
  color: #64748b;
  font-weight: 700;
  line-height: 1.4;
}

.quick {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.quick button {
  border: 1px solid #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
  border-radius: 999px;
  padding: 7px 11px;
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.quick button:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 14px rgba(37, 99, 235, 0.12);
}

.messages {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 2px;
  min-height: 0;
}

.msg {
  display: flex;
}

.msg.user {
  justify-content: flex-end;
}

.msg.assistant {
  justify-content: flex-start;
}

.bubble {
  max-width: 92%;
  border-radius: 14px;
  padding: 10px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.bubble.ai {
  background: linear-gradient(180deg, #ffffff 0%, #eff6ff 100%);
  border-color: #bfdbfe;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.08);
}

.msg.user .bubble {
  background: #eff6ff;
  border-color: #bfdbfe;
}

.ai-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #1d4ed8;
}

.ai-label__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3b82f6;
  animation: antenna-blink 1.4s ease-in-out infinite;
}

.bubble pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  font-size: 12.5px;
  line-height: 1.5;
  color: #0f172a;
  font-weight: 600;
}

.typing-dots {
  display: inline-flex;
  gap: 5px;
  margin: 2px 0 6px;
}

.typing-dots i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #3b82f6;
  animation: typing-bounce 1.1s infinite ease-in-out;
}

.typing-dots i:nth-child(2) {
  animation-delay: 0.15s;
}

.typing-dots i:nth-child(3) {
  animation-delay: 0.3s;
}

.typing-text {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  font-weight: 700;
}

.actions {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.btn-action {
  border: 0;
  border-radius: 10px;
  padding: 9px 11px;
  background: linear-gradient(135deg, #1d4ed8, #2563eb);
  color: #fff;
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
  text-align: left;
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.22);
}

.error {
  margin: 0;
  font-size: 12px;
  color: #b91c1c;
  font-weight: 700;
}

.composer {
  display: flex;
  gap: 8px;
  padding-top: 2px;
}

.composer input {
  flex: 1;
  border: 1px solid #bfdbfe;
  border-radius: 12px;
  padding: 11px 12px;
  font-size: 13px;
  background: #fff;
  box-shadow: inset 0 0 0 1px rgba(153, 246, 228, 0.35);
}

.composer input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.2);
}

.composer button {
  border: 0;
  border-radius: 12px;
  padding: 10px 14px;
  background: linear-gradient(135deg, #1d4ed8, #3b82f6);
  color: #fff;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 8px 16px rgba(37, 99, 235, 0.22);
}

.composer button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes live-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.45);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(34, 197, 94, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0);
  }
}

@keyframes typing-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.45;
  }
  40% {
    transform: translateY(-4px);
    opacity: 1;
  }
}

@media (max-width: 900px) {
  .insight-head {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .insight-head__center,
  .insight-head__right {
    justify-self: stretch;
  }

  .insight-head__right {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .tabs {
    width: 100%;
    min-width: 0;
  }
}
</style>
