<template>
  <div class="settings-panel">
    <h3>⚙️ Paramètres de la partie</h3>

    <div class="settings-grid">
      <label>
        Rounds
        <input type="number" v-model.number="local.maxRounds" min="1" max="20" />
      </label>

      <label>
        Durée par round (s)
        <input type="number" v-model.number="local.roundDuration" min="10" max="60" />
      </label>

      <label>
        Type de générique
        <select v-model="local.filterType">
          <option value="">Tout (OP + ED)</option>
          <option value="OP">Openings uniquement</option>
          <option value="ED">Endings uniquement</option>
        </select>
      </label>

      <label>
        Décennie
        <select v-model="local.decade">
          <option value="0">Toutes</option>
          <option value="1990">Années 90</option>
          <option value="2000">Années 2000</option>
          <option value="2010">Années 2010</option>
          <option value="2020">Années 2020</option>
        </select>
      </label>

      <div class="full-width toggle-row">
        <span class="toggle-label">
          🎌 Filtrer par ma liste perso
          <span v-if="anyLinked" class="badge-linked">{{ linkedLabel }}</span>
          <span v-else class="badge-unlinked">non connecté</span>
          <span v-if="loadingIds" class="badge-loading">chargement…</span>
        </span>
        <button
          class="switch"
          :class="{ on: local.useAnilistFilter, loading: loadingIds }"
          :disabled="!anyLinked || loadingIds"
          @click="toggleAnilistFilter"
          type="button"
          :aria-checked="local.useAnilistFilter"
          role="switch"
        >
          <span class="switch-thumb" />
        </button>
      </div>
      <p v-if="local.useAnilistFilter && local.filterMalIds.length > 0" class="info-ids">
        ✓ {{ local.filterMalIds.length }} animés chargés depuis ta liste
      </p>
      <p v-if="!anyLinked" class="info-unlinked">
        Connecte ton compte AniList ou MAL dans ton profil pour activer ce filtre.
      </p>

      <label class="full-width">
        <input type="checkbox" v-model="local.isPrivate" /> Salon privé
        <input
          v-if="local.isPrivate"
          v-model="local.password"
          type="text"
          placeholder="Mot de passe"
          class="password-input"
        />
      </label>
    </div>

    <button @click="apply" class="btn-apply" :disabled="loadingIds">
      {{ loadingIds ? 'Chargement…' : 'Appliquer' }}
    </button>
  </div>
</template>

<script setup>
import { reactive, ref, computed, watch } from "vue";
import { authStore } from "../authStore";
import { API_URL } from "../config";

const props = defineProps({
  socket: Object,
  initialSettings: Object,
});

// Vérifie si l'utilisateur a au moins un compte de tracking lié (AniList ou MAL)
const anilistLinked = computed(() => !!authStore.user?.anilist_username);
const malLinked     = computed(() => !!authStore.user?.mal_username);
const anyLinked     = computed(() => anilistLinked.value || malLinked.value);
const linkedLabel   = computed(() => {
  const parts = [];
  if (anilistLinked.value) parts.push('AniList');
  if (malLinked.value)     parts.push('MAL');
  return parts.join(' + ');
});

const loadingIds = ref(false);

const local = reactive({
  maxRounds: props.initialSettings?.maxRounds ?? 5,
  roundDuration: props.initialSettings?.roundDuration ?? 20,
  filterType: props.initialSettings?.filterType ?? "",
  decade: props.initialSettings?.decade ?? 0,
  isPrivate: props.initialSettings?.isPrivate ?? false,
  password: props.initialSettings?.password ?? "",
  useAnilistFilter: false,
  filterMalIds: [],
});

watch(() => props.initialSettings, (s) => {
  if (!s) return;
  Object.assign(local, s);
}, { deep: true });

const toggleAnilistFilter = async () => {
  if (!anyLinked.value) return;
  local.useAnilistFilter = !local.useAnilistFilter;

  if (local.useAnilistFilter) {
    if (local.filterMalIds.length === 0) {
      loadingIds.value = true;
      try {
        const res = await fetch(`${API_URL}/api/me/anime-ids`, {
          headers: authStore.authHeaders(),
        });
        if (res.ok) {
          const data = await res.json();
          local.filterMalIds = Array.isArray(data) ? data : [];
        }
      } catch (e) {
        console.error("Erreur récupération liste perso :", e);
        local.useAnilistFilter = false;
      } finally {
        loadingIds.value = false;
      }
    }
  } else {
    local.filterMalIds = [];
  }

  // Auto-appliquer immédiatement sans avoir à cliquer "Appliquer"
  apply();
};

