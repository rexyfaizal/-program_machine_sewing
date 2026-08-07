<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

defineProps({
  modelValue: {
    type: String,
    default: "dashboard",
  },
});

const emit = defineEmits(["update:modelValue"]);

const isOpen = ref(false);
const menuRef = ref(null);

function toggleMenu() {
  isOpen.value = !isOpen.value;
}

function closeMenu() {
  isOpen.value = false;
}

function selectMenu(page) {
  emit("update:modelValue", page);
  closeMenu();
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
    </div>
  </div>
</template>

<style scoped>
.top-menu {
  position: relative;
  z-index: 40;
}

.hamburger-btn {
  width: 42px;
  height: 42px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.92);
  display: grid;
  place-content: center;
  gap: 5px;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.hamburger-btn span {
  display: block;
  width: 16px;
  height: 2px;
  border-radius: 999px;
  background: #0f172a;
}

.menu-dropdown {
  position: absolute;
  top: calc(100% + 10px);
  left: 0;
  min-width: 280px;
  padding: 8px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid rgba(191, 219, 254, 0.9);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.14);
  display: grid;
  gap: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  border: 0;
  background: transparent;
  text-align: left;
  padding: 12px 12px;
  border-radius: 14px;
  cursor: pointer;
  color: #0f172a;
}

.menu-item:hover,
.menu-item.active {
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
}

.menu-item strong {
  display: block;
  font-size: 0.92rem;
  font-weight: 800;
}

.menu-item small {
  display: block;
  margin-top: 2px;
  color: #64748b;
  font-size: 0.75rem;
}

.menu-icon {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.menu-icon.master {
  background: #eff6ff;
  border-color: #bfdbfe;
}

.icon-grid,
.icon-list,
.icon-location,
.icon-master {
  position: relative;
  width: 16px;
  height: 16px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 3px;
}

.icon-grid i {
  width: 6px;
  height: 6px;
  border-radius: 2px;
  background: #2563eb;
}

.icon-list {
  display: grid;
  gap: 3px;
  align-content: center;
}

.icon-list i {
  height: 2px;
  border-radius: 999px;
  background: #0f172a;
}

.icon-location i:nth-child(1) {
  position: absolute;
  inset: 1px 4px 4px;
  border: 2px solid #ef4444;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
}

.icon-location i:nth-child(2) {
  position: absolute;
  left: 50%;
  top: 42%;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #ef4444;
  transform: translate(-50%, -50%);
}

.icon-location i:nth-child(3) {
  display: none;
}

.icon-master {
  display: grid;
  gap: 3px;
  align-content: center;
}

.icon-master i {
  height: 2px;
  border-radius: 999px;
  background: #2563eb;
}
</style>
