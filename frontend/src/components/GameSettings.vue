<template>
  <div class="settings-panel">

    <div class="settings-header">
      <h3>⚙️ Paramètres</h3>
      <transition name="fade-sync">
        <span v-if="synced" class="synced-badge">✓ Synchronisé</span>
      </transition>
    </div>

    <!-- ── Partie ──────────────────────────────────────────── -->
    <div class="settings-section">
      <p class="section-label">Partie</p>
      <div class="section-row">
        <div class="field">
          <span class="field-label">Rounds</span>
          <div class="num-ctrl">
            <button type="button" class="num-btn" @click="adj('maxRounds', -1, 1, 20)" :disabled="local.maxRounds <= 1">−</button>
            <span class="num-val">{{ local.maxRounds }}</span>
            <button type="button" class="num-btn" @click="adj('maxRounds', +1, 1, 20)" :disabled="local.maxRounds >= 20">+</button>
          </div>
        </div>
        <div class="field">
          <span class="field-label">Durée par round</span>
          <div class="num-ctrl">
            <button type="button" class="num-btn" @click="adj('roundDuration', -5, 10, 60)" :disabled="local.roundDuration <= 10">−</button>
            <span class="num-val">{{ local.roundDuration }}s</span>
            <button type="button" class="num-btn" @click="adj('roundDuration', +5, 10, 60)" :disabled="local.roundDuration >= 60">+</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Réponses ────────────────────────────────────────── -->
    <div class="settings-section">
      <p class="section-label">Réponses</p>
      <div class="guess-pills">
        <button
          v-for="m in guessModes"
          :key="m.value"
          type="button"
          class="guess-pill"
          :class="{ active: local.guessMode === m.value }"
          @click="local.guessMode = m.value"
        >{{ m.label }}</button>
      </div>
      <div class="field" style="margin-top:4px">
        <span class="field-label">Vies par joueur</span>
        <div class="guess-pills">
          <button
            v-for="l in livesModes" :key="l.value" type="button"
            class="guess-pill" :class="{ active: local.livesMode === l.value }"
            @click="local.livesMode = l.value"
          >{{ l.label }}</button>
        </div>
      </div>

      <div class="toggle-row" :class="{ disabled: local.guessMode === 'multiple' }">
        <span class="toggle-label">🔔 Mode buzzer
          <span class="toggle-hint">buzze avant de répondre</span>
        </span>
        <button
          type="button"
          class="switch"
          :class="{ on: local.buzzerMode }"
          :disabled="local.guessMode === 'multiple'"
          @click="local.buzzerMode = !local.buzzerMode"
          role="switch"
          :aria-checked="local.buzzerMode"
        ><span class="switch-thumb" /></button>
      </div>
    </div>

    <!-- ── Filtres ─────────────────────────────────────────── -->
    <div class="settings-section">
      <p class="section-label">Filtres</p>
      <div class="section-row">
        <div class="field">
          <span class="field-label">Génériques</span>
          <select v-model="local.filterType" class="field-select">
            <option value="">OP + ED</option>
            <option value="OP">Openings</option>
            <option value="ED">Endings</option>
          </select>
        </div>
        <div class="field">
          <span class="field-label">Décennie</span>
          <select v-model="local.decade" class="field-select">
            <option value="0">Toutes</option>
            <option value="1990">Années 90</option>
            <option value="2000">Années 2000</option>
            <option value="2010">Années 2010</option>
            <option value="2020">Années 2020</option>
          </select>
        </div>
      </div>

      <div class="toggle-row">
        <span class="toggle-label">
          🎌 Liste perso
          <span v-if="anyLinked" class="badge-linked">{{ linkedLabel }}</span>
          <span v-else class="badge-muted">non connecté</span>
          <span v-if="loadingIds" class="badge-muted">chargement…</span>
        </span>
        <button
          type="button"
          class="switch"
          :class="{ on: local.useAnilistFilter, loading: loadingIds }"
          :disabled="!anyLinked || loadingIds"
          @click="toggleAnilistFilter"
          role="switch"
          :aria-checked="local.useAnilistFilter"
        ><span class="switch-thumb" /></button>
      </div>
      <p v-if="local.useAnilistFilter && playableTracks > 0" class="info-ok">
        ✓ {{ playableAnime }} animes · {{ playableTracks }} titres disponibles
      </p>
      <p v-if="local.useAnilistFilter && local.filterMalIds.length > 0 && playableTracks === 0" class="info-warn">
        ⚠️ Aucun anime de ta liste n'est encore disponible — filtre ignoré.
      </p>
      <p v-if="!anyLinked" class="info-muted">
        Connecte AniList ou MAL dans ton profil pour filtrer par ta liste.
      </p>
    </div>

    <!-- ── Salon ───────────────────────────────────────────── -->
    <div class="settings-section">
      <p class="section-label">Salon</p>
      <div class="toggle-row">
        <span class="toggle-label">🔒 Salon privé</span>
        <button
          type="button"
          class="switch"
          :class="{ on: local.isPrivate }"
          @click="local.isPrivate = !local.isPrivate"
          role="switch"
          :aria-checked="local.isPrivate"
        ><span class="switch-thumb" /></button>
      </div>
      <input
        v-if="local.isPrivate"
        v-model="local.password"
        type="text"
        placeholder="Mot de passe du salon"
        class="password-input"
      />
    </div>

    <!-- ── Config partagée ────────────────────────────────── -->
    <div class="config-share">
      <button type="button" class="btn-ghost" @click="copyConfig">📋 Copier la config</button>
      <div class="import-row">
        <input
          v-model="importCode"
          type="text"
          placeholder="Coller un code…"
          class="config-input"
          @keyup.enter="importConfig"
        />
        <button type="button" class="btn-ghost" @click="importConfig" :disabled="!importCode.trim()">📥</button>
      </div>
    </div>

  </div>
