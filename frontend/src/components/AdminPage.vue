<template>
  <div class="admin-page">
    <h2>🛠️ Administration</h2>
    <p class="admin-sub">Outils de gestion de la librairie. Réservé aux administrateurs.</p>

    <!-- Données de la librairie -->
    <div class="admin-card">
      <h3>📊 Données de la librairie</h3>
      <div v-if="stats" class="stat-row">
        <span><strong>{{ stats.tracks }}</strong> pistes</span>
        <span><strong>{{ stats.playable_tracks }}</strong> jouables</span>
        <span><strong>{{ stats.animes }}</strong> animes</span>
        <span><strong>{{ stats.with_alt_titles }}</strong> avec titres alt.</span>
        <span><strong>{{ stats.users }}</strong> joueurs</span>
      </div>
      <div class="admin-row">
        <input
          v-model="trackQuery"
          type="text"
          placeholder="Rechercher un anime ou un titre…"
          class="search-input"
          @keyup.enter="searchTracks"
        />
        <button class="btn-run" @click="searchTracks">Rechercher</button>
      </div>
      <div v-if="tracks.length" class="tracks-table-wrap">
        <table class="tracks-table">
          <thead>
            <tr><th>Anime</th><th>Titres alt.</th><th>Musique</th><th>Type</th><th>Année</th><th>Audio</th></tr>
          </thead>
          <tbody>
            <tr v-for="t in tracks" :key="t.id">
              <td>{{ t.anime_name }}</td>
              <td>
                <span v-if="t.alt_titles && t.alt_titles.length" class="alt-ok">{{ t.alt_titles.join(', ') }}</span>
                <span v-else class="alt-none">—</span>
              </td>
              <td>{{ t.title }}<span v-if="t.artist" class="muted"> · {{ t.artist }}</span></td>
              <td>{{ t.track_type }}</td>
              <td>{{ t.anime_year || '—' }}</td>
              <td>{{ t.has_audio ? '✓' : '✗' }}</td>
            </tr>
          </tbody>
        </table>
        <div class="pager">
          <button class="btn-ghost" :disabled="trackOffset === 0" @click="prevTracks">← Précédent</button>
          <span class="muted">{{ trackOffset + 1 }}–{{ trackOffset + tracks.length }}</span>
          <button class="btn-ghost" :disabled="tracks.length < trackLimit" @click="nextTracks">Suivant →</button>
        </div>
      </div>
      <p v-else-if="tracksLoaded" class="empty-tracks">Aucune piste trouvée.</p>
    </div>

    <!-- Import en masse -->
    <div class="admin-card">
      <h3>📥 Import en masse</h3>
      <p class="card-desc">
        Importe les animes les plus populaires de MyAnimeList (25 par page).
        Maximise le recouvrement avec les listes perso des joueurs.
      </p>
      <div class="admin-row">
        <label>
          Pages
          <input type="number" v-model.number="seedPages" min="1" max="40" :disabled="seed.running" />
        </label>
        <button class="btn-run" @click="startSeed" :disabled="seed.running">
          {{ seed.running ? 'En cours…' : 'Lancer l\'import' }}
        </button>
      </div>
      <div v-if="seed.total || seed.running || seed.finished_at" class="progress">
        <div class="progress-bar"><div class="progress-fill" :style="{ width: seedPct + '%' }"></div></div>
        <div class="progress-stats">
          <span>{{ seed.processed }}/{{ seed.total }} traités</span>
          <span class="ok">✓ {{ seed.imported }} importés</span>
          <span class="muted">{{ seed.skipped }} ignorés</span>
          <span v-if="seed.failed" class="err">{{ seed.failed }} échecs</span>
          <span v-if="!seed.running && seed.finished_at" class="done">— terminé</span>
        </div>
        <p v-if="seed.last_error" class="last-err">⚠️ {{ seed.last_error }}</p>
      </div>
    </div>

    <!-- Vérification des liens audio -->
    <div class="admin-card">
      <h3>🔊 Vérification des liens audio</h3>
      <p class="card-desc">
        Vérifie chaque URL audio ; les liens morts (404/410) sont retirés du jeu.
        Les erreurs passagères (timeout) sont laissées intactes.
      </p>
      <button class="btn-run" @click="startAudio" :disabled="audio.running">
        {{ audio.running ? 'Vérification…' : 'Vérifier les liens' }}
      </button>
      <div v-if="audio.total || audio.running || audio.finished_at" class="progress">
        <div class="progress-bar"><div class="progress-fill" :style="{ width: audioPct + '%' }"></div></div>
        <div class="progress-stats">
          <span>{{ audio.checked }}/{{ audio.total }} vérifiés</span>
          <span class="err">{{ audio.dead }} morts</span>
          <span class="muted">{{ audio.unreachable }} injoignables</span>
          <span v-if="!audio.running && audio.finished_at" class="done">— terminé</span>
        </div>
        <p v-if="audio.last_error" class="last-err">⚠️ {{ audio.last_error }}</p>
      </div>
    </div>

    <!-- Backfill des titres alternatifs -->
    <div class="admin-card">
      <h3>🌐 Titres alternatifs (anglais)</h3>
      <p class="card-desc">
        Recharge les titres anglais et synonymes des animes <strong>déjà</strong> en base
        (l'import en masse ignore les animes existants). À lancer une fois pour que les
        réponses en anglais soient acceptées sur l'ancienne bibliothèque.
      </p>
      <button class="btn-run" @click="startBackfill" :disabled="backfill.running">
        {{ backfill.running ? 'En cours…' : 'Recharger les titres' }}
      </button>
      <div v-if="backfill.total || backfill.running || backfill.finished_at" class="progress">
        <div class="progress-bar"><div class="progress-fill" :style="{ width: backfillPct + '%' }"></div></div>
        <div class="progress-stats">
          <span>{{ backfill.processed }}/{{ backfill.total }} traités</span>
          <span class="ok">✓ {{ backfill.updated }} mis à jour</span>
          <span v-if="backfill.failed" class="err">{{ backfill.failed }} échecs</span>
          <span v-if="!backfill.running && backfill.finished_at" class="done">— terminé</span>
        </div>
        <p v-if="backfill.last_error" class="last-err">⚠️ {{ backfill.last_error }}</p>
      </div>
    </div>

    <!-- Annonces Discord -->
    <div class="admin-card">
      <h3>📣 Annonces Discord</h3>
      <p class="card-desc">
        Publie une news sur le salon Discord (webhook). Choisis l'annonce à diffuser.
      </p>
      <div v-for="n in news" :key="n.id" class="announce-row">
        <span class="announce-info">
          <span class="announce-tag" :class="`tag-${(n.tag || '').toLowerCase()}`">{{ n.tag }}</span>
          {{ n.title }} <small>· {{ n.date }}</small>
        </span>
        <button class="btn-ghost" :disabled="announcing === n.id" @click="announce(n)">
          {{ announcing === n.id ? '…' : '📣 Publier' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from "vue";
import { authStore } from "../authStore";
import { API_URL } from "../config";
import { useToast } from "../composables/useToast";
import { useNews } from "../composables/useNews";

const toast = useToast();
const { news } = useNews();
const announcing = ref(null);

const announce = async (n) => {
  announcing.value = n.id;
  try {
    const res = await authFetch(`${API_URL}/api/admin/announce`, {
      method: "POST",
      body: JSON.stringify({ title: n.title, content: n.content, tag: n.tag, date: n.date }),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) toast.success("Publié sur Discord !");
    else toast.error(data.error || "Échec de la publication");
  } catch {
    toast.error("Erreur réseau");
  } finally {
    announcing.value = null;
  }
};

const seedPages = ref(4);
const seed = reactive({ running: false, pages: 0, total: 0, processed: 0, imported: 0, skipped: 0, failed: 0, finished_at: "", last_error: "" });
const audio = reactive({ running: false, total: 0, checked: 0, dead: 0, unreachable: 0, finished_at: "", last_error: "" });
const backfill = reactive({ running: false, total: 0, processed: 0, updated: 0, failed: 0, finished_at: "", last_error: "" });

let seedTimer = null;
let audioTimer = null;
let backfillTimer = null;

const seedPct = computed(() => (seed.total ? Math.round((seed.processed / seed.total) * 100) : 0));
const audioPct = computed(() => (audio.total ? Math.round((audio.checked / audio.total) * 100) : 0));
const backfillPct = computed(() => (backfill.total ? Math.round((backfill.processed / backfill.total) * 100) : 0));

const authFetch = (url, opts = {}) =>
  fetch(url, { ...opts, headers: { ...authStore.authHeaders(), ...(opts.headers || {}) } });

// ── Données de la librairie (lecture seule) ──
const stats = ref(null);
const trackQuery = ref("");
const tracks = ref([]);
const tracksLoaded = ref(false);
const trackOffset = ref(0);
const trackLimit = 25;

const loadStats = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/stats`);
    if (res.ok) stats.value = await res.json();
  } catch { /* ignore */ }
};

const loadTracks = async () => {
  try {
    const url = `${API_URL}/api/admin/tracks?q=${encodeURIComponent(trackQuery.value)}&limit=${trackLimit}&offset=${trackOffset.value}`;
    const res = await authFetch(url);
    if (res.ok) tracks.value = await res.json();
  } catch { /* ignore */ } finally {
    tracksLoaded.value = true;
  }
};

const searchTracks = () => { trackOffset.value = 0; loadTracks(); };
const nextTracks = () => { trackOffset.value += trackLimit; loadTracks(); };
const prevTracks = () => { trackOffset.value = Math.max(0, trackOffset.value - trackLimit); loadTracks(); };

const pollSeed = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/seed/status`);
    if (res.ok) Object.assign(seed, await res.json());
  } catch { /* ignore */ }
  if (!seed.running && seedTimer) { clearInterval(seedTimer); seedTimer = null; }
};

const startSeed = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/seed`, {
      method: "POST",
      body: JSON.stringify({ pages: seedPages.value }),
    });
    if (res.status === 403) { toast.error("Accès réservé aux administrateurs"); return; }
    const data = await res.json().catch(() => ({}));
    if (res.ok || res.status === 409) {
      if (data.progress) Object.assign(seed, data.progress);
      if (!seedTimer) seedTimer = setInterval(pollSeed, 2000);
      toast.info(res.status === 409 ? "Un import est déjà en cours" : "Import démarré");
    } else {
      toast.error(data.error || "Échec du démarrage");
    }
  } catch { toast.error("Erreur réseau"); }
};

const pollAudio = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/audio/healthcheck/status`);
    if (res.ok) Object.assign(audio, await res.json());
  } catch { /* ignore */ }
  if (!audio.running && audioTimer) { clearInterval(audioTimer); audioTimer = null; }
};

