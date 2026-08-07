<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import {
  isMechanicUnlocked,
  unlockMechanicAccess,
} from "../utils/mechanicAccess";

const props = defineProps({
  modelValue: {
    type: String,
    default: "dashboard",
  },
});

const emit = defineEmits(["update:modelValue"]);

const isOpen = ref(false);
const menuRef = ref(null);

const passwordModalOpen = ref(false);
const passwordInput = ref("");
const passwordError = ref("");
const passwordInputRef = ref(null);

function toggleMenu() {
  isOpen.value = !isOpen.value;
}

function closeMenu() {
  isOpen.value = false;
}

function closePasswordModal() {
  passwordModalOpen.value = false;
  passwordInput.value = "";
  passwordError.value = "";
}

async function openPasswordModal() {
  passwordModalOpen.value = true;
  passwordInput.value = "";
  passwordError.value = "";
  closeMenu();

  await nextTick();
  passwordInputRef.value?.focus?.();
}

function selectMenu(page) {
  if (page === "mechanic" && !isMechanicUnlocked()) {
    openPasswordModal();
    return;
  }

  emit("update:modelValue", page);
  closeMenu();
}

function submitPassword() {
  if (!unlockMechanicAccess(passwordInput.value)) {
    passwordError.value = "Password salah";
    passwordInput.value = "";
    passwordInputRef.value?.focus?.();
    return;
  }

  closePasswordModal();
  emit("update:modelValue", "mechanic");
}

function handleClickOutside(event) {
  if (!menuRef.value) return;

  if (!menuRef.value.contains(event.target)) {
    closeMenu();
  }
}

function handleEsc(event) {
  if (event.key === "Escape") {
    closeMenu();
    if (passwordModalOpen.value) {
      closePasswordModal();
    }
  }
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  document.addEventListener("keydown", handleEsc);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
  document.removeEventListener("keydown", handleEsc);
});
</script>

<template>
  <div class="top-menu" ref="menuRef">
    <button class="hamburger-btn" @click.stop="toggleMenu" title="Menu">
      <span></span>
      <span></span>
      <span></span>
    </button>

    <div v-if="isOpen" class="menu-dropdown">
      <!-- Dashboard Utama -->
      <button
        class="menu-item"
        :class="{ active: modelValue === 'dashboard' }"
        @click="selectMenu('dashboard')"
      >
        <span class="menu-icon">
          <span class="icon-grid">
            <i></i>
            <i></i>
            <i></i>
            <i></i>
          </span>
        </span>

        <span>
          <strong>Dashboard Utama</strong>
          <small>Ringkasan produktivitas mesin</small>
        </span>
      </button>

      <!-- Detail Proses Mesin -->
      <button
        class="menu-item"
        :class="{ active: modelValue === 'detail' }"
        @click="selectMenu('detail')"
      >
        <span class="menu-icon">
          <span class="icon-list">
            <i></i>
            <i></i>
            <i></i>
          </span>
        </span>

        <span>
          <strong>Detail Proses Mesin</strong>
          <small>Detail proses per mesin</small>
        </span>
      </button>

      <!-- Location Template -->
      <button
        class="menu-item"
        :class="{ active: modelValue === 'location' }"
        @click="selectMenu('location')"
      >
        <span class="menu-icon">
          <span class="icon-location">
            <i></i>
            <i></i>
            <i></i>
          </span>
        </span>

        <span>
          <strong>Location Template</strong>
          <small>Layout lokasi mesin per line</small>
        </span>
      </button>

      <!-- Master IE -->
      <button
        class="menu-item"
        :class="{ active: modelValue === 'master-ie' }"
        @click="selectMenu('master-ie')"
      >
        <span class="menu-icon master">
          <span class="icon-master">
            <i></i>
            <i></i>
            <i></i>
          </span>
        </span>

        <span>
          <strong>Master IE</strong>
          <small>Input style dan proses</small>
        </span>
      </button>

      <!-- Dashboard Mekanik -->
      <button
        class="menu-item"
        :class="{ active: modelValue === 'mechanic' }"
        @click="selectMenu('mechanic')"
      >
        <span class="menu-icon mechanic">
          <span class="icon-mechanic">
            <i></i>
            <i></i>
            <i></i>
          </span>
        </span>

        <span>
          <strong>Dashboard Mekanik</strong>
          <small>Mesin rusak · Ambil · Selesai</small>
        </span>
      </button>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="passwordModalOpen"
      class="password-backdrop"
      @click.self="closePasswordModal"
    >
      <form class="password-modal" @submit.prevent="submitPassword">
        <p class="password-kicker">DASHBOARD MEKANIK</p>
        <h3>Masukkan Password</h3>
        <p class="password-hint">Akses menu mekanik membutuhkan password.</p>

        <label>
          Password
          <input
            ref="passwordInputRef"
            v-model="passwordInput"
            type="password"
            placeholder="Password"
            autocomplete="current-password"
          />
        </label>

        <p v-if="passwordError" class="password-error">{{ passwordError }}</p>

        <div class="password-actions">
          <button type="button" class="pwd-btn ghost" @click="closePasswordModal">
            Batal
          </button>
          <button type="submit" class="pwd-btn primary">Masuk</button>
        </div>
      </form>
    </div>
  </Teleport>
