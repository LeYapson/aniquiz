<template>
  <div class="landing">

    <!-- ═══════════════════════════════════════════ HERO ══ -->
    <section class="hero">
      <div class="hero-particles" aria-hidden="true">
        <span>♪</span><span>♫</span><span>♩</span><span>♪</span>
        <span>♬</span><span>♫</span><span>♩</span>
      </div>
      <div class="hero-content">
        <img :src="img.logo" alt="AniQuiz" class="hero-logo" />
        <h1 class="hero-h1">Blindtest d'anime en ligne gratuit</h1>
        <p class="hero-tagline">
          Reconnais les génériques d'anime avant tes adversaires.<br />
          En multi, en solo ou en Speed Run — prouve que tu es le vrai otaku.
        </p>
        <div class="hero-actions">
          <button @click="emit('play')" class="btn-play">Jouer maintenant</button>
          <button @click="emit('leaderboard')" class="btn-lb"><span aria-hidden="true">🏆</span> Classement</button>
        </div>
        <div class="hero-stats">
          <div class="hero-stat hero-stat--live" v-if="playersOnline !== null">
            <strong>
              <span class="live-dot" aria-hidden="true"></span>{{ playersOnline }}
            </strong>
            <span>en ligne</span>
          </div>
          <div class="hero-stat-sep" v-if="playersOnline !== null"></div>
          <div class="hero-stat"><strong>500+</strong><span>animes</span></div>
          <div class="hero-stat-sep"></div>
          <div class="hero-stat"><strong>3 modes</strong><span>de jeu</span></div>
          <div class="hero-stat-sep"></div>
          <div class="hero-stat"><strong>Gratuit</strong><span>sans pub</span></div>
        </div>
      </div>
      <img :src="img.mascot" alt="Kora" class="hero-kora" />
      <div class="hero-scroll-hint" aria-hidden="true">
        <span class="scroll-arrow"></span>
      </div>
    </section>

    <!-- ═══════════════════════════════════ COMMENT ÇA MARCHE ══ -->
    <section class="section howto-section">
      <h2 class="section-title">Comment ça marche ?</h2>
      <p class="section-sub">Trois étapes, zéro prise de tête</p>
      <div class="howto-steps">
        <template v-for="(step, i) in steps" :key="step.num">
          <div class="step reveal" :style="{ transitionDelay: (i * 0.15) + 's' }">
            <div class="step-num">{{ step.num }}</div>
            <h3>{{ step.title }}</h3>
            <p>{{ step.body }}</p>
          </div>
          <div v-if="i < steps.length - 1" class="step-arrow" aria-hidden="true"></div>
        </template>
      </div>
    </section>

    <!-- ══════════════════════════════════════ MODES DE JEU ══ -->
    <section class="section modes-section">
      <h2 class="section-title">Modes de jeu</h2>
      <p class="section-sub">Plusieurs façons de tester ta culture anime</p>

      <div class="modes-grid">
        <div
          v-for="(mode, idx) in modes"
          :key="mode.title"
          class="mode-card reveal"
          :class="[`mode-card--${mode.key}`, { 'mode-card--soon': mode.soon }]"
          :style="{ transitionDelay: (idx * 0.12) + 's' }"
        >
          <div class="mode-card-header">
            <span class="mode-card-icon" aria-hidden="true">{{ mode.icon }}</span>
          </div>
          <div class="mode-card-body">
            <h3>{{ mode.title }}</h3>
            <p>{{ mode.body }}</p>
            <span class="mode-badge" :class="mode.soon ? 'mode-soon' : 'mode-available'">
              {{ mode.soon ? 'Bientôt' : 'Disponible' }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════════════════════════════════ QUIZ DU JOUR ══ -->
    <section class="section daily-section">
      <div class="daily-inner">
        <div class="daily-text reveal--left">
          <span class="daily-badge">📅 Nouveau</span>
          <h2 class="section-title left">Quiz du jour</h2>
          <p class="daily-desc">
            Chaque jour à minuit, une nouvelle piste est tirée pour tout le monde.
            Même extrait, même règles — seul ton temps de réponse fait la différence.
          </p>
          <ul class="daily-features">
            <li v-for="feat in dailyFeatures" :key="feat"><span class="rf-dot"></span>{{ feat }}</li>
          </ul>
          <button class="btn-play daily-cta" @click="emit('play')">Essayer le quiz →</button>
        </div>
        <div class="daily-visual reveal--right">
          <img :src="img.koraDaily" alt="Kora et le quiz du jour" class="kora-daily" />
          <div class="daily-stats-pills">
            <div class="daily-pill">📅 1 quiz / jour</div>
            <div class="daily-pill">⏱ Classé par temps</div>
            <div class="daily-pill">🔄 Reset minuit UTC</div>
          </div>
        </div>
      </div>
    </section>

    <!-- ═══════════════════════════════════ PROGRESSION & SOCIAL ══ -->
    <section class="section progression-section">
      <h2 class="section-title">Progresse et personnalise</h2>
      <p class="section-sub">Chaque partie contribue à ta progression permanente</p>
      <div class="progression-grid">
        <div class="prog-card reveal" v-for="(feat, idx) in progressionFeats" :key="feat.icon" :style="{ transitionDelay: (idx * 0.15) + 's' }">
          <div class="prog-icon" aria-hidden="true">{{ feat.icon }}</div>
          <h3>{{ feat.title }}</h3>
          <p>{{ feat.body }}</p>
        </div>
      </div>
    </section>

    <!-- ═══════════════════════════════════════ RANKED / DIVISIONS ══ -->
    <section class="section ranked-section">
      <div class="ranked-inner">
        <div class="ranked-text reveal--left">
          <h2 class="section-title left">Système de divisions</h2>
          <p class="ranked-desc">
            Chaque partie classée te rapporte des points de ranking. Monte les divisions,
            débloques des récompenses exclusives et défends ta place en fin de saison.
          </p>
          <ul class="ranked-features">
            <li v-for="feat in rankedFeatures" :key="feat"><span class="rf-dot"></span>{{ feat }}</li>
          </ul>
          <span class="coming-soon-pill">Bientôt disponible</span>
        </div>

        <div class="ranked-visual reveal--right">
          <img :src="img.koraProf" alt="Kora explique le système ranked" class="kora-prof" />
          <div class="badges-grid">
            <div v-for="(badge, idx) in badges" :key="badge.name" class="badge-item reveal" :style="{ transitionDelay: (0.3 + idx * 0.1) + 's' }">
              <img :src="badge.src" :alt="badge.name" class="badge-img" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════════════════════════════════════ SUPPORT ══ -->
    <section class="section support-section">
      <img :src="img.mascot" alt="Kora" class="support-kora" />
      <div class="support-content reveal">
        <h2 class="section-title">Soutiens le projet</h2>
        <p class="support-desc">
          AniQuiz est entièrement gratuit et sans publicité. Si tu aimes le projet et veux
          aider à le faire grandir (nouveaux animes, nouvelles features, serveurs), un petit
          café fait toute la différence.
        </p>
        <a href="https://ko-fi.com/yatokishi" target="_blank" rel="noopener" class="btn-kofi">
          <svg class="kofi-cup" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M2 3h18l-2 13H4L2 3zm16 0c0 0 1 4-2 6s-6 1-6 1"/><path d="M6 21h12"/></svg>
          Soutenir sur Ko-fi
        </a>
        <p class="support-note">Aucun contenu payant, juste de la bonne volonté ☕</p>
      </div>
    </section>

    <!-- ═══════════════════════════════════════ COMMUNAUTÉ ══ -->
    <section class="section community-section">
      <div class="community-inner">
        <img :src="img.koraFr" alt="Kora avec le drapeau français" class="kora-fr" />
        <div class="community-text reveal--left">
          <h2 class="section-title left">Fait en France, pour la communauté</h2>
          <p class="community-desc">
            AniQuiz est un projet indépendant né d'une passion pour l'anime et le jeu en ligne.
            Développé en France, il évolue grâce aux retours de ses joueurs.
          </p>
          <p class="community-desc">
            Tu as une idée, tu as trouvé un bug, un anime manque dans la bibliothèque ?
            Chaque retour compte et contribue directement à améliorer l'expérience pour tout le monde.
          </p>
          <div class="community-actions">
            <a href="https://discord.gg/RZhW7qparB" target="_blank" rel="noopener" class="btn-discord">
              <svg class="discord-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/></svg>
              Rejoindre le Discord
            </a>
            <a href="https://github.com/LeYapson/aniquiz/issues" target="_blank" rel="noopener" class="btn-feedback">
              Donner mon avis
            </a>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════════════════════════════════════ FOOTER ══ -->
    <footer class="landing-footer">
      <span>© {{ currentYear }} AniQuiz — Fait avec passion par des fans d'anime</span>
      <div class="footer-links">
        <button @click="emit('play')" class="footer-link">Jouer</button>
        <button @click="emit('leaderboard')" class="footer-link">Classement</button>
        <RouterLink to="/legal" class="footer-link">Mentions légales</RouterLink>
        <RouterLink to="/terms" class="footer-link">CGU</RouterLink>
        <RouterLink to="/privacy" class="footer-link">Confidentialité</RouterLink>
        <a href="https://discord.gg/RZhW7qparB" target="_blank" rel="noopener" class="footer-link">Discord</a>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { API_URL } from '../config';

const emit = defineEmits(['play', 'leaderboard']);

const currentYear = new Date().getFullYear();
const playersOnline = ref(null);

let statsInterval = null;

async function fetchStats() {
  try {
    const res = await fetch(`${API_URL}/api/stats`);
    if (res.ok) {
      const data = await res.json();
      playersOnline.value = data.players_online ?? 0;
    }
  } catch { /* silencieux — stat non critique */ }
}

onMounted(() => {
  fetchStats();
  statsInterval = setInterval(fetchStats, 30_000);

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.classList.add('is-visible');
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }
  );
  document.querySelectorAll('.reveal, .reveal--left, .reveal--right').forEach(el => observer.observe(el));
});

