<template>
  <footer class="app-footer">
    <div class="footer-grid">
      <!-- Marque -->
      <div class="footer-brand">
        <img :src="logoSrc" alt="AniQuiz" class="footer-logo" />
        <p class="footer-tagline">Le blindtest anime multijoueur, made in France.</p>
      </div>

      <!-- Liens site -->
      <div class="footer-col">
        <h4>Le jeu</h4>
        <a href="/">Jouer</a>
        <a href="/">Classement</a>
        <a href="/">Profil</a>
        <a href="/">News</a>
      </div>

      <!-- Liens légaux -->
      <div class="footer-col">
        <h4>Informations</h4>
        <RouterLink to="/legal">Mentions légales</RouterLink>
        <RouterLink to="/terms">CGU</RouterLink>
        <RouterLink to="/privacy">Vie privée</RouterLink>
        <a href="https://discord.gg/RZhW7qparB" target="_blank" rel="noopener" class="discord-link">💬 Discord</a>
        <a href="https://ko-fi.com/yatokishi" target="_blank" rel="noopener" class="kofi-link">☕ Soutenir</a>
      </div>
    </div>

    <div class="footer-bottom">
      <span>© 2026 AniQuiz · Fait avec passion en France 🇫🇷</span>
      <a href="https://github.com/LeYapson/aniquiz" target="_blank" rel="noopener" class="github-link">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor" aria-hidden="true">
          <path d="M12 .3a12 12 0 0 0-3.8 23.4c.6.1.8-.3.8-.6v-2c-3.3.7-4-1.6-4-1.6-.6-1.4-1.4-1.8-1.4-1.8-1-.7.1-.7.1-.7 1.2 0 1.9 1.2 1.9 1.2 1 1.8 2.8 1.3 3.5 1 0-.8.4-1.3.7-1.6-2.7-.3-5.5-1.3-5.5-6 0-1.2.5-2.3 1.3-3.1-.2-.4-.6-1.6 0-3.2 0 0 1-.3 3.4 1.2a11.5 11.5 0 0 1 6 0C17 4.6 18 5 18 5c.6 1.6.2 2.8 0 3.2.8.8 1.3 1.9 1.3 3.1 0 4.6-2.8 5.7-5.5 6 .5.4.9 1 .9 2.2v3.3c0 .3.2.7.8.6A12 12 0 0 0 12 .3" />
        </svg>
        GitHub
      </a>
    </div>
  </footer>
</template>

<script setup>
const logoSrc = '/logo.png';
</script>

<style scoped>
.app-footer {
  background: #070714;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  padding: 48px 24px 28px;
  color: #64748b;
}

.footer-grid {
  max-width: 1100px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 48px;
  padding-bottom: 36px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.footer-logo {
  height: 38px;
  margin-bottom: 14px;
}
.footer-tagline {
  font-size: 0.88rem;
  color: #475569;
  line-height: 1.6;
  max-width: 280px;
}

.footer-col {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.footer-col h4 {
  color: #94a3b8;
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  margin: 0 0 4px;
  font-weight: 700;
}
.footer-col a {
  color: #64748b;
  text-decoration: none;
  font-size: 0.88rem;
  transition: color 0.15s;
  width: fit-content;
}
.footer-col a:hover { color: #f97316; }
.discord-link { color: #7289da !important; }
.kofi-link { color: #fb923c !important; }

.footer-bottom {
  max-width: 1100px;
  margin: 0 auto;
  padding-top: 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
  color: #334155;
}
.github-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #475569;
  text-decoration: none;
  transition: color 0.15s;
}
.github-link:hover { color: #f1f5f9; }

@media (max-width: 700px) {
  .footer-grid {
    grid-template-columns: 1fr 1fr;
    gap: 32px;
  }
  .footer-brand { grid-column: 1 / -1; }
  .footer-bottom {
    flex-direction: column;
    gap: 12px;
    text-align: center;
  }
}
</style>
