<script setup>
import { computed, ref, watch } from "vue";
import QRCode from "qrcode";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  machines: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["close"]);

const keyword = ref("");
const baseUrl = ref(window.location.origin);
const qrItems = ref([]);
const generating = ref(false);
const errorMessage = ref("");

const validMachines = computed(() => {
  return props.machines.filter((m) => {
    const uuid = String(m?.uuid || "").trim();
    return uuid && uuid !== "-";
  });
});

const filteredQrItems = computed(() => {
  const key = keyword.value.toLowerCase().trim();

  if (!key) return qrItems.value;

  return qrItems.value.filter((item) => {
    return [
      item.machineName,
      item.originalMachineName,
      item.location,
      item.ip,
      item.uuid,
      item.pic,
      item.spv,
    ]
      .join(" ")
      .toLowerCase()
      .includes(key);
  });
});

function closeModal() {
  emit("close");
}

function cleanBaseUrl(value) {
  return String(value || "")
    .trim()
    .replace(/\/+$/, "");
}

function buildQrUrl(machine) {
  const uuid = String(machine?.uuid || "").trim();
  const url = cleanBaseUrl(baseUrl.value || window.location.origin);

  return `${url}/?page=operator-pic-spv&uuid=${encodeURIComponent(uuid)}`;
}

function normalizeMachine(machine) {
  return {
    uuid: String(machine?.uuid || "").trim(),
    machineName: String(machine?.machineName || machine?.nickName || "-"),
    originalMachineName: String(machine?.originalMachineName || ""),
    location: String(machine?.location || "-"),
    ip: String(machine?.ip || "-"),
    pic: String(machine?.pic || ""),
    spv: String(machine?.spv || ""),
  };
}

