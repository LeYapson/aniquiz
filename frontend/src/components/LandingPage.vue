<template>
  <div class="landing">
    <div class="landing-content">
      <img :src="logo" alt="AniQuiz" class="landing-logo" />

      <p class="landing-tagline">
        Reconnais les génériques d'anime avant tes adversaires.<br />
        Prouve que tu es le vrai otaku.
      </p>

      <div class="landing-actions">
        <button @click="emit('play')" class="btn-play">Jouer maintenant</button>
        <button @click="emit('leaderboard')" class="btn-lb">🏆 Classement</button>
      </div>

      <div class="landing-features">
        <div class="feature-card">
          <span class="feat-icon">🎵</span>
          <strong>Blindtest</strong>
          <small>Identifie l'anime rien qu'à l'écoute</small>
        </div>
        <div class="feature-card">
          <span class="feat-icon">⚔️</span>
          <strong>Multijoueur</strong>
          <small>Affronte tes amis en temps réel</small>
        </div>
        <div class="feature-card">
          <span class="feat-icon">⭐</span>
          <strong>Progression</strong>
          <small>Gagne de l'XP et monte de niveau</small>
        </div>
        <div class="feature-card">
          <span class="feat-icon">📥</span>
          <strong>Communauté</strong>
          <small>Enrichis la bibliothèque toi-même</small>
        </div>
      </div>
    </div>

    <img :src="kora" alt="Kora" class="landing-kora" />
  </div>
</template>

<script setup>
const emit = defineEmits(['play', 'leaderboard']);
const logo = '/logo.png';
const kora = '/mascot_kora.png';
</script>

<style scoped>
.landing {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0f23 0%, #1a1a2e 50%, #16213e 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  position: relative;
  overflow: hidden;
}

/* Cercles décoratifs en arrière-plan */
.landing::before,
.landing::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  opacity: 0.06;
}
.landing::before {
  width: 600px; height: 600px;
  background: #f97316;
  top: -200px; right: -100px;
}
.landing::after {
  width: 400px; height: 400px;
  background: #3b82f6;
  bottom: -150px; left: -100px;
}

.landing-content {
  max-width: 600px;
  text-align: center;
  z-index: 1;
}

.landing-logo {
  width: 220px;
  margin-bottom: 24px;
  filter: drop-shadow(0 4px 24px rgba(249,115,22,0.3));
}

.landing-tagline {
  color: #cbd5e1;
  font-size: 1.1rem;
  line-height: 1.7;
  margin-bottom: 36px;
}

.landing-actions {
  display: flex;
  gap: 14px;
  justify-content: center;
  margin-bottom: 48px;
  flex-wrap: wrap;
}

.btn-play {
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: white;
  border: none;
  padding: 14px 36px;
  border-radius: 50px;
  font-size: 1.05rem;
  font-weight: bold;
  cursor: pointer;
  box-shadow: 0 4px 20px rgba(249,115,22,0.4);
  transition: transform 0.15s, box-shadow 0.15s;
}
.btn-play:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 28px rgba(249,115,22,0.5);
}

.btn-lb {
  background: transparent;
  color: #94a3b8;
  border: 1px solid #334155;
  padding: 14px 28px;
  border-radius: 50px;
  font-size: 1rem;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}
.btn-lb:hover {
  color: #f97316;
  border-color: #f97316;
}

.landing-features {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.feature-card {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 18px 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  transition: background 0.15s;
}
.feature-card:hover {
  background: rgba(249,115,22,0.08);
  border-color: rgba(249,115,22,0.2);
}

.feat-icon { font-size: 1.6rem; }
.feature-card strong { color: #f1f5f9; font-size: 0.95rem; }
.feature-card small  { color: #64748b; font-size: 0.8rem; text-align: center; }

/* Kora — positionnée à droite, partiellement hors-écran sur desktop */
.landing-kora {
  position: absolute;
  right: -10px;
  bottom: 0;
  height: 420px;
  z-index: 1;
  filter: drop-shadow(-8px 0 32px rgba(0,0,0,0.5));
  pointer-events: none;
}

@media (max-width: 900px) {
  .landing-kora { display: none; }
  .landing-features { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 500px) {
  .landing-logo { width: 160px; }
  .landing-tagline { font-size: 0.95rem; }
  .landing-features { grid-template-columns: 1fr; }
}
</style>