onUnmounted(() => clearInterval(statsInterval));

const img = {
  logo: '/logo.png',
  mascot: '/mascot_kora.png',
  koraProf: '/kora_prof.png',
  koraFr: '/kora-fr.png',
  koraDaily: '/kora-daily.png',
};

const steps = [
  { num: 1, title: 'Rejoins un salon', body: "Crée une partie ou rejoins un salon public. Tu peux aussi jouer seul pour t'entraîner." },
  { num: 2, title: 'Écoute et reconnais', body: "Une musique d'anime se lance. Opening, ending ou OST — à toi de trouver le titre avant les autres." },
  { num: 3, title: 'Marque des points', body: 'Plus tu réponds vite, plus tu marques. Le premier à trouver reçoit un bonus. Grimpe au classement !' },
];

const modes = [
  { key: 'multi',    icon: '⚔️', title: 'Multijoueur', soon: false, body: "Rejoins ou crée un salon, affronte d'autres joueurs en temps réel et sois le premier à trouver le bon anime." },
  { key: 'solo',     icon: '🎯', title: 'Solo',         soon: false, body: "Joue à ton rythme, sans pression. Configure tes rounds et tes filtres — parfait pour s'entraîner ou explorer de nouveaux animes." },
  { key: 'speedrun', icon: '⚡', title: 'Speed Run',    soon: false, body: "5 minutes, un maximum d'animes. Enchaîne les pistes, entretiens ta série de bonnes réponses et bats ton meilleur score." },
  { key: 'ranked',   icon: '🏆', title: 'Classé',       soon: true,  body: 'Grimpe les divisions, accumule des points de ranking et prouve ta valeur face aux meilleurs joueurs de la saison.' },
];