const startAudio = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/audio/healthcheck`, { method: "POST" });
    if (res.status === 403) { toast.error("Accès réservé aux administrateurs"); return; }
    const data = await res.json().catch(() => ({}));
    if (res.ok || res.status === 409) {
      if (data.progress) Object.assign(audio, data.progress);
      if (!audioTimer) audioTimer = setInterval(pollAudio, 2000);
      toast.info(res.status === 409 ? "Une vérification est déjà en cours" : "Vérification démarrée");
    } else {
      toast.error(data.error || "Échec du démarrage");
    }
  } catch { toast.error("Erreur réseau"); }
};

const pollBackfill = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/backfill-titles/status`);
    if (res.ok) Object.assign(backfill, await res.json());
  } catch { /* ignore */ }
  if (!backfill.running && backfillTimer) { clearInterval(backfillTimer); backfillTimer = null; }
};

const startBackfill = async () => {
  try {
    const res = await authFetch(`${API_URL}/api/admin/backfill-titles`, { method: "POST" });
    if (res.status === 403) { toast.error("Accès réservé aux administrateurs"); return; }
    const data = await res.json().catch(() => ({}));
    if (res.ok || res.status === 409) {
      if (data.progress) Object.assign(backfill, data.progress);
      if (!backfillTimer) backfillTimer = setInterval(pollBackfill, 2000);
      toast.info(res.status === 409 ? "Un backfill est déjà en cours" : "Backfill démarré");
    } else {
      toast.error(data.error || "Échec du démarrage");
    }
  } catch { toast.error("Erreur réseau"); }
};

