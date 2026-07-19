<script setup>
defineProps({
  selectedFactory: {
    type: String,
    default: "-",
  },
  lineCount: {
    type: Number,
    default: 0,
  },
  assignedCount: {
    type: Number,
    default: 0,
  },
});
</script>

<template>
  <section class="info-grid">
    <article class="info-card primary">
      <div class="card-kicker">Factory</div>
      <div class="card-value">{{ selectedFactory || "-" }}</div>
    </article>

    <article class="info-card">
      <div class="card-kicker">Jumlah Line</div>
      <div class="card-value">{{ lineCount }}</div>
    </article>

    <article class="info-card">
      <div class="card-kicker">Mesin Terpasang</div>
      <div class="card-value">{{ assignedCount }}</div>
    </article>
  </section>
</template>

<style scoped>
.info-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.info-card {
  position: relative;
  overflow: hidden;
  min-height: 88px;
  padding: 14px 16px;
  border-radius: 18px;
  border: 1px solid #dbeafe;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
  box-shadow:
    0 10px 20px rgba(15, 23, 42, 0.045),
    inset 0 1px 0 rgba(255, 255, 255, 0.85);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.info-card::after {
  content: "";
  position: absolute;
  right: -22px;
  top: -22px;
  width: 84px;
  height: 84px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.06);
}

.info-card.primary {
  background:
    linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border-color: #93c5fd;
}

.card-kicker {
  position: relative;
  z-index: 1;
  color: #64748b;
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.10em;
  margin-bottom: 8px;
}

.card-value {
  position: relative;
  z-index: 1;
  color: #0f172a;
  font-size: 30px;
  line-height: 1;
  font-weight: 950;
  letter-spacing: -0.04em;
}

.primary .card-value {
  color: #1d4ed8;
}

@media (max-width: 980px) {
  .info-grid {
    grid-template-columns: 1fr;
  }

  .card-value {
    font-size: 28px;
  }
}
</style>