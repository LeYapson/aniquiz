<template>
  <div class="landing">

    <!-- ═══════════════════════════════════════════ HERO ══ -->
    <section class="hero">
      <div class="hero-content">
        <img :src="img.logo" alt="AniQuiz" class="hero-logo" />
        <p class="hero-tagline">
          Reconnais les génériques d'anime avant tes adversaires.<br />
          En multi, en solo ou en Speed Run — prouve que tu es le vrai otaku.
        </p>
        <div class="hero-actions">
          <button @click="emit('play')" class="btn-play">Jouer maintenant</button>
          <button @click="emit('leaderboard')" class="btn-lb"><span aria-hidden="true">🏆</span> Classement</button>
        </div>
        <div class="hero-stats">
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
          <div class="step">
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
          v-for="mode in modes"
          :key="mode.title"
          class="mode-card"
          :class="{ 'mode-card--soon': mode.soon }"
        >
          <div class="mode-icon" aria-hidden="true">{{ mode.icon }}</div>
          <h3>{{ mode.title }}</h3>
          <p>{{ mode.body }}</p>
          <span class="mode-badge" :class="mode.soon ? 'mode-soon' : 'mode-available'">{{ mode.soon ? 'Bientôt' : 'Disponible' }}</span>
        </div>
      </div>
    </section>

    <!-- ═══════════════════════════════════ PROGRESSION & SOCIAL ══ -->
    <section class="section progression-section">
      <h2 class="section-title">Progresse et personnalise</h2>
      <p class="section-sub">Chaque partie contribue à ta progression permanente</p>
      <div class="progression-grid">
        <div class="prog-card" v-for="feat in progressionFeats" :key="feat.icon">
          <div class="prog-icon" aria-hidden="true">{{ feat.icon }}</div>
          <h3>{{ feat.title }}</h3>
          <p>{{ feat.body }}</p>
        </div>
      </div>
    </section>

    <!-- ═══════════════════════════════════════ RANKED / DIVISIONS ══ -->
    <section class="section ranked-section">
      <div class="ranked-inner">
        <div class="ranked-text">
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

        <div class="ranked-visual">
          <img :src="img.koraProf" alt="Kora explique le système ranked" class="kora-prof" />
          <div class="badges-grid">
            <div v-for="badge in badges" :key="badge.name" class="badge-item">
              <img :src="badge.src" :alt="badge.name" class="badge-img" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ══════════════════════════════════════════ SUPPORT ══ -->
    <section class="section support-section">
      <img :src="img.mascot" alt="Kora" class="support-kora" />
      <div class="support-content">
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
        <div class="community-text">
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
        <a href="https://discord.gg/RZhW7qparB" target="_blank" rel="noopener" class="footer-link">Discord</a>
      </div>
    </footer>
  </div>
</template>

<script setup>
const emit = defineEmits(['play', 'leaderboard']);

const currentYear = new Date().getFullYear();

const img = {
  logo: '/logo.png',
  mascot: '/mascot_kora.png',
  koraProf: '/kora_prof.png',
  koraFr: '/kora-fr.png',
};

const steps = [
  { num: 1, title: 'Rejoins un salon', body: "Crée une partie ou rejoins un salon public. Tu peux aussi jouer seul pour t'entraîner." },
  { num: 2, title: 'Écoute et reconnais', body: "Une musique d'anime se lance. Opening, ending ou OST — à toi de trouver le titre avant les autres." },
  { num: 3, title: 'Marque des points', body: 'Plus tu réponds vite, plus tu marques. Le premier à trouver reçoit un bonus. Grimpe au classement !' },
];

const modes = [
  { icon: '⚔️', title: 'Multijoueur', soon: false, body: "Rejoins ou crée un salon, affronte d'autres joueurs en temps réel et sois le premier à trouver le bon anime." },
  { icon: '🎯', title: 'Solo', soon: false, body: "Joue à ton rythme, sans pression. Configure tes rounds et tes filtres — parfait pour s'entraîner ou explorer de nouveaux animes." },
  { icon: '⚡', title: 'Speed Run', soon: false, body: "5 minutes, un maximum d'animes. Enchaîne les pistes, entretiens ta série de bonnes réponses et bats ton meilleur score." },
  { icon: '🏆', title: 'Classé', soon: true, body: 'Grimpe les divisions, accumule des points de ranking et prouve ta valeur face aux meilleurs joueurs de la saison.' },
];

const progressionFeats = [
  { icon: '⭐', title: 'XP & Niveaux', body: "Gagne de l'XP à chaque partie et monte en niveau. Ta progression est visible sur ton profil et dans le classement global." },
  { icon: '🖼️', title: 'Cadres d\'avatar', body: 'Débloque des cadres exclusifs du Bronze au Rainbow en montant de niveau. Affiche ton rang avec style.' },
  { icon: '👥', title: 'Amis & Invitations', body: 'Ajoute tes amis et invite-les directement dans tes salons depuis le header. Jouer ensemble n\'a jamais été aussi simple.' },
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
  margin-bottom: 28px;
  filter: drop-shadow(0 4px 24px rgba(249,115,22,0.35));
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
.hero-stat strong { color: #f97316; font-size: 1rem; font-weight: 700; }
.hero-stat span   { color: #475569; font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.05em; }
.hero-stat-sep    { width: 1px; height: 28px; background: rgba(255,255,255,0.1); }

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

.mode-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 28px 24px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: border-color 0.2s, background 0.2s;
}
.mode-card:hover {
  background: rgba(249,115,22,0.06);
  border-color: rgba(249,115,22,0.25);
}
.mode-card--soon {
  opacity: 0.65;
}
.mode-card--soon:hover {
  background: rgba(100,116,139,0.06);
  border-color: rgba(100,116,139,0.2);
}

.mode-icon { font-size: 2rem; }
.mode-card h3 { font-size: 1.1rem; font-weight: 700; color: #f1f5f9; margin: 0; }
.mode-card p  { font-size: 0.88rem; color: #64748b; line-height: 1.6; margin: 0; flex: 1; }

.mode-badge {
  align-self: flex-start;
  padding: 3px 10px;
  border-radius: 50px;
  font-size: 0.75rem;
  font-weight: 600;
}
.mode-available { background: rgba(34,197,94,0.15); color: #4ade80; }
.mode-soon      { background: rgba(100,116,139,0.15); color: #94a3b8; }

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

.footer-links { display: flex; gap: 20px; }
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
</style>