</template>

<style scoped>
.top-menu {
  position: relative;
  display: inline-block;
  margin-bottom: 18px;
  z-index: 50;
}

.hamburger-btn {
  width: 46px;
  height: 42px;
  border: 1px solid #d8e2ee;
  background: #ffffff;
  border-radius: 12px;
  display: grid;
  place-items: center;
  gap: 4px;
  padding: 9px;
  cursor: pointer;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.06);
}

.hamburger-btn span {
  display: block;
  width: 20px;
  height: 2px;
  background: #0f172a;
  border-radius: 999px;
}

.hamburger-btn:hover {
  background: #f8fafc;
}

.menu-dropdown {
  position: absolute;
  top: 52px;
  left: 0;
  width: 330px;
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
  padding: 10px;
  box-shadow: 0 18px 45px rgba(15, 23, 42, 0.18);
}

.menu-dropdown::before {
  content: "";
  position: absolute;
  top: -7px;
  left: 17px;
  width: 14px;
  height: 14px;
  background: #ffffff;
  border-left: 1px solid #dbe4ef;
  border-top: 1px solid #dbe4ef;
  transform: rotate(45deg);
}

.menu-item {
  position: relative;
  width: 100%;
  border: 0;
  background: transparent;
  display: grid;
  grid-template-columns: 36px 1fr;
  gap: 10px;
  align-items: center;
  text-align: left;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  color: #0f172a;
  transition:
    background 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.menu-item:hover {
  background: #1d4ed8;
  color: white;
}

.menu-item.active {
  background: #2563eb;
  color: white;
}

.menu-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: #eaf1ff;
}

.menu-icon.master {
  background: #dcfce7;
}

.menu-icon.mechanic {
  background: #ffedd5;
}

.menu-item:hover .menu-icon,
.menu-item.active .menu-icon {
  background: rgba(255, 255, 255, 0.16);
}

/* Icon Dashboard */
.icon-grid {
  display: grid;
  grid-template-columns: repeat(2, 7px);
  grid-template-rows: repeat(2, 7px);
  gap: 3px;
}

.icon-grid i {
  display: block;
  width: 7px;
  height: 7px;
  border-radius: 2px;
  background: #1d4ed8;
}

/* Icon Detail */
.icon-list {
  display: grid;
  gap: 4px;
}

.icon-list i {
  display: block;
  width: 18px;
  height: 3px;
  border-radius: 999px;
  background: #1d4ed8;
}

/* Icon Location */
.icon-location {
  position: relative;
  width: 20px;
  height: 20px;
  display: block;
}

.icon-location i:nth-child(1) {
  position: absolute;
  left: 3px;
  top: 3px;
  width: 14px;
  height: 14px;
  border: 2px solid #1d4ed8;
  border-radius: 5px;
}

.icon-location i:nth-child(2) {
  position: absolute;
  left: 8px;
  top: 8px;
  width: 4px;
  height: 4px;
  background: #1d4ed8;
  border-radius: 999px;
}

.icon-location i:nth-child(3) {
  position: absolute;
  left: 9px;
  top: 14px;
  width: 2px;
  height: 5px;
  background: #1d4ed8;
  border-radius: 999px;
}

/* Icon Master IE */
.icon-master {
  position: relative;
  width: 20px;
  height: 20px;
  display: block;
}