async function generateQrCodes() {
  if (!props.show) return;

  generating.value = true;
  errorMessage.value = "";

  try {
    const items = [];

    for (const machine of validMachines.value) {
      const normalized = normalizeMachine(machine);
      const qrUrl = buildQrUrl(normalized);

      const qrDataUrl = await QRCode.toDataURL(qrUrl, {
        errorCorrectionLevel: "M",
        margin: 1,
        width: 260,
        color: {
          dark: "#0f172a",
          light: "#ffffff",
        },
      });

      items.push({
        ...normalized,
        qrUrl,
        qrDataUrl,
      });
    }

    qrItems.value = items;
  } catch (err) {
    errorMessage.value = `Gagal membuat QR Code: ${err.message}`;
  } finally {
    generating.value = false;
  }
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function printQrCodes() {
  const items = filteredQrItems.value;

  if (!items.length) {
    alert("Tidak ada QR Code untuk dicetak.");
    return;
  }

  const htmlCards = items
    .map((item) => {
      return `
        <div class="qr-card">
          <div class="qr-head">
            <h3>${escapeHtml(item.machineName)}</h3>
            <p>${escapeHtml(item.location)}</p>
          </div>

          <img src="${item.qrDataUrl}" alt="QR ${escapeHtml(item.uuid)}" />

          <div class="qr-info">
            <p><b>UUID</b>: ${escapeHtml(item.uuid)}</p>
            <p><b>IP</b>: ${escapeHtml(item.ip)}</p>
            <p><b>URL</b>: ${escapeHtml(item.qrUrl)}</p>
          </div>
        </div>
      `;
    })
    .join("");

  const printWindow = window.open("", "_blank");

  if (!printWindow) {
    alert("Popup print diblokir browser. Izinkan popup terlebih dahulu.");
    return;
  }

  printWindow.document.write(`
    <!doctype html>
    <html>
      <head>
        <title>QR Code Mesin</title>
        <style>
          * {
            box-sizing: border-box;
          }

          body {
            margin: 0;
            padding: 14px;
            font-family: Arial, sans-serif;
            color: #0f172a;
            background: #ffffff;
          }

          .page-title {
            margin: 0 0 12px;
            font-size: 18px;
            font-weight: 800;
          }

          .qr-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 10px;
          }

          .qr-card {
            border: 1px solid #cbd5e1;
            border-radius: 12px;
            padding: 10px;
            page-break-inside: avoid;
            break-inside: avoid;
          }

          .qr-head {
            min-height: 52px;
            margin-bottom: 8px;
          }

          .qr-head h3 {
            margin: 0;
            font-size: 12px;
            line-height: 1.25;
            font-weight: 800;
          }

          .qr-head p {
            margin: 4px 0 0;
            font-size: 10px;
            color: #475569;
            font-weight: 700;
          }

          img {
            display: block;
            width: 100%;
            max-width: 150px;
            margin: 0 auto 8px;
          }

          .qr-info p {
            margin: 3px 0;
            font-size: 8.5px;
            line-height: 1.25;
            word-break: break-all;
          }

          @media print {
            body {
              padding: 8px;
            }

            .qr-grid {
              grid-template-columns: repeat(3, 1fr);
            }

            .qr-card {
              border-color: #111827;
            }
          }
        </style>
      </head>

      <body>
        <h1 class="page-title">QR Code Mesin - Input PIC & SPV</h1>
        <div class="qr-grid">
          ${htmlCards}
        </div>

        <script>
          window.onload = function () {
            window.print();
          };
        <\/script>
      </body>
    </html>
  `);

  printWindow.document.close();
}

watch(
  () => props.show,
  (value) => {
    if (value) {
      generateQrCodes();
    }
  },
  { immediate: true }
);

watch(
  () => [props.machines, baseUrl.value],
  () => {
    if (props.show) {
      generateQrCodes();
    }
  },
  { deep: true }
);
</script>

<template>
  <div v-if="show" class="modal-backdrop" @click.self="closeModal">
    <section class="modal">
      <div class="modal-head">
        <div>
          <p class="eyebrow">QR Code Mesin</p>
          <h3>Generate QR PIC & SPV</h3>
          <p>
            QR berisi URL halaman operator berdasarkan UUID mesin.
          </p>
        </div>

        <button class="btn-secondary btn-small" type="button" @click="closeModal">
          Tutup
        </button>
      </div>

      <div class="toolbar">
        <label class="field base-field">
          <span>Base URL</span>
          <input
            v-model="baseUrl"
            type="text"
            placeholder="Contoh: http://10.5.0.8:5173"
            @change="generateQrCodes"
          />
        </label>

        <label class="field search-field">
          <span>Pencarian</span>
          <input
            v-model="keyword"
            type="text"
            placeholder="Cari mesin, location, IP, UUID..."
          />
        </label>

        <button class="btn-print" type="button" @click="printQrCodes">
          Print QR
        </button>
      </div>

      <div v-if="generating" class="alert">
        Membuat QR Code...
      </div>

      <div v-if="errorMessage" class="alert error">
        {{ errorMessage }}
      </div>

      <div class="summary">
        <span>Total mesin valid: {{ validMachines.length }}</span>
        <span>QR tampil: {{ filteredQrItems.length }}</span>
      </div>

      <div class="qr-scroll">
        <div v-if="!filteredQrItems.length && !generating" class="empty">
          Tidak ada data mesin untuk dibuat QR.
        </div>

        <div v-else class="qr-grid">
          <article
            v-for="item in filteredQrItems"
            :key="item.uuid"
            class="qr-card"
          >
            <div class="qr-title">
              <h4 :title="item.machineName">
                {{ item.machineName }}
              </h4>
              <p>{{ item.location || "-" }}</p>
            </div>

            <img :src="item.qrDataUrl" :alt="`QR ${item.uuid}`" />

            <div class="qr-detail">
              <div>
                <span>UUID</span>
                <strong>{{ item.uuid }}</strong>
              </div>

              <div>
                <span>IP</span>
                <strong>{{ item.ip || "-" }}</strong>
              </div>
            </div>

            <p class="qr-url">
              {{ item.qrUrl }}
            </p>
          </article>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(15, 23, 42, 0.55);
  display: grid;
  place-items: center;
  padding: 20px;
}

