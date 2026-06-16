<template>
  <div class="room-selection-container">
    <h2>🎯 Salons Disponibles</h2>

    <button @click="isCreating = !isCreating" class="btn-toggle">
      {{ isCreating ? "Voir les salons" : "🛠️ Créer un salon personnalisé" }}
    </button>

    <hr />

    <div v-if="isCreating" class="form-container">
      <h3>Créer un salon</h3>
      <div class="form-group">
        <label>Nom du salon :</label>
        <input v-model="config.room_id" type="text" placeholder="Ex: Salon de Théau" />
      </div>

      <div class="form-group checkbox">
        <input v-model="config.is_private" type="checkbox" id="private" />
        <label for="private">Salon privé (avec mot de passe)</label>
      </div>

      <div v-if="config.is_private" class="form-group">
        <label>Mot de passe :</label>
        <input v-model="config.password" type="password" placeholder="Mot de passe requis" />
      </div>

      <p class="settings-hint">⚙️ Les paramètres de jeu (rounds, durée, filtres) sont configurables dans le lobby.</p>

      <button @click="submitCreate" class="btn-submit" :disabled="!config.room_id">
        Créer et Rejoindre
      </button>
    </div>

    <div v-else class="rooms-list-container">
      <button @click="fetchRooms" class="btn-refresh">🔄 Actualiser la liste</button>

      <div v-if="rooms.length === 0" class="no-rooms">
        Aucun salon actif pour le moment. Créez-en un !
      </div>

      <div v-else class="rooms-grid">
        <div v-for="room in rooms" :key="room.id" class="room-card">
          <div class="room-info">
            <strong>{{ room.id }}</strong>
            <span>👥 Joueurs : {{ room.players_count }}</span>
            <span>⏱️ Configuration : {{ room.max_rounds }} rounds</span>
            <span v-if="room.is_private" class="badge-private">🔒 Privé</span>
            <span v-else class="badge-public">🌐 Public</span>
          </div>

          <div v-if="selectedRoomId === room.id && room.is_private" class="password-prompt">
            <input v-model="inputPassword" type="password" placeholder="Code d'accès" />
            <button @click="submitJoin(room)" class="btn-join-confirm">Valider</button>
          </div>

          <button v-else @click="handleRoomClick(room)" class="btn-join" :class="{ 'btn-spectate': room.state !== 'LOBBY' }">
            {{ room.state === 'LOBBY' ? 'Rejoindre' : '👁 Regarder' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { authStore } from '../authStore';
import { API_URL } from '../config';

const emit = defineEmits(['room-created', 'room-joined']);

const rooms = ref([]);
const isCreating = ref(false);
const selectedRoomId = ref(null);
const inputPassword = ref("");

const config = ref({
  room_id: "",
  is_private: false,
  password: "",
});

// Récupérer la liste des salons depuis l'API Back
const fetchRooms = async () => {
  try {
    const response = await fetch(`${API_URL}/rooms`);
    if (response.ok) {
      rooms.value = await response.json() || [];
    }
  } catch (error) {
    console.error("Erreur récupération salons:", error);
  }
};

// Demander la création au Back
const submitCreate = async () => {
  try {
    const response = await fetch(`${API_URL}/rooms`, {
      method: "POST",
      headers: authStore.authHeaders(),
      body: JSON.stringify(config.value)
    });

    if (response.ok) {
      const data = await response.json();
      // On émet l'événement vers App.vue avec les infos nécessaires
      emit('room-created', {
        room_id: data.room_id,
        creator_id: data.creator_id,
        password: config.value.password,
        isCreator: true,
      });
    } else {
      const err = await response.json();
      alert("Erreur: " + err.error);
    }
  } catch (error) {
    console.error("Erreur création salon:", error);
  }
};

const handleRoomClick = (room) => {
  if (room.is_private) {
    selectedRoomId.value = room.id;
    inputPassword.value = "";
  } else {
    submitJoin(room);
  }
};

const submitJoin = (room) => {
  emit('room-joined', {
    room_id: room.id,
    password: room.is_private ? inputPassword.value : ""
  });
};

onMounted(() => {
  fetchRooms();
});
</script>

<style scoped>
.room-selection-container {
  max-width: 860px;
  margin: 0 auto;
  padding: 40px 24px;
  color: #e2e8f0;
}
.room-selection-container h2 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 20px;
  color: #f1f5f9;
}

hr { border: none; border-top: 1px solid rgba(255,255,255,0.07); margin: 16px 0; }

.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
  margin-top: 16px;
}
.room-card {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-left: 4px solid #f97316;
  padding: 16px;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  transition: border-color 0.15s;
}
.room-card:hover { border-left-color: #ea580c; }
.room-info { display: flex; flex-direction: column; gap: 5px; font-size: 0.88rem; color: #94a3b8; }
.room-info strong { color: #f1f5f9; font-size: 1rem; }
.badge-private { color: #fb923c; font-weight: 600; }
.badge-public  { color: #34d399; font-weight: 600; }
.no-rooms { color: #64748b; text-align: center; padding: 40px 0; font-style: italic; }

.form-container {
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 12px;
  padding: 24px;
  max-width: 480px;
}
.form-container h3 { color: #f1f5f9; margin-bottom: 18px; }
.form-group { margin-bottom: 14px; display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.85rem; color: #94a3b8; }
.form-group.checkbox { flex-direction: row; align-items: center; gap: 10px; }
.settings-hint { color: #64748b; font-size: 0.8rem; font-style: italic; margin: -4px 0 12px; }

input[type="text"],
input[type="number"],
input[type="password"] {
  padding: 9px 12px;
  background: #0f0f23;
  border: 1px solid rgba(255,255,255,0.1);
  color: #f1f5f9;
  border-radius: 7px;
  outline: none;
  transition: border-color 0.15s;
}
input[type="text"]:focus,
input[type="number"]:focus,
input[type="password"]:focus { border-color: #f97316; }
input[type="text"]::placeholder,
input[type="password"]::placeholder { color: #475569; }

button {
  padding: 9px 16px;
  border: none;
  border-radius: 7px;
  cursor: pointer;
  background: #f97316;
  color: white;
  font-weight: 700;
  font-size: 0.875rem;
  transition: opacity 0.15s;
}
button:hover { opacity: 0.85; }
button:disabled { background: #334155; color: #64748b; cursor: not-allowed; opacity: 1; }
.btn-toggle { background: #1e2a45; color: #cbd5e1; border: 1px solid rgba(255,255,255,0.1); margin-bottom: 4px; }
.btn-refresh { background: #3b82f6; margin-bottom: 14px; }
.btn-spectate { background: #6366f1; }
.btn-join-confirm { background: #f97316; }

.password-prompt { display: flex; gap: 8px; }
.password-prompt input { flex: 1; }
</style>