</template>

<script setup>
import { reactive, ref, computed, watch } from 'vue';
import { authStore, apiFetch } from '../authStore';
import { API_URL } from '../config';
import { useToast } from '../composables/useToast';

const toast = useToast();

const props = defineProps({
  socket: Object,
  initialSettings: Object,
});

const guessModes = [
  { value: 'anime',    label: 'Anime' },
  { value: 'title',   label: 'Titre' },
  { value: 'artist',  label: 'Artiste' },
  { value: 'multiple', label: '🔀 QCM' },
];

const livesModes = [
  { value: 0, label: 'Désactivé' },
  { value: 3, label: '❤️ 3 vies' },
  { value: 5, label: '❤️ 5 vies' },
];

const anilistLinked = computed(() => !!authStore.user?.anilist_username);
const malLinked     = computed(() => !!authStore.user?.mal_username);
const anyLinked     = computed(() => anilistLinked.value || malLinked.value);
const linkedLabel   = computed(() => {
  const parts = [];
  if (anilistLinked.value) parts.push('AniList');
  if (malLinked.value)     parts.push('MAL');
  return parts.join(' + ');
});

const loadingIds    = ref(false);
const playableAnime = ref(0);
const playableTracks = ref(0);
const synced        = ref(false);
const importCode    = ref('');
let syncTimer       = null;
let applyTimer      = null;

const local = reactive({
  maxRounds:        props.initialSettings?.maxRounds      ?? 5,
  roundDuration:    props.initialSettings?.roundDuration  ?? 20,
  filterType:       props.initialSettings?.filterType     ?? '',
  guessMode:        props.initialSettings?.guessMode      ?? 'anime',
  decade:           props.initialSettings?.decade         ?? 0,
  isPrivate:        props.initialSettings?.isPrivate      ?? false,
  password:         props.initialSettings?.password       ?? '',
  buzzerMode:       props.initialSettings?.buzzerMode     ?? false,
  livesMode:        props.initialSettings?.livesMode      ?? 0,
  useAnilistFilter: false,
  filterMalIds:     [],
});

watch(() => props.initialSettings, (s) => {
  if (s) Object.assign(local, s);
}, { deep: true });

// ── Auto-apply ────────────────────────────────────────────────────────────────
// Immédiat pour les toggles et selects, débouncé pour les steppers numériques.

watch(() => [local.filterType, local.guessMode, local.decade, local.buzzerMode, local.isPrivate, local.password, local.livesMode], apply);
watch(() => [local.maxRounds, local.roundDuration], () => {
  clearTimeout(applyTimer);
  applyTimer = setTimeout(apply, 400);
});

// Mode buzzer incompatible avec QCM : on le coupe automatiquement.
watch(() => local.guessMode, (v) => {
  if (v === 'multiple') local.buzzerMode = false;
});

function adj(field, delta, min, max) {
  local[field] = Math.min(max, Math.max(min, local[field] + delta));
}