const progressionFeats = [
  { icon: '⭐', title: 'XP & Niveaux', body: "Gagne de l'XP à chaque partie et monte en niveau. Ta progression est visible sur ton profil et dans le classement global." },
  { icon: '🖼️', title: 'Cadres d\'avatar', body: 'Débloque des cadres exclusifs du Bronze au Rainbow en montant de niveau. Affiche ton rang avec style.' },
  { icon: '👥', title: 'Amis & Invitations', body: 'Ajoute tes amis et invite-les directement dans tes salons depuis le header. Jouer ensemble n\'a jamais été aussi simple.' },
];

const dailyFeatures = [
  'Même piste pour tous les joueurs ce jour-là',
  'Une seule tentative — pas de deuxième chance',
  'Classement remis à zéro chaque jour à minuit',
  'Choix multiples pour jouer sans se tromper d\'orthographe',
];

const rankedFeatures = [
  '5 divisions : Bronze → Challenger',
  'Saisons de 3 mois avec réinitialisation',
  'Récompenses cosmétiques par division',
  'Top 500 affiché en temps réel',
];

const badges = [
  { name: 'Bronze',     src: '/badge_bronze.png' },
  { name: 'Silver',     src: '/badge_silver.png' },
  { name: 'Gold',       src: '/badge_gold.png' },
  { name: 'Platinum',   src: '/badge_platinum.png' },
  { name: 'Challenger', src: '/badge_challenger.png' },
];
</script>