// Reprend l'affichage si un job tourne déjà à l'ouverture du panneau.
onMounted(() => { pollSeed(); pollAudio(); pollBackfill(); loadStats(); loadTracks(); });
onUnmounted(() => {
  if (seedTimer) clearInterval(seedTimer);
  if (audioTimer) clearInterval(audioTimer);
  if (backfillTimer) clearInterval(backfillTimer);
});
</script>

<style scoped>
.admin-page { max-width: 760px; margin: 0 auto; padding: 40px 24px; }
.admin-page h2 { color: #f1f5f9; margin: 0 0 4px; }
.admin-sub { color: #64748b; font-size: 0.9rem; margin: 0 0 28px; }

.admin-card {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 14px;
  padding: 20px 22px;
  margin-bottom: 20px;
}
.admin-card h3 { color: #f1f5f9; font-size: 1rem; margin: 0 0 6px; }
.card-desc { color: #94a3b8; font-size: 0.84rem; margin: 0 0 14px; line-height: 1.5; }

/* ── Annonces Discord ── */
.announce-row {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding: 9px 0; border-bottom: 1px solid rgba(255,255,255,0.05);
}
.announce-row:last-child { border-bottom: none; }
.announce-info { color: #e2e8f0; font-size: 0.86rem; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.announce-info small { color: #64748b; }
.announce-tag { font-size: 0.66rem; font-weight: 700; text-transform: uppercase; padding: 2px 8px; border-radius: 20px; }
.tag-feature { background: rgba(34,197,94,0.15); color: #4ade80; }
.tag-annonce { background: rgba(59,130,246,0.15); color: #60a5fa; }
.tag-fix { background: rgba(249,115,22,0.15); color: #fb923c; }

/* ── Données de la librairie ── */
.stat-row { display: flex; flex-wrap: wrap; gap: 16px; margin: 8px 0 14px; font-size: 0.85rem; color: #cbd5e1; }
.stat-row strong { color: #f97316; font-size: 1.05rem; }
.search-input {
  flex: 1; padding: 8px 11px;
  background: #0f0f23; border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9; border-radius: 7px; font-size: 0.88rem; outline: none;
}
.search-input:focus { border-color: #f97316; }
.tracks-table-wrap { margin-top: 14px; overflow-x: auto; }
.tracks-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
.tracks-table th {
  text-align: left; color: #f97316; font-size: 0.7rem; text-transform: uppercase;
  letter-spacing: 0.04em; padding: 8px 10px; border-bottom: 1px solid rgba(255,255,255,0.1); white-space: nowrap;
}
.tracks-table td { padding: 8px 10px; border-bottom: 1px solid rgba(255,255,255,0.05); color: #e2e8f0; vertical-align: top; }
.tracks-table .muted { color: #64748b; }
.alt-ok { color: #34d399; }
.alt-none { color: #475569; }
.pager { display: flex; align-items: center; justify-content: center; gap: 16px; margin-top: 14px; font-size: 0.82rem; color: #94a3b8; }
.btn-ghost {
  background: rgba(255,255,255,0.05); color: #cbd5e1;
  border: 1px solid rgba(255,255,255,0.12); padding: 6px 12px; border-radius: 7px;
  font-weight: 600; font-size: 0.82rem; cursor: pointer;
}
.btn-ghost:hover:not(:disabled) { background: rgba(255,255,255,0.1); }
.btn-ghost:disabled { opacity: 0.4; cursor: not-allowed; }
.empty-tracks { color: #64748b; font-style: italic; font-size: 0.86rem; margin-top: 10px; }

.admin-row { display: flex; align-items: flex-end; gap: 12px; }
.admin-row label { display: flex; flex-direction: column; gap: 6px; font-size: 0.82rem; color: #94a3b8; font-weight: 600; }
.admin-row input {
  width: 90px; padding: 8px 10px;
  background: #0f0f23; border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9; border-radius: 7px; font-size: 0.9rem; outline: none;
}
.admin-row input:focus { border-color: #f97316; }

.btn-run {
  background: #f97316; color: #fff; border: none;
  padding: 9px 18px; border-radius: 8px; font-weight: 700; font-size: 0.9rem; cursor: pointer;
  transition: opacity 0.15s;
}
.btn-run:hover:not(:disabled) { opacity: 0.88; }
.btn-run:disabled { background: #334155; color: #64748b; cursor: not-allowed; }

.progress { margin-top: 16px; }
.progress-bar { height: 8px; background: #0f0f23; border-radius: 4px; overflow: hidden; }
.progress-fill { height: 100%; background: linear-gradient(90deg, #f97316, #fb923c); border-radius: 4px; transition: width 0.4s ease; }
.progress-stats { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 10px; font-size: 0.82rem; color: #cbd5e1; }
.progress-stats .ok { color: #22c55e; }
.progress-stats .err { color: #ef4444; }
.progress-stats .muted { color: #64748b; }
.progress-stats .done { color: #f97316; font-weight: 700; }
.last-err { color: #fbbf24; font-size: 0.78rem; margin: 8px 0 0; }

@media (max-width: 600px) {
  .admin-page { padding: 24px 14px; }
  .admin-row { flex-direction: column; align-items: stretch; }
  .btn-run { width: 100%; }
}
</style>