.modal {
  width: min(1180px, 100%);
  max-height: 92vh;
  background: #ffffff;
  border: 1px solid #dbeafe;
  border-radius: 26px;
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-head {
  padding: 20px 22px;
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  border-bottom: 1px solid #dbeafe;
  background: linear-gradient(135deg, #ffffff, #eff6ff);
}

.eyebrow {
  margin: 0 0 6px;
  color: #2563eb;
  font-size: 11px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.modal-head h3 {
  margin: 0;
  color: #0f172a;
  font-size: 22px;
  letter-spacing: -0.03em;
}

.modal-head p {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.toolbar {
  padding: 16px 22px;
  display: grid;
  grid-template-columns: minmax(260px, 1.2fr) minmax(220px, 1fr) 130px;
  gap: 12px;
  align-items: stretch;
  border-bottom: 1px solid #dbeafe;
}

.field {
  min-width: 0;
  display: grid;
  gap: 5px;
  padding: 9px 12px;
  border-radius: 14px;
  background: #ffffff;
  border: 1px solid #dbeafe;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.035);
}

.field span {
  color: #64748b;
  font-size: 10px;
  font-weight: 1000;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.field input {
  width: 100%;
  min-width: 0;
  height: 24px;
  border: 0;
  outline: none;
  background: transparent;
  color: #0f172a;
  font-size: 13px;
  font-weight: 800;
}

.btn-print,
.btn-secondary {
  border: 0;
  border-radius: 14px;
  cursor: pointer;
  font-weight: 1000;
}

.btn-print {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: white;
  box-shadow: 0 12px 22px rgba(37, 99, 235, 0.22);
}

.btn-secondary {
  background: #ffffff;
  color: #0f172a;
  border: 1px solid #bfdbfe;
}

.btn-small {
  padding: 9px 12px;
  font-size: 12px;
}

.alert {
  margin: 14px 22px 0;
  padding: 12px 14px;
  border-radius: 14px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1e40af;
  font-size: 13px;
  font-weight: 900;
}

.alert.error {
  background: #fff1f2;
  border-color: #fecaca;
  color: #b91c1c;
}

.summary {
  padding: 12px 22px 0;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.summary span {
  padding: 7px 10px;
  border-radius: 999px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1e40af;
  font-size: 12px;
  font-weight: 900;
}

.qr-scroll {
  padding: 16px 22px 22px;
  overflow: auto;
}

.qr-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: 14px;
}

.qr-card {
  min-width: 0;
  border-radius: 18px;
  border: 1px solid #dbeafe;
  background: #ffffff;
  padding: 14px;
  box-shadow: 0 12px 22px rgba(15, 23, 42, 0.05);
}

.qr-title {
  min-height: 54px;
  margin-bottom: 10px;
}

.qr-title h4 {
  margin: 0;
  color: #0f172a;
  font-size: 13px;
  line-height: 1.3;
  font-weight: 1000;
  display: -webkit-box;
  line-clamp: 2;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.qr-title p {
  margin: 4px 0 0;
  color: #64748b;
  font-size: 11px;
  font-weight: 800;
}

.qr-card img {
  display: block;
  width: 150px;
  height: 150px;
  margin: 0 auto 12px;
}

.qr-detail {
  display: grid;
  gap: 7px;
}

.qr-detail div {
  min-width: 0;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 7px 8px;
}

.qr-detail span {
  display: block;
  color: #64748b;
  font-size: 9px;
  font-weight: 1000;
  text-transform: uppercase;
  margin-bottom: 3px;
}

.qr-detail strong {
  display: block;
  color: #0f172a;
  font-size: 10px;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.qr-url {
  margin: 9px 0 0;
  color: #2563eb;
  font-size: 9px;
  line-height: 1.35;
  font-weight: 800;
  word-break: break-all;
}

.empty {
  padding: 28px;
  text-align: center;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  color: #64748b;
  font-weight: 900;
}

@media (max-width: 820px) {
  .modal-backdrop {
    padding: 12px;
  }

  .modal {
    max-height: 95vh;
    border-radius: 20px;
  }

  .modal-head {
    flex-direction: column;
  }

  .toolbar {
    grid-template-columns: 1fr;
  }

  .btn-print {
    min-height: 46px;
  }

  .qr-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }
}
</style>