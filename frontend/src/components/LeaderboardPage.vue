<template>
  <div class="leaderboard-page">
    <h2>🏆 Classement Global</h2>

    <div v-if="loading" class="loading">Chargement…</div>

    <div v-else-if="entries.length === 0" class="empty">
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
          v-for="e in entries"
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
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { API_URL } from "../config";

const props = defineProps({
  ownUsername: { type: String, default: "" },
});

const entries = ref([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const res = await fetch(`${API_URL}/api/leaderboard`);
    if (res.ok) entries.value = await res.json();
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.leaderboard-page {
  max-width: 860px;
  margin: 0 auto;
  padding: 40px 24px;
}

.leaderboard-page h2 {
  font-size: 1.6rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 24px;
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

/* Podium */
.lb-gold   { background: rgba(255,215,0,0.07) !important; }
.lb-silver { background: rgba(192,192,192,0.06) !important; }
.lb-bronze { background: rgba(205,127,50,0.07) !important; }

.lb-me { outline: 2px solid #f97316; outline-offset: -2px; }
</style>
