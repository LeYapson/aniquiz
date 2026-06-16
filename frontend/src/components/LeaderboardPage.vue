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
  max-width: 800px;
  margin: 30px auto;
  padding: 0 20px;
}

.leaderboard-page h2 {
  font-size: 1.6rem;
  font-weight: bold;
  color: #1a1a2e;
  margin-bottom: 20px;
}

.loading, .empty {
  text-align: center;
  color: #888;
  padding: 40px;
}

.lb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}

.lb-table thead th {
  background: #1a1a2e;
  color: #f97316;
  padding: 10px 14px;
  text-align: left;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.lb-table tbody tr {
  border-bottom: 1px solid #eee;
  transition: background 0.15s;
}

.lb-table tbody tr:hover { background: #f9f9f9; }

.lb-table td { padding: 10px 14px; }

.lb-rank { font-size: 1.2rem; text-align: center; width: 52px; }

.lb-username { font-weight: 600; color: #1a1a2e; }

.lb-you {
  background: #1a1a2e;
  color: #f97316;
  font-size: 0.7rem;
  font-weight: bold;
  padding: 1px 7px;
  border-radius: 10px;
  margin-left: 6px;
  vertical-align: middle;
}

.lb-level {
  background: #fef3c7;
  color: #92400e;
  font-size: 0.78rem;
  font-weight: bold;
  padding: 2px 8px;
  border-radius: 10px;
}

.lb-xp { color: #f97316; font-weight: bold; }

/* Podium */
.lb-gold   { background: #fffbeb !important; }
.lb-silver { background: #f8fafc !important; }
.lb-bronze { background: #fff7ed !important; }

.lb-me { outline: 2px solid #f97316; outline-offset: -2px; }
</style>