const apply = () => {
  if (!props.socket || props.socket.readyState !== WebSocket.OPEN) return;

  const minYear = local.decade > 0 ? Number(local.decade) : 0;
  const maxYear = local.decade > 0 ? Number(local.decade) + 9 : 0;

  props.socket.send(JSON.stringify({
    type: "UPDATE_SETTINGS",
    payload: {
      max_rounds: local.maxRounds,
      round_duration: local.roundDuration,
      filter_type: local.filterType,
      min_year: minYear,
      max_year: maxYear,
      is_private: local.isPrivate,
      password: local.password,
      filter_mal_ids: local.useAnilistFilter ? local.filterMalIds : [],
    },
  }));
};
</script>

<style scoped>
.settings-panel {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 12px;
  padding: 20px;
  margin-top: 16px;
}
.settings-panel h3 {
  margin: 0 0 16px;
  font-size: 0.95rem;
  color: #f1f5f9;
  font-weight: 700;
}
.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
}
.settings-grid label {
  display: flex;
  flex-direction: column;
  font-size: 0.82rem;
  font-weight: 600;
  color: #94a3b8;
  gap: 6px;
}
.settings-grid input[type="number"],
.settings-grid select {
  padding: 7px 10px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9;
  border-radius: 7px;
  font-size: 0.88rem;
  outline: none;
  transition: border-color 0.15s;
}
.settings-grid input[type="number"]:focus,
.settings-grid select:focus { border-color: #f97316; }
.settings-grid select option { background: #1e2a45; }

.full-width { grid-column: 1 / -1; }
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.toggle-label { display: flex; align-items: center; gap: 8px; font-weight: 600; color: #e2e8f0; font-size: 0.85rem; }
.badge-linked   { font-size: 0.7rem; background: rgba(52,211,153,0.15); color: #34d399; padding: 2px 7px; border-radius: 99px; font-weight: 600; }
.badge-unlinked { font-size: 0.7rem; background: rgba(100,116,139,0.15); color: #64748b; padding: 2px 7px; border-radius: 99px; font-weight: 600; }
.badge-loading  { font-size: 0.7rem; background: rgba(249,115,22,0.1); color: #fb923c; padding: 2px 7px; border-radius: 99px; font-weight: 600; }

/* Toggle switch */
.switch {
  flex-shrink: 0;
  position: relative;
  width: 44px;
  height: 24px;
  border-radius: 99px;
  border: none;
  background: #334155;
  cursor: pointer;
  padding: 0;
  transition: background 0.2s;
  outline: none;
}
.switch:focus-visible { box-shadow: 0 0 0 2px #f97316; }
.switch.on { background: #f97316; }
.switch.loading { opacity: 0.6; cursor: wait; }
.switch:disabled { opacity: 0.35; cursor: not-allowed; }

.switch-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s cubic-bezier(.4,0,.2,1), box-shadow 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.4);
  pointer-events: none;
}
.switch.on .switch-thumb { transform: translateX(20px); }

.info-ids     { grid-column: 1 / -1; margin: -6px 0 0; font-size: 0.78rem; color: #34d399; }
.info-unlinked { grid-column: 1 / -1; margin: -6px 0 0; font-size: 0.78rem; color: #64748b; font-style: italic; }

.password-input {
  flex: 1;
  padding: 6px 10px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9;
  border-radius: 7px;
  font-size: 0.85rem;
  outline: none;
}
.password-input:focus { border-color: #f97316; }

input[type="checkbox"] { accent-color: #f97316; width: 16px; height: 16px; }

.btn-apply {
  background: #f97316;
  color: white;
  border: none;
  padding: 9px 0;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 700;
  font-size: 0.9rem;
  width: 100%;
  transition: opacity 0.15s;
}
.btn-apply:hover:not(:disabled) { opacity: 0.85; }
.btn-apply:disabled { background: #334155; color: #64748b; cursor: not-allowed; }
</style>
