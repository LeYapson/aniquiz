<template>
  <header class="app-header">
    <!-- Gauche : logo -->
    <button class="header-logo-btn" @click="emit('navigate', 'home')" aria-label="Accueil">
      <img :src="logoSrc" alt="AniQuiz" class="header-logo" />
    </button>

    <!-- Centre : navigation -->
    <nav class="header-nav" aria-label="Navigation principale">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="nav-tab"
        :class="{ active: currentView === tab.id }"
        :aria-current="currentView === tab.id ? 'page' : undefined"
        @click="emit('navigate', tab.id)"
      >
        <span class="nav-icon" aria-hidden="true">{{ tab.icon }}</span>
        <span class="nav-label">{{ tab.label }}</span>
      </button>
    </nav>

    <!-- Droite : liaisons + carte joueur -->
    <div class="header-right">
      <div class="header-links">
        <span v-if="user?.anilist_username" class="link-pill linked" title="AniList lié">
          <span class="dot anilist"></span>AniList
        </span>
        <button v-else class="link-pill connect" @click="emit('connect-anilist')">
          + AniList
        </button>

        <span v-if="user?.mal_username" class="link-pill linked" title="MAL lié">
          <span class="dot mal"></span>MAL
        </span>
        <button v-else class="link-pill connect" @click="emit('connect-mal')">
          + MAL
        </button>
      </div>

      <NotificationsBell :inGame="inGame" @join="(inv) => emit('join-room', inv)" />

      <div class="player-card">
        <div class="avatar" :class="frameClass(user?.avatar_frame)">
          {{ initial }}
          <span class="avatar-level">{{ user?.level ?? 1 }}</span>
        </div>
        <div class="player-info">
          <span class="player-name">{{ user?.username }}</span>
          <div class="xp-bar" :aria-label="`${user?.xp ?? 0} XP`">
            <div class="xp-fill" :style="{ width: xpProgress + '%' }"></div>
          </div>
          <span class="player-xp">{{ user?.xp ?? 0 }} XP · Niv. {{ user?.level ?? 1 }}</span>
        </div>
      </div>

      <button class="logout-btn" @click="emit('logout')" aria-label="Déconnexion" title="Déconnexion">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
          <polyline points="16 17 21 12 16 7" />
          <line x1="21" y1="12" x2="9" y2="12" />
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue';
import { authStore } from '../authStore';
import { frameClass } from '../cosmetics';
import NotificationsBell from './NotificationsBell.vue';

defineProps({
  currentView: { type: String, default: 'home' },
  inGame: { type: Boolean, default: false },
});

const emit = defineEmits(['navigate', 'logout', 'connect-anilist', 'connect-mal', 'join-room']);

const logoSrc = '/logo.png';
const user = computed(() => authStore.user);

const tabs = [
  { id: 'home',        icon: '🎮', label: 'Jouer' },
  { id: 'leaderboard', icon: '🏆', label: 'Classement' },
  { id: 'profile',     icon: '👤', label: 'Profil' },
  { id: 'news',        icon: '📰', label: 'News' },
];

const initial = computed(() => (user.value?.username?.[0] ?? '?').toUpperCase());

const xpProgress = computed(() => {
  const u = user.value;
  if (!u) return 0;
  const lvl = u.level ?? 1;
  const s = (lvl - 1) ** 2 * 100;
  const e = lvl ** 2 * 100;
  if (e === s) return 0;
  return Math.max(0, Math.min(1, ((u.xp ?? 0) - s) / (e - s))) * 100;
});
</script>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  background: #0a0a1a;
  box-shadow: 0 1px 0 rgba(249, 115, 22, 0.2), 0 2px 12px rgba(0, 0, 0, 0.4);
  flex-shrink: 0;
  position: relative;
  z-index: 50;
}

/* ── Logo ── */
.header-logo-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}
.header-logo {
  height: 36px;
  transition: transform 0.15s;
}
.header-logo-btn:hover .header-logo { transform: scale(1.05); }

/* ── Navigation centrale ── */
.header-nav {
  display: flex;
  gap: 4px;
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}
.nav-tab {
  display: flex;
  align-items: center;
  gap: 7px;
  background: none;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  color: #64748b;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  position: relative;
  transition: color 0.15s, background 0.15s;
}
.nav-tab:hover { color: #f1f5f9; background: rgba(255, 255, 255, 0.04); }
.nav-tab.active { color: #f1f5f9; }
.nav-tab.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 16px;
  right: 16px;
  height: 2px;
  background: #f97316;
  border-radius: 2px;
}
.nav-icon { font-size: 1rem; }

/* ── Bloc droit ── */
.header-right {
  display: flex;
  align-items: center;
  gap: 14px;
}

.header-links {
  display: flex;
  gap: 6px;
}
.link-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 20px;
  cursor: default;
}
.link-pill.linked {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}
.link-pill.connect {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #64748b;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}
.link-pill.connect:hover { border-color: #f97316; color: #f97316; }
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.dot.anilist { background: #02a9ff; }
.dot.mal { background: #2e51a2; }

/* ── Carte joueur ── */
.player-card {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar {
  position: relative;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: #fff;
  font-weight: 800;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 10px rgba(249, 115, 22, 0.35);
}
.avatar-level {
  position: absolute;
  bottom: -3px;
  right: -3px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  border-radius: 9px;
  background: #0a0a1a;
  border: 2px solid #f97316;
  color: #f97316;
  font-size: 0.62rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
}
.player-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 96px;
}
.player-name {
  font-size: 0.85rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1;
}
.xp-bar {
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}
.xp-fill {
  height: 100%;
  background: linear-gradient(90deg, #f97316, #fb923c);
  border-radius: 2px;
  transition: width 0.4s ease;
}
.player-xp {
  font-size: 0.66rem;
  color: #475569;
  line-height: 1;
}

/* ── Logout ── */
.logout-btn {
  background: none;
  border: none;
  color: #475569;
  cursor: pointer;
  padding: 6px;
  display: flex;
  align-items: center;
  border-radius: 6px;
  transition: color 0.15s, background 0.15s;
}
.logout-btn:hover { color: #ef4444; background: rgba(239, 68, 68, 0.08); }

/* ── Responsive ── */
@media (max-width: 860px) {
  .header-nav {
    position: static;
    transform: none;
  }
  .nav-label { display: none; }
  .nav-tab { padding: 8px 10px; }
  .header-links { display: none; }
  .player-info { display: none; }
}

@media (max-width: 600px) {
  .app-header { padding: 0 12px; }
  .header-logo { height: 30px; }
  .header-nav { gap: 0; }
}
</style>
