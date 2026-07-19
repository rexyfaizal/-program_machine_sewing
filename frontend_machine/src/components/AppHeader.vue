<script setup>
import { computed } from "vue";
import NavTabs from "./NavTabs.vue";

const props = defineProps({
  modelValue: {
    type: String,
    default: "dashboard",
  },
});

const emit = defineEmits(["update:modelValue"]);

const activePageModel = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const headerInfo = computed(() => {
  if (activePageModel.value === "master-ie") {
    return {
      kicker: "MASTER IE",
      title: "Master IE Style & Proses",
      statusTitle: "Master Data",
      statusText: "Style & Proses",
    };
  }

  if (activePageModel.value === "location") {
    return {
      kicker: "MONITORING LOCATION",
      title: "Monitoring Location Machine",
      statusTitle: "Location View",
      statusText: "Machine Layout",
    };
  }

  if (activePageModel.value === "detail") {
    return {
      kicker: "PROCESS DETAIL",
      title: "Detail Productivity Machine",
      statusTitle: "Detail View",
      statusText: "Machine Process",
    };
  }

  return {
    kicker: "MONITORING MACHINE",
    title: "Monitoring Productivity Machine",
    statusTitle: "Live Dashboard",
    statusText: "Production View",
  };
});

const isMasterIePage = computed(() => activePageModel.value === "master-ie");
</script>

<template>
  <header class="app-header" :class="{ 'compact-master': isMasterIePage }">
    <section class="brand-card">
      <div class="left-area">
        <div class="nav-holder">
          <NavTabs v-model="activePageModel" />
        </div>

        <div class="logo-frame">
          <img src="../assets/machine.png" alt="Monitoring Machine" />
        </div>

        <div class="brand-text">
          <div class="brand-kicker">
            <span class="live-dot"></span>
            {{ headerInfo.kicker }}
          </div>

          <h1>{{ headerInfo.title }}</h1>
        </div>
      </div>

      <div class="status-mini">
        <span>{{ headerInfo.statusTitle }}</span>
        <strong>{{ headerInfo.statusText }}</strong>
      </div>
    </section>

    <section v-if="!isMasterIePage" class="legend-card">
      <div class="legend-title">Kategori Produktivitas</div>

      <div class="legend-list">
        <div class="legend-item good">
          <span></span>
          <strong>GOOD</strong>
          <small>≥ 90%</small>
        </div>

        <div class="legend-item normal">
          <span></span>
          <strong>NORMAL</strong>
          <small>80–90%</small>
        </div>

        <div class="legend-item bad">
          <span></span>
          <strong>BAD</strong>
          <small>&lt; 80%</small>
        </div>
      </div>
    </section>

    <section v-else class="legend-card master-card">
      <div class="legend-title">Master Data IE</div>

      <div class="master-info-list">
        <div class="master-info-item">
          <span>1</span>
          <strong>Input Style</strong>
          <small>Contoh: 1101482</small>
        </div>

        <div class="master-info-item">
          <span>2</span>
          <strong>Input Proses</strong>
          <small>Contoh: Quilting Front Kanan</small>
        </div>

        <div class="master-info-item">
          <span>3</span>
          <strong>Dipakai Operator</strong>
          <small>Autocomplete saat scan QR</small>
        </div>
      </div>
    </section>
  </header>
</template>

<style scoped>
.app-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 340px);
  gap: clamp(12px, 1.2vw, 18px);
  margin: 0 0 18px;
  position: relative;
  z-index: 100;
  overflow: visible;
  min-width: 0;
}

.brand-card,
.legend-card {
  border: 1px solid rgba(147, 197, 253, 0.45);
  border-radius: clamp(20px, 1.8vw, 26px);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.92)),
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.16), transparent 36%);
  box-shadow:
    0 18px 40px rgba(15, 23, 42, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
  min-width: 0;
}

.brand-card {
  min-height: clamp(108px, 9vw, 136px);
  padding: clamp(16px, 1.6vw, 24px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: clamp(12px, 1.2vw, 22px);
  position: relative;
  overflow: visible;
  z-index: 200;
}

.brand-card::after {
  content: "";
  position: absolute;
  right: -72px;
  top: -72px;
  width: 200px;
  height: 200px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.08);
  pointer-events: none;
  z-index: 0;
}

.left-area {
  display: flex;
  align-items: center;
  gap: clamp(12px, 1.2vw, 18px);
  min-width: 0;
  position: relative;
  z-index: 5;
  flex: 1;
}

