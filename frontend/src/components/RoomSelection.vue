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

          <button v-else @click="handleRoomClick(room)" class="btn-join" :disabled="room.state !== 'LOBBY'">
            {{ room.state === 'LOBBY' ? 'Rejoindre' : 'En partie...' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { authStore } from '../authStore';

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
    const response = await fetch("http://localhost:8080/rooms");
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
    const response = await fetch("http://localhost:8080/rooms", {
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
/* Ajoute ici ton style pour les cartes, inputs et badges */
.room-selection-container { max-width: 800px; margin: 0 auto; padding: 20px; color: white; }
.rooms-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 15px; margin-top: 15px; }
.room-card { background: #2a2a2a; padding: 15px; border-radius: 8px; border-left: 5px solid #e91e63; display: flex; flex-direction: column; justify-content: space-between; }
.room-info { display: flex; flex-direction: column; gap: 5px; margin-bottom: 10px; }
.badge-private { color: #ff5722; font-weight: bold; }
.badge-public { color: #4caf50; font-weight: bold; }
.settings-hint { color: #aaa; font-size: 0.82rem; margin: -8px 0 12px; font-style: italic; }
.form-group { margin-bottom: 15px; display: flex; flex-direction: column; }
.form-group.checkbox { flex-direction: row; gap: 10px; align-items: center; }
input[type="text"], input[type="number"], input[type="password"] { padding: 8px; background: #333; border: 1px solid #555; color: white; border-radius: 4px; }
button { padding: 10px; border: none; border-radius: 4px; cursor: pointer; background: #e91e63; color: white; font-weight: bold; }
button:disabled { background: #555; cursor: not-allowed; }
.btn-refresh { background: #2196f3; margin-bottom: 15px; }
</style>