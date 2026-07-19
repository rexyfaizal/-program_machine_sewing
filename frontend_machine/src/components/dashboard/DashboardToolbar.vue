<script setup>
const props = defineProps({
  selectedDate: {
    type: String,
    required: true,
  },
  keyword: {
    type: String,
    required: true,
  },
});

const emit = defineEmits([
  "update:selectedDate",
  "update:keyword",
  "dateChange",
  "download",
]);
</script>

<template>
  <section class="toolbar-card">
    <div class="toolbar-left">
      <label class="date-box">
        <span class="box-label">Tanggal</span>

        <input
          type="date"
          :value="props.selectedDate"
          @input="emit('update:selectedDate', $event.target.value)"
          @change="emit('dateChange')"
        />
      </label>

      <label class="search-box">
        <span class="box-label">Pencarian</span>

        <input
          type="text"
          :value="props.keyword"
          placeholder="Cari mesin, IP, UUID, status, atau program jahit..."
          @input="emit('update:keyword', $event.target.value)"
        />
      </label>
    </div>

    <div class="toolbar-right">
      <button type="button" class="download-btn" @click="emit('download')">
        Download CSV
      </button>
    </div>
  </section>
</template>

<style scoped>
.toolbar-card {
  width: 100%;
  min-width: 0;
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: stretch;
  padding: clamp(14px, 1.3vw, 18px);
  border-radius: 24px;
  background: linear-gradient(135deg, #ffffff, #f3f8ff);
  border: 1px solid #dbeafe;
  box-shadow:
    0 16px 30px rgba(37, 99, 235, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  overflow: hidden;
}

.toolbar-left {
  display: grid;
  grid-template-columns: minmax(170px, 220px) minmax(0, 1fr);
  gap: 14px;
  flex: 1;
  min-width: 0;
}

.date-box,
.search-box {
  min-width: 0;
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  border-radius: 18px;
  background: #ffffff;
  border: 1px solid #dbeafe;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.04);
}

.box-label {
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #64748b;
  white-space: nowrap;
}

.date-box input,
.search-box input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: none;
  background: transparent;
  font-size: clamp(13px, 0.9vw, 15px);
  font-weight: 800;
  color: #0f172a;
  padding: 0;
  height: 26px;
}

.search-box input::placeholder {
  color: #94a3b8;
  font-weight: 700;
}

.toolbar-right {
  display: flex;
  align-items: stretch;
  flex: 0 0 auto;
}

.download-btn {
  min-width: 150px;
  border: 0;
  border-radius: 18px;
  padding: 0 22px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  font-weight: 900;
  font-size: 14px;
  cursor: pointer;
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.24);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    filter 0.2s ease;
  white-space: nowrap;
}

.download-btn:hover {
  transform: translateY(-2px);
  filter: brightness(1.03);
  box-shadow: 0 18px 30px rgba(37, 99, 235, 0.3);
}

.download-btn:active {
  transform: translateY(0);
}

@media (max-width: 1200px) {
  .toolbar-card {
    flex-direction: column;
  }

  .toolbar-left {
    grid-template-columns: minmax(160px, 220px) minmax(0, 1fr);
  }

  .toolbar-right {
    justify-content: flex-end;
  }

  .download-btn {
    min-height: 52px;
  }
}

@media (max-width: 760px) {
  .toolbar-left {
    grid-template-columns: 1fr;
  }

  .toolbar-right {
    width: 100%;
  }

  .download-btn {
    width: 100%;
    min-height: 50px;
  }
}

@media (max-width: 520px) {
  .toolbar-card {
    border-radius: 20px;
    padding: 12px;
    gap: 12px;
  }

  .date-box,
  .search-box {
    border-radius: 16px;
    padding: 11px 12px;
  }
}
</style>