.icon-master i:nth-child(1) {
  position: absolute;
  left: 2px;
  top: 3px;
  width: 16px;
  height: 14px;
  border: 2px solid #16a34a;
  border-radius: 4px;
}

.icon-master i:nth-child(2) {
  position: absolute;
  left: 6px;
  top: 7px;
  width: 8px;
  height: 2px;
  background: #16a34a;
  border-radius: 999px;
}

.icon-master i:nth-child(3) {
  position: absolute;
  left: 6px;
  top: 12px;
  width: 8px;
  height: 2px;
  background: #16a34a;
  border-radius: 999px;
}

.menu-item:hover .icon-grid i,
.menu-item.active .icon-grid i,
.menu-item:hover .icon-list i,
.menu-item.active .icon-list i {
  background: #ffffff;
}

.menu-item:hover .icon-location i:nth-child(1),
.menu-item.active .icon-location i:nth-child(1) {
  border-color: #ffffff;
}

.menu-item:hover .icon-location i:nth-child(2),
.menu-item.active .icon-location i:nth-child(2),
.menu-item:hover .icon-location i:nth-child(3),
.menu-item.active .icon-location i:nth-child(3) {
  background: #ffffff;
}

.menu-item:hover .icon-master i:nth-child(1),
.menu-item.active .icon-master i:nth-child(1) {
  border-color: #ffffff;
}

.menu-item:hover .icon-master i:nth-child(2),
.menu-item.active .icon-master i:nth-child(2),
.menu-item:hover .icon-master i:nth-child(3),
.menu-item.active .icon-master i:nth-child(3) {
  background: #ffffff;
}

/* Icon Mekanik */
.icon-mechanic {
  position: relative;
  width: 20px;
  height: 20px;
  display: block;
}

.icon-mechanic i:nth-child(1) {
  position: absolute;
  left: 3px;
  top: 2px;
  width: 14px;
  height: 10px;
  border: 2px solid #ea580c;
  border-radius: 3px;
}

.icon-mechanic i:nth-child(2) {
  position: absolute;
  left: 8px;
  top: 11px;
  width: 4px;
  height: 6px;
  background: #ea580c;
  border-radius: 1px;
}

.icon-mechanic i:nth-child(3) {
  position: absolute;
  left: 5px;
  top: 16px;
  width: 10px;
  height: 2px;
  background: #ea580c;
  border-radius: 999px;
}

.menu-item:hover .icon-mechanic i:nth-child(1),
.menu-item.active .icon-mechanic i:nth-child(1) {
  border-color: #ffffff;
}

.menu-item:hover .icon-mechanic i:nth-child(2),
.menu-item.active .icon-mechanic i:nth-child(2),
.menu-item:hover .icon-mechanic i:nth-child(3),
.menu-item.active .icon-mechanic i:nth-child(3) {
  background: #ffffff;
}

.menu-item strong {
  display: block;
  font-size: 13px;
  font-weight: 900;
}

.menu-item small {
  display: block;
  font-size: 11px;
  margin-top: 3px;
  color: #64748b;
}

.menu-item:hover small,
.menu-item.active small {
  color: rgba(255, 255, 255, 0.75);
}

@media (max-width: 760px) {
  .menu-dropdown {
    width: 280px;
  }
}

.password-backdrop {
  position: fixed;
  inset: 0;
  z-index: 200;
  background: rgba(15, 23, 42, 0.45);
  display: grid;
  place-items: center;
  padding: 16px;
}

.password-modal {
  width: min(400px, 100%);
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  padding: 18px;
  display: grid;
  gap: 10px;
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.25);
}

.password-kicker {
  margin: 0;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  color: #c2410c;
}

.password-modal h3 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
}

.password-hint {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.password-modal label {
  display: grid;
  gap: 6px;
  font-size: 12px;
  font-weight: 800;
  color: #475569;
}

.password-modal input {
  height: 44px;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 0 12px;
  font-size: 15px;
}

.password-error {
  margin: 0;
  color: #b91c1c;
  font-size: 13px;
  font-weight: 700;
}

.password-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.pwd-btn {
  height: 38px;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: 0 14px;
  font-weight: 800;
  cursor: pointer;
}

.pwd-btn.primary {
  background: #2563eb;
  color: #fff;
}

.pwd-btn.ghost {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}
</style>