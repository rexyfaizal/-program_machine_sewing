<script setup>
defineProps({
  machines: {
    type: Array,
    default: () => [],
  },
});

function pct(value) {
  return `${Number(value || 0).toFixed(2)}%`;
}
</script>

<template>
  <section class="panel attention-panel">
    <div class="panel-head">
      <div>
        <h3>Mesin Perlu Perhatian</h3>
      </div>
    </div>

    <div class="attention-list">
      <div
        v-for="(m, index) in machines"
        :key="m.uuid || index"
        class="attention-item"
      >
        <div class="attention-no">
          {{ index + 1 }}
        </div>

        <div class="item-info">
          <strong :title="m.machineName">
            {{ m.machineName || "-" }}
          </strong>

          <small>
            {{ m.ip || "-" }} · Output {{ m.output || 0 }}
          </small>
        </div>

        <div class="item-value">
          {{ pct(m.productivity) }}
        </div>
      </div>

      <div v-if="!machines.length" class="empty">
        Tidak ada data mesin.
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel {
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  border: 1px solid #dbeafe;
  border-radius: 24px;
  padding: 20px;
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.05);
}

.attention-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 18px;
}

.panel-head h3 {
  margin: 0 0 4px;
  font-size: 20px;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.panel-head p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.attention-list {
  display: grid;
  gap: 10px;
  flex: 1;
  align-content: start;
}

.attention-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) 78px;
  gap: 12px;
  align-items: center;
  min-height: 46px;
  padding: 10px 12px;
  border-radius: 16px;
  background: #fff7f7;
  border: 1px solid #fecaca;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background 0.18s ease;
}

.attention-item:hover {
  transform: translateY(-1px);
  background: #fff1f2;
  border-color: #fca5a5;
  box-shadow: 0 10px 20px rgba(239, 68, 68, 0.08);
}

.attention-no {
  width: 28px;
  height: 28px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: #fee2e2;
  color: #dc2626;
  border: 1px solid #fecaca;
  font-size: 13px;
  font-weight: 1000;
}

.item-info {
  min-width: 0;
}

.item-info strong {
  display: block;
  color: #0f172a;
  font-size: 13px;
  font-weight: 1000;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-info small {
  display: block;
  margin-top: 4px;
  color: #64748b;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-value {
  color: #dc2626;
  font-size: 13px;
  font-weight: 1000;
  text-align: right;
  white-space: nowrap;
}

.empty {
  padding: 26px;
  text-align: center;
  color: #64748b;
  font-weight: 800;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 16px;
}

@media (max-width: 640px) {
  .attention-item {
    grid-template-columns: 30px minmax(0, 1fr);
  }

  .item-value {
    grid-column: 2 / -1;
    text-align: left;
  }
}
</style>