<style scoped>
/* ─── Base ─────────────────────────────────────────────────── */
.landing {
  background: #0f0f23;
  color: #f1f5f9;
  font-family: inherit;
  overflow-x: hidden;
}

.section {
  padding: 96px 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.section-title {
  font-size: 2rem;
  font-weight: 800;
  text-align: center;
  margin-bottom: 10px;
  background: linear-gradient(135deg, #f97316, #fb923c);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.section-title.left { text-align: left; }

.section-sub {
  text-align: center;
  color: #64748b;
  margin-bottom: 52px;
  font-size: 1rem;
}

/* ─── Hero ──────────────────────────────────────────────────── */
.hero {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px 24px;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #0f0f23 0%, #1a1a2e 50%, #16213e 100%);
}

.hero::before,
.hero::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.hero::before {
  width: 600px; height: 600px;
  background: radial-gradient(circle, rgba(249,115,22,0.12), transparent 70%);
  top: -200px; right: -100px;
}
.hero::after {
  width: 400px; height: 400px;
  background: radial-gradient(circle, rgba(59,130,246,0.08), transparent 70%);
  bottom: -150px; left: -100px;
}

.hero-content {
  max-width: 600px;
  text-align: center;
  z-index: 1;
}

.hero-logo {
  width: 220px;
  margin-bottom: 16px;
  filter: drop-shadow(0 4px 24px rgba(249,115,22,0.35));
}

.hero-h1 {
  font-size: 1.35rem;
  font-weight: 700;
  color: #f97316;
  margin: 0 0 16px;
  letter-spacing: 0.01em;
}

.hero-tagline {
  color: #cbd5e1;
  font-size: 1.15rem;
  line-height: 1.75;
  margin-bottom: 40px;
}

.hero-actions {
  display: flex;
  gap: 14px;
  justify-content: center;
  flex-wrap: wrap;
}

.btn-play {
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: white;
  border: none;
  padding: 15px 40px;
  border-radius: 50px;
  font-size: 1.05rem;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 24px rgba(249,115,22,0.4);
  transition: transform 0.15s, box-shadow 0.15s;
}
.btn-play:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(249,115,22,0.55);
}

.btn-lb {
  background: transparent;
  color: #94a3b8;
  border: 1px solid #334155;
  padding: 15px 30px;
  border-radius: 50px;
  font-size: 1rem;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}
.btn-lb:hover { color: #f97316; border-color: #f97316; }

.hero-kora {
  position: absolute;
  right: 0;
  bottom: 0;
  height: 430px;
  z-index: 1;
  filter: drop-shadow(-8px 0 32px rgba(0,0,0,0.6));
  pointer-events: none;
}

/* ─── Hero stats + scroll hint ─────────────────────────────── */
.hero-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20px;
  margin-top: 40px;
  padding: 16px 28px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 50px;
  width: fit-content;
  margin-left: auto;
  margin-right: auto;
}
.hero-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.hero-stat strong { color: #f97316; font-size: 1rem; font-weight: 700; display: flex; align-items: center; gap: 6px; }
.hero-stat span   { color: #475569; font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.05em; }
.hero-stat-sep    { width: 1px; height: 28px; background: rgba(255,255,255,0.1); }

.hero-stat--live strong { color: #4ade80; }

.live-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #4ade80;
  box-shadow: 0 0 0 0 rgba(74, 222, 128, 0.5);
  animation: livePulse 2s ease-in-out infinite;
  flex-shrink: 0;
}
@keyframes livePulse {
  0%   { box-shadow: 0 0 0 0   rgba(74, 222, 128, 0.5); }
  70%  { box-shadow: 0 0 0 7px rgba(74, 222, 128, 0); }
  100% { box-shadow: 0 0 0 0   rgba(74, 222, 128, 0); }
}

.hero-scroll-hint {
  position: absolute;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 2;
}
.scroll-arrow {
  display: block;
  width: 20px;
  height: 20px;
  border-right: 2px solid rgba(249,115,22,0.5);
  border-bottom: 2px solid rgba(249,115,22,0.5);
  transform: rotate(45deg);
  animation: scrollBounce 1.6s ease-in-out infinite;
}
@keyframes scrollBounce {
  0%, 100% { transform: rotate(45deg) translateY(0); opacity: 0.5; }
  50%       { transform: rotate(45deg) translateY(5px); opacity: 1; }
}

/* ─── Comment ça marche ─────────────────────────────────────── */
.howto-section { border-top: 1px solid rgba(255,255,255,0.06); }

.howto-steps {
  display: flex;
  align-items: flex-start;
  gap: 0;
  justify-content: center;
}

.step {
  flex: 1;
  max-width: 280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 0 16px;
}

.step-num {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: white;
  font-size: 1.3rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  box-shadow: 0 4px 16px rgba(249,115,22,0.35);
}

.step h3 { font-size: 1rem; font-weight: 700; color: #f1f5f9; margin: 0 0 10px; }
.step p   { font-size: 0.85rem; color: #64748b; line-height: 1.6; margin: 0; }

.step-arrow {
  flex-shrink: 0;
  width: 40px;
  height: 2px;
  background: linear-gradient(90deg, rgba(249,115,22,0.5), rgba(249,115,22,0.15));
  margin-top: 26px;
  position: relative;
}
.step-arrow::after {
  content: '';
  position: absolute;
  right: -1px;
  top: -4px;
  width: 0;
  height: 0;
  border-left: 7px solid rgba(249,115,22,0.5);
  border-top: 5px solid transparent;
  border-bottom: 5px solid transparent;
}

/* ─── Modes ─────────────────────────────────────────────────── */
.modes-section {
  border-top: 1px solid rgba(255,255,255,0.06);
}

.modes-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

/* ─── Progression ───────────────────────────────────────────── */
.progression-section {
  border-top: 1px solid rgba(255,255,255,0.06);
}

.progression-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.prog-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 28px 24px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: border-color 0.2s, background 0.2s;
}
.prog-card:hover {
  background: rgba(249,115,22,0.05);
  border-color: rgba(249,115,22,0.2);
}

.prog-icon { font-size: 2rem; }
.prog-card h3 { font-size: 1rem; font-weight: 700; color: #f1f5f9; margin: 0; }
.prog-card p  { font-size: 0.88rem; color: #64748b; line-height: 1.6; margin: 0; }

/* ── Mode cards : structure ─────────────────────────────────── */
.mode-card {
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid rgba(255,255,255,0.08);
  display: flex;
  flex-direction: column;
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}
.mode-card:hover { transform: translateY(-4px); }
.mode-card--soon { opacity: 0.65; }
.mode-card--soon:hover { transform: none; }

/* ── Header coloré par mode ─────────────────────────────────── */
.mode-card-header {
  position: relative;
  height: 130px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

/* Shimmer au hover */
.mode-card-header::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(105deg, transparent 40%, rgba(255,255,255,0.07) 50%, transparent 60%);
  transform: translateX(-100%);
  transition: transform 0.45s ease;
}
.mode-card:hover .mode-card-header::after { transform: translateX(100%); }

.mode-card-icon {
  font-size: 3.5rem;
  z-index: 1;
  transition: transform 0.2s ease;
  filter: drop-shadow(0 2px 12px rgba(0,0,0,0.5));
}
.mode-card:hover .mode-card-icon { transform: scale(1.12); }

/* ── Couleur par mode ────────────────────────────────────────── */
.mode-card--multi .mode-card-header {
  background: linear-gradient(135deg, #1e1b4b 0%, #312e81 60%, #1e3a5f 100%);
}
.mode-card--multi:hover { border-color: rgba(129,140,248,0.4); box-shadow: 0 8px 32px rgba(129,140,248,0.12); }
.mode-card--multi .mode-card-icon { filter: drop-shadow(0 0 14px rgba(129,140,248,0.7)); }

.mode-card--solo .mode-card-header {
  background: linear-gradient(135deg, #0a2518 0%, #064e3b 60%, #0e3a2f 100%);
}
.mode-card--solo:hover { border-color: rgba(52,211,153,0.4); box-shadow: 0 8px 32px rgba(52,211,153,0.1); }
.mode-card--solo .mode-card-icon { filter: drop-shadow(0 0 14px rgba(52,211,153,0.6)); }

.mode-card--speedrun .mode-card-header {
  background: linear-gradient(135deg, #1c0800 0%, #431407 60%, #2d0f00 100%);
}
.mode-card--speedrun:hover { border-color: rgba(249,115,22,0.5); box-shadow: 0 8px 32px rgba(249,115,22,0.15); }
.mode-card--speedrun .mode-card-icon { filter: drop-shadow(0 0 16px rgba(249,115,22,0.8)); }

.mode-card--ranked .mode-card-header {
  background: linear-gradient(135deg, #1c1400 0%, #3d2e00 60%, #2a1e00 100%);
}
.mode-card--ranked:hover { border-color: rgba(251,191,36,0.4); box-shadow: 0 8px 32px rgba(251,191,36,0.1); }
.mode-card--ranked .mode-card-icon { filter: drop-shadow(0 0 14px rgba(251,191,36,0.6)); }

/* ── Corps de la carte ──────────────────────────────────────── */
.mode-card-body {
  background: rgba(255,255,255,0.03);
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}
.mode-card-body h3 { font-size: 1.05rem; font-weight: 700; color: #f1f5f9; margin: 0; }
.mode-card-body p  { font-size: 0.85rem; color: #64748b; line-height: 1.6; margin: 0; flex: 1; }

.mode-badge {
  align-self: flex-start;
  padding: 3px 10px;
  border-radius: 50px;
  font-size: 0.75rem;
  font-weight: 600;
}
.mode-available { background: rgba(34,197,94,0.15); color: #4ade80; }
.mode-soon      { background: rgba(100,116,139,0.15); color: #94a3b8; }

/* ─── Quiz du jour ──────────────────────────────────────────── */
.daily-section {
  border-top: 1px solid rgba(255,255,255,0.06);
  background: linear-gradient(180deg, transparent, rgba(249,115,22,0.02), transparent);
}
.daily-inner {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
}
.daily-badge {
  display: inline-block;
  background: rgba(249,115,22,0.15);
  color: #f97316;
  border: 1px solid rgba(249,115,22,0.3);
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 12px;
  border-radius: 99px;
  margin-bottom: 14px;
}
.daily-desc {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.7;
  margin-bottom: 20px;
}
.daily-features {
  list-style: none;
  padding: 0;
  margin: 0 0 28px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.daily-features li {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #cbd5e1;
  font-size: 0.9rem;
}
.daily-cta { max-width: 200px; }
.daily-visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}
.kora-daily {
  width: 220px;
  filter: drop-shadow(0 4px 24px rgba(249,115,22,0.25));
  animation: floatKora 5s 1s ease-in-out infinite;
}
.daily-stats-pills {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.daily-pill {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  padding: 8px 14px;
  font-size: 0.82rem;
  color: #94a3b8;
  text-align: center;
}

/* ─── Ranked ────────────────────────────────────────────────── */
.ranked-section {
  border-top: 1px solid rgba(255,255,255,0.06);
  background: linear-gradient(180deg, transparent, rgba(249,115,22,0.03), transparent);
  max-width: 100%;
  padding: 96px 0;
}

.ranked-inner {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
}

.ranked-desc {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.7;
  margin-bottom: 24px;
}

.ranked-features {
  list-style: none;
  padding: 0;
  margin: 0 0 32px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.ranked-features li {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #cbd5e1;
  font-size: 0.9rem;
}
.rf-dot {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #f97316;
  flex-shrink: 0;
}

.coming-soon-pill {
  display: inline-block;
  background: rgba(249,115,22,0.12);
  color: #f97316;
  border: 1px solid rgba(249,115,22,0.25);
  padding: 6px 18px;
  border-radius: 50px;
  font-size: 0.82rem;
  font-weight: 600;
}

.ranked-visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.kora-prof {
  width: 200px;
  filter: drop-shadow(0 4px 20px rgba(249,115,22,0.2));
}

.badges-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  width: 100%;
}

.badge-item {
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 12px;
  padding: 10px 6px;
  transition: transform 0.15s, background 0.15s;
}
.badge-item:hover {
  transform: translateY(-3px);
  background: rgba(249,115,22,0.08);
}

.badge-img {
  width: 80px;
  height: auto;
  object-fit: contain;
}

/* Challenger occupe toute la largeur sur la 2e ligne */
.badge-item:last-child {
  grid-column: 2 / 3;
}

/* ─── Support ───────────────────────────────────────────────── */
.support-section {
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  align-items: center;
  gap: 64px;
  max-width: 900px;
}

.support-kora {
  width: 160px;
  flex-shrink: 0;
  filter: drop-shadow(0 4px 16px rgba(0,0,0,0.4));
}

.support-desc {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.7;
  margin-bottom: 28px;
}

.btn-kofi {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: #ff5e5b;
  color: white;
  text-decoration: none;
  padding: 13px 28px;
  border-radius: 50px;
  font-weight: 700;
  font-size: 0.95rem;
  transition: opacity 0.15s, transform 0.15s;
  box-shadow: 0 4px 20px rgba(255,94,91,0.35);
}
.btn-kofi:hover { opacity: 0.88; transform: translateY(-2px); }

.kofi-cup {
  width: 22px;
  height: 22px;
  object-fit: contain;
}

.support-note {
  margin-top: 14px;
  color: #475569;
  font-size: 0.82rem;
}

/* ─── Communauté ────────────────────────────────────────────── */
.community-section {
  border-top: 1px solid rgba(255,255,255,0.06);
  max-width: 960px;
}

.community-inner {
  display: flex;
  align-items: center;
  gap: 56px;
}

.kora-fr {
  width: 180px;
  flex-shrink: 0;
  filter: drop-shadow(0 4px 20px rgba(0,0,0,0.4));
}

.community-desc {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.7;
  margin-bottom: 14px;
}
.community-desc:last-of-type { margin-bottom: 28px; }

.community-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.btn-discord {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #5865f2;
  color: #fff;
  text-decoration: none;
  padding: 11px 26px;
  border-radius: 50px;
  font-weight: 700;
  font-size: 0.9rem;
  transition: opacity 0.15s, transform 0.15s;
  box-shadow: 0 4px 16px rgba(88,101,242,0.35);
}
.btn-discord:hover { opacity: 0.88; transform: translateY(-2px); }

.discord-icon { width: 18px; height: 18px; flex-shrink: 0; }

.btn-feedback {
  display: inline-block;
  background: transparent;
  color: #f97316;
  border: 1px solid rgba(249,115,22,0.5);
  text-decoration: none;
  padding: 11px 26px;
  border-radius: 50px;
  font-weight: 600;
  font-size: 0.9rem;
  transition: background 0.15s, border-color 0.15s;
}
.btn-feedback:hover {
  background: rgba(249,115,22,0.1);
  border-color: #f97316;
}

/* ─── Footer ────────────────────────────────────────────────── */
.landing-footer {
  border-top: 1px solid rgba(255,255,255,0.06);
  padding: 28px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #334155;
  font-size: 0.82rem;
  max-width: 1100px;
  margin: 0 auto;
}

.footer-links { display: flex; gap: 20px; flex-wrap: wrap; justify-content: center; }
.footer-link {
  background: none;
  border: none;
  color: #475569;
  font-size: 0.82rem;
  cursor: pointer;
  transition: color 0.15s;
  padding: 0;
  text-decoration: none;
}
.footer-link:hover { color: #f97316; }

/* ─── Responsive ────────────────────────────────────────────── */
@media (max-width: 900px) {
  .hero-kora { display: none; }
  .hero-stats { gap: 14px; padding: 12px 20px; }
  .howto-steps { flex-direction: column; align-items: center; gap: 24px; }
  .step-arrow { width: 2px; height: 32px; background: linear-gradient(180deg, rgba(249,115,22,0.5), rgba(249,115,22,0.15)); }
  .step-arrow::after { right: -4px; top: auto; bottom: -1px; border-left: 5px solid transparent; border-right: 5px solid transparent; border-top: 7px solid rgba(249,115,22,0.5); }
  .modes-grid { grid-template-columns: 1fr; }
  .progression-grid { grid-template-columns: 1fr; }
  .ranked-inner { grid-template-columns: 1fr; gap: 40px; }
  .daily-inner { grid-template-columns: 1fr; gap: 32px; }
  .daily-visual { order: -1; }
  .kora-daily { width: 160px; }
  .daily-cta { max-width: 100%; }
  .section-title.left { text-align: center; }
  .support-section { flex-direction: column; text-align: center; gap: 32px; }
  .support-kora { width: 120px; }
  .community-inner { flex-direction: column; text-align: center; gap: 28px; }
  .kora-fr { width: 130px; }
  .community-actions { justify-content: center; }
  .landing-footer { flex-direction: column; gap: 12px; text-align: center; }
}

@media (max-width: 500px) {
  .hero-logo { width: 160px; }
  .hero-tagline { font-size: 0.95rem; }
  .section { padding: 64px 16px; }
  .badges-grid { grid-template-columns: repeat(2, 1fr); }
  .badge-item:last-child { grid-column: 1 / -1; }
}

/* ─── Keyframes ─────────────────────────────────────────────── */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(24px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes fadeInDown {
  from { opacity: 0; transform: translateY(-16px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes slideInRight {
  from { opacity: 0; transform: translateX(48px); }
  to   { opacity: 1; transform: translateX(0); }
}
@keyframes floatKora {
  0%, 100% { transform: translateY(0); }
  50%      { transform: translateY(-14px); }
}
@keyframes orbPulse {
  0%, 100% { transform: scale(1);    opacity: 1; }
  50%      { transform: scale(1.18); opacity: 0.7; }
}
@keyframes noteFloat {
  0%   { transform: translateY(0)   rotate(0deg);  opacity: 0; }
  10%  { opacity: 0.5; }
  90%  { opacity: 0.2; }
  100% { transform: translateY(-130px) rotate(18deg); opacity: 0; }
}

/* ─── Hero : animations au chargement ──────────────────────── */
.hero-logo    { animation: fadeInDown  0.6s ease both; }
.hero-h1      { animation: fadeInUp    0.6s 0.2s  ease both; }
.hero-tagline { animation: fadeInUp    0.6s 0.35s ease both; }
.hero-actions { animation: fadeInUp    0.6s 0.5s  ease both; }
.hero-stats   { animation: fadeInUp    0.6s 0.65s ease both; }
.hero-kora    { animation: slideInRight 0.7s 0.3s ease both, floatKora 5s 1.2s ease-in-out infinite; }

.hero::before { animation: orbPulse  8s      ease-in-out infinite; }
.hero::after  { animation: orbPulse 10s 2.5s ease-in-out infinite; }

/* ─── Notes de musique flottantes ───────────────────────────── */
.hero-particles {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
}
.hero-particles span {
  position: absolute;
  bottom: 8%;
  color: rgba(249, 115, 22, 0.35);
  opacity: 0;
  animation: noteFloat linear infinite;
  user-select: none;
}
.hero-particles span:nth-child(1) { left:  8%; font-size: 1.2rem; animation-duration:  7s; animation-delay:  0s;   }
.hero-particles span:nth-child(2) { left: 18%; font-size: 1.0rem; animation-duration:  9s; animation-delay:  1.5s; }
.hero-particles span:nth-child(3) { left: 32%; font-size: 1.5rem; animation-duration:  6s; animation-delay:  3.2s; }
.hero-particles span:nth-child(4) { left: 50%; font-size: 1.1rem; animation-duration:  8s; animation-delay:  0.8s; }
.hero-particles span:nth-child(5) { left: 63%; font-size: 0.9rem; animation-duration: 11s; animation-delay:  2.0s; }
.hero-particles span:nth-child(6) { left: 76%; font-size: 1.3rem; animation-duration:  7s; animation-delay:  4.1s; }
.hero-particles span:nth-child(7) { left: 88%; font-size: 1.0rem; animation-duration:  9s; animation-delay:  1.0s; }

/* ─── Scroll reveal ─────────────────────────────────────────── */
.reveal,
.reveal--left,
.reveal--right {
  opacity: 0;
  transition: opacity 0.6s ease, transform 0.6s ease;
}
.reveal       { transform: translateY(28px); }
.reveal--left { transform: translateX(-32px); }
.reveal--right{ transform: translateX(32px); }

.reveal.is-visible,
.reveal--left.is-visible,
.reveal--right.is-visible {
  opacity: 1;
  transform: none;
}
</style>
