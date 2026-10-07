import { ref } from "vue";

import { postAiChat } from "../api/machineApi";
import { collectMachineOperatorNotes } from "../utils/machineTableFormat";

function toMachinePayload(machines = []) {
  return (Array.isArray(machines) ? machines : [])
    .map((m) => {
      const rows = Array.isArray(m?.operatorRows)
        ? m.operatorRows
        : Array.isArray(m?.operatorDisplayRows)
          ? m.operatorDisplayRows
          : [];
      const operatorName = String(
        rows[0]?.operatorName ||
          rows[0]?.name ||
          rows[0]?.label ||
          m?.operatorName ||
          m?.pic ||
          ""
      )
        .trim()
        .replace(/^\d+\s*-\s*/, "");

      return {
        uuid: String(m?.uuid || "").trim(),
        machineName: String(m?.displayMachineName || m?.machineName || "").trim(),
        location: String(m?.location || "").trim(),
        macState: String(m?.macState ?? "").trim(),
        status: String(m?.status || "").trim(),
        productivityPct: Number(m?.productivityPct ?? m?.productivity ?? 0) || 0,
        runtimeSec: Number(m?.runtime ?? m?.runtimeSec ?? 0) || 0,
        procSec: Number(m?.procTime ?? m?.procSec ?? 0) || 0,
        lossTimeSec: Number(m?.lossTime ?? m?.lossTimeSec ?? 0) || 0,
        operatorName,
        operatorNotes: collectMachineOperatorNotes(m),
      };
    })
    .filter((m) => m.uuid);
}

export function useAiCopilot() {
  const open = ref(false);
  const sending = ref(false);
  const errorMessage = ref("");
  const inputText = ref("");
  const messages = ref([
    {
      role: "assistant",
      content:
        "Halo, saya Asisten Produksi.\n\nCara pakai sederhana:\n1) Tanya AI — ketik pertanyaan\n2) Insight — lihat prioritas mesin\n3) Presentasi — tampilan untuk atasan (skor, chart, briefing)\n\nAtau tekan tombol cepat di atas.",
      actions: [],
    },
  ]);

  function toggle() {
    open.value = !open.value;
  }

  function close() {
    open.value = false;
  }

  function pushAssistantMessage(content, options = {}) {
    const text = String(content || "").trim();
    if (!text) return;
    messages.value.push({
      role: "assistant",
      content: text,
      actions: Array.isArray(options.actions) ? options.actions : [],
      warning: options.warning || "",
      model: options.model || "executive",
    });
    open.value = true;
  }

  function pushLocalExchange(userText, assistantText, options = {}) {
    const q = String(userText || "").trim();
    const a = String(assistantText || "").trim();
    if (!q && !a) return;
    if (q) {
      messages.value.push({
        role: "user",
        content: q,
        actions: [],
      });
    }
    if (a) {
      messages.value.push({
        role: "assistant",
        content: a,
        actions: Array.isArray(options.actions) ? options.actions : [],
        warning: options.warning || "",
        model: options.model || "executive",
      });
    }
    open.value = true;
  }

  async function ask(question, context = {}) {
    const q = String(question || inputText.value || "").trim();
    if (!q || sending.value) return;

    errorMessage.value = "";
    sending.value = true;
    inputText.value = "";

    messages.value.push({
      role: "user",
      content: q,
      actions: [],
    });

    try {
      const history = messages.value
        .filter((m) => m.role === "user" || m.role === "assistant")
        .slice(-8)
        .map((m) => ({
          role: m.role,
          content: String(m.content || "").trim(),
        }))
        .filter((m) => m.content);

      // Jangan kirim pesan user terakhir dua kali di history (sudah di question).
      const historyWithoutLastUser = history.slice(0, -1);

      const data = await postAiChat({
        question: q,
        page: context.page || "dashboard",
        date: context.date || "",
        shift: context.shift || "",
        location: context.location || "",
        status: context.status || "",
        history: historyWithoutLastUser,
        machines: toMachinePayload(context.machines || []),
        insight: context.insight || null,
      });

      messages.value.push({
        role: "assistant",
        content: String(data?.reply || "Tidak ada jawaban.").trim(),
        actions: Array.isArray(data?.actions) ? data.actions : [],
        warning: data?.warning || "",
        model: data?.model || "",
      });
    } catch (err) {
      errorMessage.value = err?.message || "Gagal menghubungi Copilot.";
      messages.value.push({
        role: "assistant",
        content: `Maaf, Copilot gagal merespons: ${errorMessage.value}`,
        actions: [],
      });
    } finally {
      sending.value = false;
    }
  }

  return {
    open,
    sending,
    errorMessage,
    inputText,
    messages,
    toggle,
    close,
    ask,
    pushAssistantMessage,
    pushLocalExchange,
  };
}
