<template>
  <div class="profile-page">
    <div v-if="loading" class="loading">Chargement…</div>

    <template v-else-if="profile">
      <!-- En-tête profil -->
      <div class="profile-header">
        <div class="avatar">{{ profile.username.charAt(0).toUpperCase() }}</div>
        <div class="profile-info">
          <h2>{{ profile.username }}</h2>
          <div class="level-badge">Niveau {{ profile.level }}</div>
          <div class="xp-bar-wrap">
            <div class="xp-bar" :style="{ width: xpProgress + '%' }"></div>
          </div>
          <small>{{ profile.xp }} XP · {{ xpToNextLevel }} XP pour le niveau {{ profile.level + 1 }}</small>
        </div>
      </div>

      <!-- Comptes liés -->
      <div class="section">
        <h3>Comptes liés</h3>
        <div class="linked-accounts">
          <div class="account-card" :class="{ linked: profile.anilist_username }">
            <span class="acc-icon anilist-color">A</span>
            <span v-if="profile.anilist_username">AniList : <strong>{{ profile.anilist_username }}</strong></span>
            <span v-else class="not-linked">AniList non connecté</span>
          </div>
          <div class="account-card" :class="{ linked: profile.mal_username }">
            <span class="acc-icon mal-color">M</span>
            <span v-if="profile.mal_username">MAL : <strong>{{ profile.mal_username }}</strong></span>
            <span v-else class="not-linked">MyAnimeList non connecté</span>
          </div>
        </div>
      </div>

      <!-- Stats globales -->
      <div class="section">
        <h3>Statistiques</h3>
        <div class="stats-grid">
          <div class="stat-card">
            <span class="stat-value">{{ stats.totalGames }}</span>
            <span class="stat-label">Parties jouées</span>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats.bestScore }}</span>
            <span class="stat-label">Meilleur score</span>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats.avgScore }}</span>
            <span class="stat-label">Score moyen</span>
          </div>
          <div class="stat-card">
            <span class="stat-value">{{ stats.totalXP }}</span>
            <span class="stat-label">XP total gagné</span>
          </div>
        </div>
      </div>

      <!-- Historique -->
      <div class="section">
        <h3>Dernières parties</h3>
        <div v-if="history.length === 0" class="empty">Aucune partie jouée pour l'instant.</div>
        <table v-else class="history-table">
          <thead>
            <tr><th>Date</th><th>Score</th><th>XP gagné</th></tr>
          </thead>
          <tbody>
            <tr v-for="r in history" :key="r.id">
              <td>{{ formatDate(r.played_at) }}</td>
              <td>{{ r.score }} pts</td>
              <td>+{{ r.xp_gained }} XP</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <div v-else class="error">Impossible de charger le profil.</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { authStore } from "../authStore";

const profile = ref(null);
const history = ref([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const [profileRes, historyRes] = await Promise.all([
      fetch("http://localhost:8080/api/profile", { headers: authStore.authHeaders() }),
      fetch("http://localhost:8080/api/history", { headers: authStore.authHeaders() }),
    ]);
    if (profileRes.ok) profile.value = await profileRes.json();
    if (historyRes.ok) history.value = await historyRes.json();
  } finally {
    loading.value = false;
  }
});

// XP nécessaire pour atteindre un niveau : level = floor(sqrt(xp/100)) + 1
// => xp_seuil(level) = (level - 1)^2 * 100
const xpThreshold = (level) => Math.pow(level - 1, 2) * 100;

const xpProgress = computed(() => {
  if (!profile.value) return 0;
  const current = profile.value.xp - xpThreshold(profile.value.level);
  const needed = xpThreshold(profile.value.level + 1) - xpThreshold(profile.value.level);
  return Math.min(Math.round((current / needed) * 100), 100);
});

const xpToNextLevel = computed(() => {
  if (!profile.value) return 0;
  return xpThreshold(profile.value.level + 1) - profile.value.xp;
});

const stats = computed(() => {
  if (!history.value.length) return { totalGames: 0, bestScore: 0, avgScore: 0, totalXP: 0 };
  const scores = history.value.map((r) => r.score);
  return {
    totalGames: history.value.length,
    bestScore: Math.max(...scores),
    avgScore: Math.round(scores.reduce((a, b) => a + b, 0) / scores.length),
    totalXP: history.value.reduce((a, r) => a + r.xp_gained, 0),
  };
});

const formatDate = (iso) => {
  const d = new Date(iso);
  return d.toLocaleDateString("fr-FR", { day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit" });
};
</script>

<style scoped>
.profile-page { max-width: 700px; margin: 30px auto; padding: 0 20px; }
.loading, .error, .empty { text-align: center; color: #888; padding: 40px; }

.profile-header {
  display: flex; gap: 20px; align-items: center;
  background: #1a1a2e; color: #fff; border-radius: 12px; padding: 24px; margin-bottom: 24px;
}
.avatar {
  width: 64px; height: 64px; border-radius: 50%;
  background: #e91e63; color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.8rem; font-weight: bold; flex-shrink: 0;
}
.profile-info { flex: 1; }
.profile-info h2 { margin: 0 0 4px; font-size: 1.4rem; }
.level-badge {
  display: inline-block; background: #ffd700; color: #1a1a2e;
  font-weight: bold; font-size: 0.8rem; padding: 2px 10px;
  border-radius: 20px; margin-bottom: 8px;
}
.xp-bar-wrap { background: #333; border-radius: 4px; height: 8px; margin-bottom: 4px; }
.xp-bar { background: #e91e63; height: 8px; border-radius: 4px; transition: width 0.4s ease; }
.profile-info small { color: #aaa; font-size: 0.8rem; }

.section { margin-bottom: 28px; }
.section h3 { font-size: 1rem; font-weight: bold; color: #333; border-bottom: 2px solid #eee; padding-bottom: 6px; margin-bottom: 14px; }

.linked-accounts { display: flex; gap: 12px; }
.account-card {
  flex: 1; display: flex; align-items: center; gap: 10px;
  padding: 12px 16px; border-radius: 8px;
  background: #f4f4f4; border: 1px solid #ddd;
}
.account-card.linked { background: #f0fdf4; border-color: #86efac; }
.acc-icon {
  width: 28px; height: 28px; border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; font-size: 0.9rem; color: #fff; flex-shrink: 0;
}
.anilist-color { background: #02a9ff; }
.mal-color { background: #2e51a2; }
.not-linked { color: #999; font-size: 0.9rem; }

.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.stat-card {
  background: #f4f4f4; border-radius: 8px; padding: 16px;
  display: flex; flex-direction: column; align-items: center; gap: 4px;
}
.stat-value { font-size: 1.6rem; font-weight: bold; color: #1a1a2e; }
.stat-label { font-size: 0.75rem; color: #666; text-align: center; }

.history-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
.history-table th { background: #f4f4f4; padding: 8px 12px; text-align: left; font-weight: 600; }
.history-table td { padding: 8px 12px; border-bottom: 1px solid #eee; }
.history-table tr:last-child td { border-bottom: none; }
</style>