.nav-holder {
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  position: relative;
  z-index: 9999;
  overflow: visible;
}

.nav-holder :deep(*) {
  overflow: visible;
}

.nav-holder :deep(button) {
  box-shadow: none;
}

.nav-holder :deep(.nav-tabs),
.nav-holder :deep(.tabs-wrap),
.nav-holder :deep(.menu-wrap),
.nav-holder :deep(.nav-menu),
.nav-holder :deep(.dropdown),
.nav-holder :deep(.menu-panel) {
  position: relative;
  top: auto;
  left: auto;
  margin: 0;
  z-index: 10000;
  overflow: visible;
}

.nav-holder :deep(.menu-list),
.nav-holder :deep(.dropdown-menu),
.nav-holder :deep(.nav-dropdown),
.nav-holder :deep(ul) {
  z-index: 10001;
}

.logo-frame {
  width: clamp(56px, 4.8vw, 68px);
  height: clamp(56px, 4.8vw, 68px);
  flex: 0 0 auto;
  border-radius: 20px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #ffffff, #eff6ff);
  border: 1px solid #bfdbfe;
  box-shadow:
    0 12px 24px rgba(37, 99, 235, 0.13),
    inset 0 0 0 1px rgba(255, 255, 255, 0.72);
  position: relative;
  z-index: 2;
}

.logo-frame img {
  width: clamp(40px, 3.4vw, 48px);
  height: clamp(40px, 3.4vw, 48px);
  object-fit: contain;
}

.brand-text {
  min-width: 0;
  position: relative;
  z-index: 2;
}

.brand-kicker {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #2563eb;
  font-size: clamp(10px, 0.75vw, 12px);
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  margin-bottom: 8px;
  white-space: nowrap;
}

.live-dot {
  width: 9px;
  height: 9px;
  border-radius: 999px;
  background: #22c55e;
  box-shadow: 0 0 0 5px rgba(34, 197, 94, 0.13);
  flex: 0 0 auto;
}

.brand-text h1 {
  margin: 0;
  color: #0f172a;
  font-size: clamp(26px, 2.25vw, 40px);
  line-height: 1;
  font-weight: 950;
  letter-spacing: -0.055em;
  max-width: 100%;
}

.status-mini {
  flex: 0 0 auto;
  position: relative;
  z-index: 2;
  display: grid;
  gap: 4px;
  padding: 11px 13px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.75);
  border: 1px solid rgba(191, 219, 254, 0.8);
  text-align: right;
  white-space: nowrap;
}

.status-mini span {
  color: #64748b;
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.status-mini strong {
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 950;
}

.legend-card {
  min-height: clamp(108px, 9vw, 136px);
  padding: clamp(16px, 1.5vw, 22px);
  display: grid;
  align-content: center;
  gap: 14px;
  position: relative;
  z-index: 1;
}

.legend-title {
  color: #64748b;
  font-size: 12px;
  font-weight: 950;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.legend-list {
  display: grid;
  gap: 9px;
}

.legend-item {
  display: grid;
  grid-template-columns: 12px auto 1fr;
  align-items: center;
  gap: 9px;
  width: 100%;
  border-radius: 14px;
  padding: 9px 11px;
  font-size: 12px;
  border: 1px solid transparent;
  min-width: 0;
}

.legend-item span {
  width: 10px;
  height: 10px;
  border-radius: 999px;
}

.legend-item strong {
  font-weight: 950;
  white-space: nowrap;
}

.legend-item small {
  justify-self: end;
  font-size: 11px;
  font-weight: 900;
  white-space: nowrap;
}

.legend-item.good {
  background: #dcfce7;
  color: #166534;
  border-color: #bbf7d0;
}

.legend-item.good span {
  background: #22c55e;
}

.legend-item.normal {
  background: #fef3c7;
  color: #92400e;
  border-color: #fde68a;
}

.legend-item.normal span {
  background: #f59e0b;
}

.legend-item.bad {
  background: #fee2e2;
  color: #991b1b;
  border-color: #fecaca;
}

.legend-item.bad span {
  background: #ef4444;
}

.master-card {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.96), rgba(240, 253, 244, 0.9)),
    radial-gradient(circle at top left, rgba(34, 197, 94, 0.14), transparent 36%);
}

.master-info-list {
  display: grid;
  gap: 9px;
}

.master-info-item {
  display: grid;
  grid-template-columns: 26px 1fr;
  gap: 9px;
  align-items: center;
  padding: 9px 11px;
  border-radius: 14px;
  background: #ecfdf5;
  border: 1px solid #bbf7d0;
  min-width: 0;
}

