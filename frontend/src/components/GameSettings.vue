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

    <button @click="apply" class="btn-apply">Appliquer</button>
  </div>
</template>

<script setup>
import { reactive, watch } from "vue";

const props = defineProps({
  socket: Object,
  initialSettings: Object,
});

const local = reactive({
  maxRounds: props.initialSettings?.maxRounds ?? 5,
  roundDuration: props.initialSettings?.roundDuration ?? 20,
  filterType: props.initialSettings?.filterType ?? "",
  decade: props.initialSettings?.decade ?? 0,
  isPrivate: props.initialSettings?.isPrivate ?? false,
  password: props.initialSettings?.password ?? "",
});

// Sync si les settings changent de l'extérieur (SETTINGS_UPDATED reçu)
watch(() => props.initialSettings, (s) => {
  if (!s) return;
  Object.assign(local, s);
}, { deep: true });

const apply = () => {
  if (!props.socket || props.socket.readyState !== WebSocket.OPEN) return;

  const minYear = local.decade > 0 ? local.decade : 0;
  const maxYear = local.decade > 0 ? local.decade + 9 : 0;

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
    },
  }));
};
</script>

<style scoped>
.settings-panel {
  background: #f8f8f8;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 16px;
  margin-top: 16px;
}
.settings-panel h3 { margin: 0 0 12px; font-size: 0.95rem; }
.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
}
.settings-grid label {
  display: flex;
  flex-direction: column;
  font-size: 0.82rem;
  font-weight: 600;
  color: #444;
  gap: 4px;
}
.settings-grid input[type="number"],
.settings-grid select {
  padding: 6px 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 0.9rem;
}
.full-width { grid-column: 1 / -1; flex-direction: row; align-items: center; gap: 8px; }
.password-input { flex: 1; padding: 4px 8px; border: 1px solid #ccc; border-radius: 4px; }
.btn-apply {
  background: #1e90ff;
  color: white;
  border: none;
  padding: 8px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  width: 100%;
}
</style>
