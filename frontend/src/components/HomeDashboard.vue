<template>
  <div class="dashboard">
    <div class="dashboard-inner">
      <!-- Bandeau d'accueil -->
      <div class="greeting-bar">
        <div class="greeting-text">
          <h1>{{ greeting }}, <span class="username">{{ user?.username ?? 'joueur' }}</span> !</h1>
          <p>Prêt à reconnaître quelques openings ?</p>
        </div>
        <div class="greeting-stats">
          <div class="stat">
            <strong>{{ user?.level ?? 1 }}</strong>
            <span>Niveau</span>
          </div>
          <div class="stat-sep"></div>
          <div class="stat">
            <strong>{{ user?.xp ?? 0 }}</strong>
            <span>XP</span>
          </div>
          <div class="stat-sep"></div>
          <div class="stat">
            <strong>{{ user?.games_played ?? 0 }}</strong>
            <span>Parties</span>
          </div>
        </div>
      </div>

      <!-- Contenu : salons + devlog -->
      <div class="content-grid">
        <section class="rooms-col">
          <RoomSelection
            @room-created="(d) => emit('room-created', d)"
            @room-joined="(d) => emit('room-joined', d)"
          />
        </section>

        <aside class="devlog-col" aria-label="Actualités du projet">
          <div class="devlog-header">
            <h2><span aria-hidden="true">📰</span> Dernières nouvelles</h2>
          </div>
          <article v-for="log in devlogs" :key="log.id" class="devlog-card">
            <div class="devlog-meta">
              <span class="devlog-tag" :class="`tag-${tagColor(log.tag)}`">{{ log.tag }}</span>
              <time class="devlog-date" :datetime="log.datetime">{{ log.date }}</time>
            </div>
            <h3 class="devlog-title">{{ log.title }}</h3>
            <p class="devlog-body" v-html="log.bodyHtml"></p>
          </article>
        </aside>
      </div>
    </div>

    <AppFooter />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import RoomSelection from './RoomSelection.vue';
import AppFooter from './AppFooter.vue';
import { authStore } from '../authStore';
import { useNews } from '../composables/useNews';

const emit = defineEmits(['room-created', 'room-joined']);

const user = computed(() => authStore.user);

const greeting = computed(() => {
  const hour = new Date().getHours();
  if (hour >= 6 && hour < 12) return 'Bonjour';
  if (hour >= 12 && hour < 18) return 'Bon après-midi';
  return 'Bonsoir';
});

const TAG_COLORS = {
  Feature: 'green',
  Annonce: 'blue',
  Fix: 'orange',
};
const tagColor = (tag) => TAG_COLORS[tag] ?? 'orange';

const { news: devlogs } = useNews();
</script>

<style scoped>
.dashboard {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #0f0f23;
  min-height: calc(100vh - 64px);
}

.dashboard-inner {
  flex: 1;
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
  padding: 32px 24px 56px;
}

/* ── Bandeau d'accueil ── */
.greeting-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 24px 28px;
  background: linear-gradient(135deg, rgba(249, 115, 22, 0.08), rgba(59, 130, 246, 0.05));
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 16px;
  margin-bottom: 32px;
}
.greeting-text h1 {
  font-size: 1.5rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0 0 4px;
}
.greeting-text .username { color: #f97316; }
.greeting-text p {
  color: #64748b;
  font-size: 0.92rem;
  margin: 0;
}

.greeting-stats {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
}
.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.stat strong {
  font-size: 1.4rem;
  font-weight: 800;
  color: #f97316;
  line-height: 1;
}
.stat span {
  font-size: 0.7rem;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.stat-sep {
  width: 1px;
  height: 32px;
  background: rgba(255, 255, 255, 0.1);
}

/* ── Grille contenu ── */
.content-grid {
  display: grid;
  grid-template-columns: 3fr 2fr;
  gap: 28px;
  align-items: start;
}

.rooms-col {
  min-width: 0;
}

/* ── Devlog ── */
.devlog-col {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.devlog-header h2 {
  font-size: 1rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 4px;
}
.devlog-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 12px;
  padding: 16px 18px;
  transition: border-color 0.15s, transform 0.15s;
}
.devlog-card:hover {
  border-color: rgba(249, 115, 22, 0.25);
  transform: translateY(-2px);
}
.devlog-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.devlog-tag {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 9px;
  border-radius: 20px;
}
.tag-green { background: rgba(34, 197, 94, 0.15); color: #4ade80; }
.tag-blue  { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
.tag-orange { background: rgba(249, 115, 22, 0.15); color: #fb923c; }
.devlog-date {
  font-size: 0.72rem;
  color: #475569;
}
.devlog-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: #e2e8f0;
  margin: 0 0 6px;
  line-height: 1.35;
}
.devlog-body {
  font-size: 0.82rem;
  color: #64748b;
  line-height: 1.5;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ── Responsive ── */
@media (max-width: 880px) {
  .content-grid { grid-template-columns: 1fr; }
  .greeting-bar { flex-direction: column; align-items: flex-start; }
}

@media (max-width: 500px) {
  .dashboard-inner { padding: 20px 14px 40px; }
  .greeting-stats { gap: 14px; }
  .greeting-text h1 { font-size: 1.25rem; }
}
</style>
