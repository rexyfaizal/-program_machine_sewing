<script setup>
const emit = defineEmits(["addLine"]);
</script>

<template>
  <div class="add-line-card" @click="emit('addLine')">
    <div class="plus-icon">+</div>
    <strong>Tambah Line</strong>
    <small>Klik untuk menambahkan line baru</small>
  </div>
</template>

<style scoped>
.add-line-card {
  background: #ffffff;
  border: 2px dashed #cbd5e1;
  border-radius: 12px;
  min-height: 520px;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 10px;
  cursor: pointer;
  color: #64748b;
  transition:
    background 0.2s ease,
    color 0.2s ease,
    border-color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.add-line-card:hover {
  background: #2563eb;
  color: #ffffff;
  border-color: #2563eb;
  transform: translateY(-3px);
  box-shadow: 0 14px 32px rgba(37, 99, 235, 0.24);
}

.plus-icon {
  width: 54px;
  height: 54px;
  border-radius: 999px;
  background: #eff6ff;
  color: #1d4ed8;
  display: grid;
  place-items: center;
  font-size: 34px;
  line-height: 1;
  font-weight: 900;
}

.add-line-card:hover .plus-icon {
  background: rgba(255, 255, 255, 0.16);
  color: #ffffff;
}

.add-line-card strong {
  font-size: 14px;
  font-weight: 900;
}

.add-line-card small {
  max-width: 120px;
  text-align: center;
  font-size: 11px;
  line-height: 1.4;
  color: #94a3b8;
}

.add-line-card:hover small {
  color: rgba(255, 255, 255, 0.78);
}
</style>