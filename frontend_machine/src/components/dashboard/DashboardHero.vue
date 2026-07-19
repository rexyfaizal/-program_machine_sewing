<script setup>
import heroLogo from "../../assets/machine.png";
// kalau file Anda png, ganti jadi:
// import heroLogo from "../../assets/monitoring-machine.png";
</script>

<template>
  <section class="hero">
    <div class="hero-main">
      <div class="hero-header">
        <div class="logo-box">
          <img :src="heroLogo" alt="Monitoring Machine" />
        </div>

        <div class="hero-text">
          <div class="hero-label">MONITORING MACHINE</div>

          <h1>Monitoring Productivity Machine</h1>

        </div>
      </div>
    </div>

    <div class="hero-side">

      <div class="status-rule">
        <span class="rule good">GOOD = 90%</span>
        <span class="rule normal">NORMAL 80–90%</span>
        <span class="rule bad">BAD &lt; 80%</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.8fr) minmax(320px, 0.9fr);
  gap: 18px;
  margin-bottom: 18px;
}

.hero-main {
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 20px;
  padding: 20px 22px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
  min-height: 160px;
  display: flex;
  align-items: center;
}

.hero-header {
  display: flex;
  align-items: center;
  gap: 18px;
  width: 100%;
}

.logo-box {
  width: 92px;
  height: 92px;
  min-width: 92px;
  border-radius: 18px;
  background: #f8fafc;
  border: 1px solid #dbe4ef;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
}

.logo-box img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
}

.hero-text {
  min-width: 0;
}

.hero-label {
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.12em;
  color: #6d28d9;
  text-transform: uppercase;
}

.hero-text h1 {
  margin: 0 0 10px;
  color: #0f172a;
  font-size: clamp(28px, 3vw, 44px);
  line-height: 1.1;
  letter-spacing: -0.04em;
  font-weight: 900;
}

.hero-text p {
  margin: 0;
  color: #475569;
  font-size: 15px;
  line-height: 1.7;
  max-width: 960px;
}

.hero-text b {
  color: #0f172a;
  font-weight: 900;
}

.hero-side {
  background: #ffffff;
  border: 1px solid #dbe4ef;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
}

.hero-side h3 {
  margin: 0 0 14px;
  font-size: 15px;
  color: #0f172a;
  font-weight: 900;
  text-transform: uppercase;
}

.formula-line {
  border-left: 4px solid #2563eb;
  padding-left: 12px;
  color: #475569;
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 18px;
}

.status-rule {
  display: grid;
  gap: 10px;
}

.rule {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: fit-content;
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 900;
}

.good {
  background: #dcfce7;
  color: #166534;
}

.normal {
  background: #fef3c7;
  color: #92400e;
}

.bad {
  background: #fee2e2;
  color: #991b1b;
}

@media (max-width: 1200px) {
  .hero {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .hero-main {
    padding: 18px;
  }

  .hero-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .logo-box {
    width: 78px;
    height: 78px;
    min-width: 78px;
  }

  .hero-text h1 {
    font-size: 30px;
  }
}
</style>