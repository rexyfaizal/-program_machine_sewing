<script setup>
import { ref, watch, onMounted } from "vue";

import GlobalStatusDots from "./components/GlobalStatusDots.vue";
import AppHeader from "./components/AppHeader.vue";
import DashboardPage from "./pages/DashboardPage.vue";
import ProcessDetailPage from "./pages/ProcessDetailPage.vue";
import LocationTemplatePage from "./pages/LocationTemplatePage.vue";
import OperatorPicSpvPage from "./pages/OperatorPicSpvPage.vue";
import { getInitialAdminMode } from "./utils/adminMode";
import ProcessStyleMasterPage from "./pages/ProcessStyleMasterPage.vue";

const STORAGE_KEY = "machineDashboardActivePage";

const activePage = ref("dashboard");
const selectedDate = ref(todayLocal());

const globalSocketStatus = ref("offline");
const globalIsAdmin = ref(false);
const operatorUuid = ref("");

function todayLocal() {
  const d = new Date();
  const offset = d.getTimezoneOffset();
  const local = new Date(d.getTime() - offset * 60000);
  return local.toISOString().slice(0, 10);
}

function readUrlRoute() {
  const params = new URLSearchParams(window.location.search);
  const page = params.get("page");
  const uuid = params.get("uuid");

  if (page === "operator-pic-spv" && uuid) {
    operatorUuid.value = uuid;
    activePage.value = "operator-pic-spv";
    return true;
  }

  return false;
}

function clearOperatorUrl() {
  const params = new URLSearchParams(window.location.search);

  params.delete("page");
  params.delete("uuid");

  const query = params.toString();
  const nextUrl = query
    ? `${window.location.pathname}?${query}`
    : window.location.pathname;

  window.history.replaceState({}, "", nextUrl);
}

function changePage(page) {
  activePage.value = page;

  if (page !== "operator-pic-spv") {
    operatorUuid.value = "";
    localStorage.setItem(STORAGE_KEY, page);
    clearOperatorUrl();
  }
}

function backToDashboard() {
  changePage("dashboard");
}

function openMachineDetail(machine) {
  if (!machine?.uuid) return;

  localStorage.setItem("machineDashboardDetailUuid", machine.uuid);
  localStorage.setItem(
    "machineDashboardDetailMachineName",
    machine.machineName || machine.nickName || ""
  );

  changePage("detail");
}

function updateGlobalSocketStatus(status) {
  globalSocketStatus.value = status || "offline";
}

function updateGlobalAdminMode(value) {
  globalIsAdmin.value = Boolean(value);
}

onMounted(async () => {
  globalIsAdmin.value = await Promise.resolve(getInitialAdminMode());

  const isOperatorRoute = readUrlRoute();

  if (isOperatorRoute) {
    return;
  }

  const savedPage = localStorage.getItem(STORAGE_KEY);

  if (["dashboard", "detail", "location", "master-ie"].includes(savedPage)) {
    activePage.value = savedPage;
  }
});

watch(activePage, (newPage) => {
  if (["dashboard", "detail", "location", "master-ie"].includes(newPage)) {
    localStorage.setItem(STORAGE_KEY, newPage);
  }
});
</script>

<template>
  <OperatorPicSpvPage
    v-if="activePage === 'operator-pic-spv'"
    :uuid="operatorUuid"
    @back-dashboard="backToDashboard"
  />

  <main v-else class="page">
    <GlobalStatusDots
      :socket-status="globalSocketStatus"
      :is-admin="globalIsAdmin"
    />

    <div class="shell">
      <AppHeader v-model="activePage" />

      <DashboardPage
        v-if="activePage === 'dashboard'"
        v-model:selectedDate="selectedDate"
        @socket-status-change="updateGlobalSocketStatus"
        @admin-mode-change="updateGlobalAdminMode"
      />

      <ProcessDetailPage
        v-else-if="activePage === 'detail'"
        v-model:selectedDate="selectedDate"
      />

      <LocationTemplatePage
        v-else-if="activePage === 'location'"
        @open-detail-machine="openMachineDetail"
      />
      
      <ProcessStyleMasterPage
        v-else-if="activePage === 'master-ie'"
      />
    </div>
  </main>
</template>

<style>
* {
  box-sizing: border-box;
}

html,
body,
#app {
  width: 100%;
  min-height: 100%;
  overflow-x: hidden;
}

body {
  margin: 0;
  background: #eaf2fb;
  color: #0f172a;
  font-family:
    "Inter",
    "Plus Jakarta Sans",
    ui-sans-serif,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    sans-serif;
}

.page {
  --app-scale: 1;

  min-height: 100vh;
  width: 100%;
  padding: clamp(14px, 1.6vw, 28px);
  padding-top: clamp(42px, 4vw, 52px);
  background:
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.08), transparent 34%),
    #eaf2fb;
  overflow-x: hidden;
}

.shell {
  width: 100%;
  max-width: 1680px;
  margin: 0 auto;
  min-width: 0;
  zoom: var(--app-scale);
}

.shell > * {
  min-width: 0;
}

button,
input,
select {
  font-family: inherit;
}

img {
  max-width: 100%;
}

/* Monitor besar */
@media (min-width: 1701px) {
  .page {
    --app-scale: 1;
  }
}

/* Desktop / monitor 1600px */
@media (max-width: 1700px) {
  .page {
    --app-scale: 0.96;
  }
}

/* Laptop 1440px */
@media (max-width: 1500px) {
  .page {
    --app-scale: 0.92;
  }
}

/* Laptop 1366px */
@media (max-width: 1400px) {
  .page {
    --app-scale: 0.88;
  }
}

/* Laptop kecil */
@media (max-width: 1280px) {
  .page {
    --app-scale: 0.84;
  }
}

/*
  Untuk tablet/mobile jangan diperkecil pakai zoom.
  Biarkan responsive CSS masing-masing komponen yang bekerja.
*/
@media (max-width: 1100px) {
  .page {
    --app-scale: 1;
    padding: 42px 14px 18px;
  }

  .shell {
    max-width: 100%;
  }
}

@media (max-width: 600px) {
  .page {
    padding: 40px 10px 14px;
  }
}
</style>