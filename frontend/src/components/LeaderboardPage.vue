<template>
  <div class="leaderboard-page">
    <h2>Classement Global</h2>
    <div class="lb-tabs">
      <button
        class="lb-tab"
        :class="{ active: activeTab === 'global' }"
        @click="activeTab = 'global'"
      >🏆 Global</button>
      <button
        class="lb-tab"
        :class="{ active: activeTab === 'speedrun' }"
        @click="activeTab = 'speedrun'"
      >⚡ Speed Run</button>
    </div>

    <!-- Classement Global -->
    <template v-if="activeTab === 'global'">
      <div v-if="loadingGlobal" class="loading">Chargement…</div>
      <div v-else-if="globalEntries.length === 0" class="empty">
        Aucun joueur classé pour l'instant. Jouez une partie pour apparaître ici !
      </div>
      <table v-else class="lb-table">
        <thead>
          <tr>
            <th>#</th>
            <th>Joueur</th>
            <th>Niveau</th>
            <th>XP</th>
            <th>Parties</th>
            <th>Meilleur score</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="e in globalEntries"
            :key="e.user_id"
            :class="{
              'lb-gold':   e.rank === 1,
              'lb-silver': e.rank === 2,
              'lb-bronze': e.rank === 3,
              'lb-me':     e.username === ownUsername,
            }"
          >
            <td class="lb-rank">
              {{ e.rank === 1 ? '🥇' : e.rank === 2 ? '🥈' : e.rank === 3 ? '🥉' : `#${e.rank}` }}
            </td>
            <td class="lb-username">
              {{ e.username }}
              <span v-if="e.username === ownUsername" class="lb-you">vous</span>
            </td>
            <td><span class="lb-level">Niv. {{ e.level }}</span></td>
            <td class="lb-xp">{{ e.xp.toLocaleString() }} XP</td>
            <td>{{ e.total_games }}</td>
            <td>{{ e.best_score }} pts</td>
          </tr>
        </tbody>
      </table>
    </template>

    <!-- Classement Speed Run -->
    <template v-if="activeTab === 'speedrun'">
      <div v-if="loadingSpeedrun" class="loading">Chargement…</div>
      <div v-else-if="speedrunEntries.length === 0" class="empty">
        Personne n'a encore joué en Speed Run. Sois le premier !
      </div>
      <table v-else class="lb-table">
        <thead>
          <tr>
            <th>#</th>
            <th>Joueur</th>
            <th>Meilleur score</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="e in speedrunEntries"
            :key="e.user_id"
            :class="{
              'lb-gold':   e.rank === 1,
              'lb-silver': e.rank === 2,
              'lb-bronze': e.rank === 3,
              'lb-me':     e.username === ownUsername,
            }"
          >
            <td class="lb-rank">
              {{ e.rank === 1 ? '🥇' : e.rank === 2 ? '🥈' : e.rank === 3 ? '🥉' : `#${e.rank}` }}
            </td>
            <td class="lb-username">
              {{ e.username }}
              <span v-if="e.username === ownUsername" class="lb-you">vous</span>
            </td>
            <td class="lb-sr-score">{{ e.best_score }} <span class="lb-sr-unit">animes</span></td>
            <td class="lb-date">{{ formatDate(e.played_at) }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import { API_URL } from "../config";

const props = defineProps({
  ownUsername: { type: String, default: "" },
});

const activeTab = ref('global')
const globalEntries = ref([])
const speedrunEntries = ref([])
const loadingGlobal = ref(true)
const loadingSpeedrun = ref(false)

onMounted(async () => {
  try {
    const res = await fetch(`${API_URL}/api/leaderboard`);
    if (res.ok) globalEntries.value = await res.json();
  } finally {
    loadingGlobal.value = false;
  }
})

watch(activeTab, async (tab) => {
  if (tab === 'speedrun' && speedrunEntries.value.length === 0) {
    loadingSpeedrun.value = true
    try {
      const res = await fetch(`${API_URL}/api/leaderboard/speedrun`)
      if (res.ok) speedrunEntries.value = await res.json()
    } finally {
      loadingSpeedrun.value = false
    }
  }
})

function formatDate(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('fr-FR', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.leaderboard-page {
  max-width: 860px;
  margin: 0 auto;
  padding: 40px 24px;
  min-width: 0;
  width: 100%;
}

.lb-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.lb-tab {
  padding: 0.55rem 1.25rem;
  border-radius: 8px;
  border: 1px solid rgba(255,255,255,0.07);
  background: transparent;
  color: #94a3b8;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}
.lb-tab:hover { color: #f1f5f9; border-color: rgba(255,255,255,0.15); }
.lb-tab.active {
  background: rgba(249,115,22,0.15);
  color: #f97316;
  border-color: rgba(249,115,22,0.35);
}

.loading, .empty {
  text-align: center;
  color: #64748b;
  padding: 60px;
  font-style: italic;
}

.lb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
  background: #16213e;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(255,255,255,0.07);
}

.lb-table thead th {
  background: #0f0f23;
  color: #f97316;
  padding: 12px 16px;
  text-align: left;
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.lb-table tbody tr {
  border-bottom: 1px solid rgba(255,255,255,0.05);
  transition: background 0.15s;
  color: #cbd5e1;
}

.lb-table tbody tr:hover { background: rgba(255,255,255,0.04); }
.lb-table td { padding: 12px 16px; }

.lb-rank { font-size: 1.2rem; text-align: center; width: 52px; }
.lb-username { font-weight: 600; color: #f1f5f9; }

.lb-you {
  background: rgba(249,115,22,0.2);
  color: #f97316;
  font-size: 0.68rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 10px;
  margin-left: 6px;
  vertical-align: middle;
}

.lb-level {
  background: rgba(249,115,22,0.15);
  color: #fb923c;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 9px;
  border-radius: 10px;
}

.lb-xp { color: #f97316; font-weight: 700; }

.lb-sr-score { font-weight: 700; color: #f97316; }
.lb-sr-unit { font-size: 0.78rem; color: #64748b; font-weight: 400; }

.lb-date { color: #64748b; font-size: 0.82rem; }

.lb-gold   { background: rgba(255,215,0,0.07) !important; }
.lb-silver { background: rgba(192,192,192,0.06) !important; }
.lb-bronze { background: rgba(205,127,50,0.07) !important; }
.lb-me { outline: 2px solid #f97316; outline-offset: -2px; }

/* ── Mobile ── */
@media (max-width: 600px) {
  .leaderboard-page { padding: 24px 14px; }
  /* Tables larges : défilement horizontal interne plutôt que débordement de page. */
  .lb-table { display: block; overflow-x: auto; -webkit-overflow-scrolling: touch; white-space: nowrap; }
  .lb-table thead th, .lb-table td { padding: 10px 12px; }
}
</style>