function apply() {
  if (!props.socket || props.socket.readyState !== WebSocket.OPEN) return;
  const minYear = local.decade > 0 ? Number(local.decade) : 0;
  const maxYear = local.decade > 0 ? Number(local.decade) + 9 : 0;
  props.socket.send(JSON.stringify({
    type: 'UPDATE_SETTINGS',
    payload: {
      max_rounds:      local.maxRounds,
      round_duration:  local.roundDuration,
      filter_type:     local.filterType,
      min_year:        minYear,
      max_year:        maxYear,
      is_private:      local.isPrivate,
      password:        local.password,
      buzzer_mode:     local.buzzerMode,
      guess_mode:      local.guessMode,
      filter_mal_ids:  local.useAnilistFilter ? local.filterMalIds : [],
      lives_mode:      local.livesMode,
    },
  }));
  synced.value = true;
  clearTimeout(syncTimer);
  syncTimer = setTimeout(() => { synced.value = false; }, 2000);
}

// ── AniList / MAL filter ──────────────────────────────────────────────────────
const toggleAnilistFilter = async () => {
  if (!anyLinked.value) return;
  local.useAnilistFilter = !local.useAnilistFilter;
  if (local.useAnilistFilter && local.filterMalIds.length === 0) {
    loadingIds.value = true;
    try {
      const res = await apiFetch(`${API_URL}/api/me/anime-ids`);
      if (res.ok) {
        const data = await res.json();
        const ids = Array.isArray(data) ? data : (data.ids ?? []);
        local.filterMalIds   = ids;
        playableAnime.value  = data.playable_anime  ?? ids.length;
        playableTracks.value = data.playable_tracks ?? 0;
      }
    } catch {
      local.useAnilistFilter = false;
    } finally {
      loadingIds.value = false;
    }
  } else if (!local.useAnilistFilter) {
    local.filterMalIds   = [];
    playableAnime.value  = 0;
    playableTracks.value = 0;
  }
  apply();
};

// ── Config partagée ───────────────────────────────────────────────────────────
const CONFIG_KEYS    = ['maxRounds', 'roundDuration', 'filterType', 'guessMode', 'decade', 'buzzerMode'];
const ALLOWED_DECADES = ['0', '1990', '2000', '2010', '2020'];
const clamp = (n, min, max) => Math.min(Math.max(n, min), max);

const copyConfig = async () => {
  const cfg = {};
  for (const k of CONFIG_KEYS) cfg[k] = local[k];
  const code = btoa(unescape(encodeURIComponent(JSON.stringify(cfg))));
  try {
    await navigator.clipboard.writeText(code);
    toast.success('Config copiée !');
  } catch {
    toast.info(code, { title: 'Copie manuelle' });
  }
};

const importConfig = () => {
  const raw = importCode.value.trim();
  if (!raw) return;
  let cfg;
  try {
    cfg = JSON.parse(decodeURIComponent(escape(atob(raw))));
  } catch {
    toast.error('Code de config invalide');
    return;
  }
  if (typeof cfg.maxRounds    === 'number') local.maxRounds    = clamp(Math.round(cfg.maxRounds), 1, 20);
  if (typeof cfg.roundDuration === 'number') local.roundDuration = clamp(Math.round(cfg.roundDuration), 10, 60);
  if (['', 'OP', 'ED'].includes(cfg.filterType)) local.filterType = cfg.filterType;
  if (['anime', 'title', 'artist', 'multiple'].includes(cfg.guessMode)) local.guessMode = cfg.guessMode;
  if (ALLOWED_DECADES.includes(String(cfg.decade))) local.decade = String(cfg.decade);
  if (typeof cfg.buzzerMode === 'boolean') local.buzzerMode = cfg.buzzerMode;
  importCode.value = '';
  apply();
  toast.success('Config importée !');
};
</script>

