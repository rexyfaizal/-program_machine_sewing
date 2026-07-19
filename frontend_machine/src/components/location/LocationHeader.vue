<script setup>
import headerLogo from "../../assets/machine.png";
</script>

<template>
  <div class="location-header hero-header">
    <div class="hero-logo-box">
      <img :src="headerLogo" alt="Monitoring Machine Logo" class="hero-logo" />
    </div>

    <div class="hero-text">
      <div class="hero-kicker">MONITORING MACHINE</div>
      <h1>Monitoring Location Machine</h1>
    </div>
  </div>
</template>

<style scoped>
.location-header {
  margin-bottom: 18px;
}

.hero-header {
  display: flex;
  align-items: center;
  gap: 18px;
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 22px;
  padding: 18px 22px;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.04);
}

.hero-logo-box {
  width: 88px;
  height: 88px;
  border-radius: 18px;
  background: #f8fafc;
  border: 1px solid #dbe4ef;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.hero-logo {
  width: 68px;
  height: 68px;
  object-fit: contain;
}

.hero-text {
  min-width: 0;
}

.hero-kicker {
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: #7c3aed;
  margin-bottom: 6px;
}

.hero-text h1 {
  margin: 0;
  font-size: clamp(34px, 4vw, 58px);
  line-height: 1.05;
  letter-spacing: -0.04em;
  font-weight: 900;
  color: #0f172a;
}

@media (max-width: 768px) {
  .hero-header {
    align-items: flex-start;
    gap: 14px;
    padding: 16px;
  }

  .hero-logo-box {
    width: 72px;
    height: 72px;
  }

  .hero-logo {
    width: 56px;
    height: 56px;
  }

  .hero-text h1 {
    font-size: clamp(26px, 7vw, 40px);
  }
}
</style>