.master-info-item span {
  width: 24px;
  height: 24px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: #22c55e;
  color: white;
  font-size: 12px;
  font-weight: 1000;
}

.master-info-item strong {
  color: #166534;
  font-size: 12px;
  font-weight: 1000;
}

.master-info-item small {
  grid-column: 2;
  margin-top: -5px;
  color: #64748b;
  font-size: 11px;
  font-weight: 800;
}

/* KHUSUS MASTER IE: HEADER COMPACT */
.app-header.compact-master {
  grid-template-columns: minmax(0, 1fr) minmax(240px, 290px);
  gap: 12px;
  margin-bottom: 12px;
}

.app-header.compact-master .brand-card,
.app-header.compact-master .legend-card {
  min-height: 82px;
  border-radius: 18px;
}

.app-header.compact-master .brand-card {
  padding: 12px 16px;
}

.app-header.compact-master .legend-card {
  padding: 10px 12px;
  gap: 6px;
}

.app-header.compact-master .brand-card::after {
  width: 130px;
  height: 130px;
  right: -48px;
  top: -48px;
}

.app-header.compact-master .left-area {
  gap: 12px;
}

.app-header.compact-master .nav-holder {
  width: 40px;
  height: 40px;
}

.app-header.compact-master .logo-frame {
  width: 44px;
  height: 44px;
  border-radius: 14px;
}

.app-header.compact-master .logo-frame img {
  width: 30px;
  height: 30px;
}

.app-header.compact-master .brand-kicker {
  margin-bottom: 4px;
  font-size: 10px;
}

.app-header.compact-master .live-dot {
  width: 8px;
  height: 8px;
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.13);
}

.app-header.compact-master .brand-text h1 {
  font-size: clamp(22px, 1.55vw, 30px);
  line-height: 1.05;
}

.app-header.compact-master .status-mini {
  padding: 7px 10px;
  border-radius: 13px;
}

.app-header.compact-master .status-mini span {
  font-size: 10px;
}

.app-header.compact-master .status-mini strong {
  font-size: 12px;
}

.app-header.compact-master .legend-title {
  font-size: 10px;
}

.app-header.compact-master .master-info-list {
  gap: 5px;
}

.app-header.compact-master .master-info-item {
  padding: 6px 8px;
  border-radius: 11px;
  grid-template-columns: 20px 1fr;
}

.app-header.compact-master .master-info-item span {
  width: 19px;
  height: 19px;
  font-size: 10px;
}

.app-header.compact-master .master-info-item strong {
  font-size: 10.5px;
}

.app-header.compact-master .master-info-item small {
  font-size: 9.5px;
}

@media (max-width: 1366px) {
  .app-header {
    grid-template-columns: 1fr;
  }

  .brand-card {
    min-height: 118px;
  }

  .legend-card {
    min-height: auto;
  }

  .legend-list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .master-info-list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .app-header.compact-master {
    grid-template-columns: 1fr;
  }

  .app-header.compact-master .brand-card {
    min-height: 82px;
  }

  .app-header.compact-master .legend-card {
    min-height: auto;
  }

  .app-header.compact-master .master-info-list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .brand-card {
    align-items: flex-start;
  }

  .left-area {
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .status-mini {
    display: none;
  }

  .brand-text {
    flex-basis: 100%;
  }

  .brand-text h1 {
    font-size: clamp(28px, 7vw, 36px);
  }

  .legend-list,
  .master-info-list,
  .app-header.compact-master .master-info-list {
    grid-template-columns: 1fr;
  }

  .app-header.compact-master .brand-text {
    flex-basis: auto;
  }

  .app-header.compact-master .brand-text h1 {
    font-size: 24px;
  }
}

@media (max-width: 520px) {
  .brand-card {
    border-radius: 22px;
    padding: 16px;
  }

  .nav-holder {
    width: 42px;
    height: 42px;
  }

  .logo-frame {
    width: 56px;
    height: 56px;
  }

  .logo-frame img {
    width: 40px;
    height: 40px;
  }

  .brand-kicker {
    font-size: 10px;
  }

  .brand-text h1 {
    font-size: 28px;
  }

  .brand-text p {
    font-size: 13px;
  }

  .app-header.compact-master .brand-card {
    padding: 14px;
  }

  .app-header.compact-master .left-area {
    flex-wrap: wrap;
  }

  .app-header.compact-master .brand-text {
    flex-basis: 100%;
  }
}
</style>