<style scoped>
.settings-panel {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 14px;
  padding: 18px 20px;
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* ── Header ── */
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.settings-header h3 {
  margin: 0;
  font-size: 0.92rem;
  font-weight: 700;
  color: #f1f5f9;
}
.synced-badge {
  font-size: 0.72rem;
  font-weight: 700;
  color: #4ade80;
  background: rgba(74,222,128,0.1);
  border: 1px solid rgba(74,222,128,0.25);
  padding: 2px 9px;
  border-radius: 99px;
}
.fade-sync-enter-active, .fade-sync-leave-active { transition: opacity 0.3s; }
.fade-sync-enter-from, .fade-sync-leave-to { opacity: 0; }

/* ── Section ── */
.settings-section {
  padding: 14px 0;
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.section-label {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #334155;
  margin: 0 0 2px;
}

/* ── Champ numérique ── */
.section-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
}
.num-ctrl {
  display: flex;
  align-items: center;
  gap: 0;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  overflow: hidden;
}
.num-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 1.1rem;
  width: 34px;
  height: 36px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  flex-shrink: 0;
}
.num-btn:hover:not(:disabled) { background: rgba(249,115,22,0.1); color: #f97316; }
.num-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.num-val {
  flex: 1;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 700;
  color: #f1f5f9;
}

/* ── Pill buttons (guess mode) ── */
.guess-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.guess-pill {
  padding: 5px 13px;
  border-radius: 99px;
  font-size: 0.78rem;
  font-weight: 600;
  border: 1px solid rgba(255,255,255,0.1);
  background: transparent;
  color: #64748b;
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}
.guess-pill:hover { color: #f1f5f9; border-color: rgba(255,255,255,0.25); }
.guess-pill.active {
  background: rgba(249,115,22,0.15);
  color: #f97316;
  border-color: rgba(249,115,22,0.4);
}

/* ── Toggle row ── */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toggle-row.disabled { opacity: 0.4; pointer-events: none; }
.toggle-label {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 0.82rem;
  font-weight: 600;
  color: #cbd5e1;
  flex-wrap: wrap;
}
.toggle-hint {
  font-size: 0.7rem;
  color: #475569;
  font-weight: 400;
}

/* ── Select ── */
.field-select {
  padding: 7px 10px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9;
  border-radius: 8px;
  font-size: 0.83rem;
  outline: none;
  transition: border-color 0.15s;
  cursor: pointer;
}
.field-select:focus { border-color: #f97316; }
.field-select option { background: #1e2a45; }

/* ── Badges ── */
.badge-linked { font-size: 0.68rem; background: rgba(52,211,153,0.15); color: #34d399; padding: 2px 7px; border-radius: 99px; font-weight: 600; }
.badge-muted  { font-size: 0.68rem; background: rgba(100,116,139,0.1);  color: #64748b;  padding: 2px 7px; border-radius: 99px; font-weight: 600; }

/* ── Infos ── */
.info-ok   { margin: -4px 0 0; font-size: 0.75rem; color: #34d399; }
.info-warn { margin: -4px 0 0; font-size: 0.75rem; color: #fbbf24; }
.info-muted{ margin: -4px 0 0; font-size: 0.75rem; color: #475569; font-style: italic; }

/* ── Password ── */
.password-input {
  padding: 7px 10px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9;
  border-radius: 8px;
  font-size: 0.83rem;
  outline: none;
  width: 100%;
  box-sizing: border-box;
}
.password-input:focus { border-color: #f97316; }

/* ── Switch ── */
.switch {
  flex-shrink: 0;
  position: relative;
  width: 44px; height: 24px;
  border-radius: 99px;
  border: none;
  background: #334155;
  cursor: pointer;
  padding: 0;
  transition: background 0.2s;
  outline: none;
}
.switch:focus-visible { box-shadow: 0 0 0 2px #f97316; }
.switch.on      { background: #f97316; }
.switch.loading { opacity: 0.6; cursor: wait; }
.switch:disabled { opacity: 0.35; cursor: not-allowed; }
.switch-thumb {
  position: absolute;
  top: 3px; left: 3px;
  width: 18px; height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s cubic-bezier(.4,0,.2,1);
  box-shadow: 0 1px 3px rgba(0,0,0,0.4);
  pointer-events: none;
}
.switch.on .switch-thumb { transform: translateX(20px); }

/* ── Config share ── */
.config-share {
  padding-top: 14px;
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.import-row { display: flex; gap: 6px; }
.config-input {
  flex: 1; padding: 7px 10px;
  background: #0f0f23; border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9; border-radius: 8px; font-size: 0.8rem; outline: none;
}
.config-input:focus { border-color: #f97316; }
.config-input::placeholder { color: #475569; }
.btn-ghost {
  background: rgba(255,255,255,0.05);
  color: #94a3b8;
  border: 1px solid rgba(255,255,255,0.1);
  padding: 7px 12px; border-radius: 8px; cursor: pointer;
  font-size: 0.8rem; font-weight: 600; white-space: nowrap;
  transition: background 0.15s, color 0.15s;
}
.btn-ghost:hover:not(:disabled) { background: rgba(255,255,255,0.1); color: #f1f5f9; }
.btn-ghost:disabled { opacity: 0.35; cursor: not-allowed; }

@media (max-width: 480px) {
  .section-row { grid-template-columns: 1fr; }
  .import-row  { flex-direction: column; }
}
</style>
