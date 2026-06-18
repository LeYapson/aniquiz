<template>
  <div class="room-selection-container">
    <!-- ── Vue d'accueil : un seul bouton Jouer ── -->
    <div v-if="!showMultiPanel" class="play-home">
      <button class="btn-play-main" @click="openModal">Jouer</button>
    </div>

    <!-- ── Vue Multi : liste des salons + création (flux existant) ── -->
    <div v-else class="multi-panel">
      <div class="multi-header">
        <h2>Salons disponibles</h2>
        <button class="btn-back" @click="closeMulti">← Retour</button>
      </div>

      <div class="room-actions">
        <button @click="isCreating = !isCreating" class="btn-toggle">
          {{ isCreating ? "← Voir les salons" : "Créer un salon" }}
        </button>
      </div>

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

        <p class="settings-hint">Les paramètres de jeu (rounds, durée, filtres) sont configurables dans le lobby.</p>

        <button @click="submitCreate" class="btn-submit" :disabled="!config.room_id">
          Créer et Rejoindre
        </button>
      </div>

      <div v-else class="rooms-list-container">
        <button @click="fetchRooms" class="btn-refresh">Actualiser la liste</button>

        <div v-if="rooms.length === 0" class="no-rooms">
          Aucun salon actif pour le moment. Créez-en un !
        </div>

        <div v-else class="rooms-grid">
          <div v-for="room in rooms" :key="room.id" class="room-card">
            <div class="room-info">
              <strong>{{ room.id }}</strong>
              <span>Joueurs : {{ room.players_count }}</span>
              <span>Configuration : {{ room.max_rounds }} rounds</span>
              <span v-if="room.is_private" class="badge-private">Privé</span>
              <span v-else class="badge-public">Public</span>
            </div>

            <div v-if="selectedRoomId === room.id && room.is_private" class="password-prompt">
              <input v-model="inputPassword" type="password" placeholder="Code d'accès" />
              <button @click="submitJoin(room)" class="btn-join-confirm">Valider</button>
            </div>

            <button v-else @click="handleRoomClick(room)" class="btn-join" :class="{ 'btn-spectate': room.state !== 'LOBBY' }">
              {{ room.state === 'LOBBY' ? 'Rejoindre' : 'Regarder' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Modale de sélection de mode ── -->
    <Teleport to="body">
      <div
        v-if="showModeModal"
        class="mode-modal-overlay"
        role="dialog"
        aria-modal="true"
        aria-label="Choix du mode de jeu"
        @click.self="showModeModal = false"
      >
        <div class="mode-modal">
          <button class="mode-modal-close" aria-label="Fermer" @click="showModeModal = false">✕</button>
          <h2 class="mode-modal-title">Choisis ton mode</h2>

          <div class="mode-cards">
            <button
              v-for="mode in modes"
              :key="mode.key"
              class="mode-card-btn"
              :class="{ 'mode-card-btn--soon': !mode.available }"
              :style="{ backgroundImage: `url(${mode.img}), ${mode.gradient}` }"
              :disabled="!mode.available"
              @click="selectMode(mode)"
            >
              <span class="mode-card-overlay" aria-hidden="true"></span>
              <span class="mode-card-content">
                <span class="mode-card-title">{{ mode.title }}</span>
                <span class="mode-card-desc">{{ mode.desc }}</span>
              </span>
              <span v-if="!mode.available" class="mode-card-soon">Bientôt</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { authStore } from '../authStore';
import { API_URL } from '../config';

const router = useRouter();

const emit = defineEmits(['room-created', 'room-joined']);

const rooms = ref([]);
const isCreating = ref(false);
const selectedRoomId = ref(null);
const inputPassword = ref("");

const showModeModal = ref(false);
const showMultiPanel = ref(false);

const config = ref({
  room_id: "",
  is_private: false,
  password: "",
});

// Modes de jeu. `img` : asset optionnel dans /public ; `gradient` sert de
// secours visuel si l'image est absente (background-image multi-couches).
const modes = [
  { key: 'solo',     title: 'Solo',       desc: "Entraîne-toi à ton rythme",      img: '/kora-solo.png',     gradient: 'linear-gradient(160deg, #f97316, #b45309)', available: true },
  { key: 'speedrun', title: 'Speed Run',  desc: '5 min · Enchaîne les animes',    img: '/kora-speedrun.png', gradient: 'linear-gradient(160deg, #ef4444, #7f1d1d)', available: true },
  { key: 'multi',    title: 'Multi',      desc: "Affronte d'autres joueurs",      img: '/kora-multi.png',    gradient: 'linear-gradient(160deg, #3b82f6, #1e3a8a)', available: true },
  { key: 'classe',   title: 'Classé',     desc: 'Grimpe les divisions',           img: '/kora-ranked.png',   gradient: 'linear-gradient(160deg, #a855f7, #581c87)', available: false },
];

const openModal = () => { showModeModal.value = true; };

const selectMode = (mode) => {
  if (!mode.available) return;
  showModeModal.value = false;
  if (mode.key === 'solo') {
    startSolo();
  } else if (mode.key === 'speedrun') {
    router.push('/speedrun');
  } else if (mode.key === 'multi') {
    showMultiPanel.value = true;
    fetchRooms();
  }
};

const closeMulti = () => {
  showMultiPanel.value = false;
  isCreating.value = false;
};

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

const startSolo = async () => {
  const soloId = `solo-${authStore.user?.username ?? 'player'}-${Date.now()}`;
  try {
    const res = await fetch(`${API_URL}/rooms`, {
      method: 'POST',
      headers: authStore.authHeaders(),
      body: JSON.stringify({ room_id: soloId, is_private: true, password: '' }),
    });
    if (res.ok) {
      const data = await res.json();
      emit('room-created', { room_id: data.room_id, creator_id: data.creator_id, password: '', isCreator: true });
    }
  } catch (err) {
    console.error('Erreur création solo:', err);
  }
};

// Fermeture de la modale via Échap
const onKeydown = (e) => {
  if (e.key === 'Escape') showModeModal.value = false;
};

onMounted(() => {
  fetchRooms();
  window.addEventListener('keydown', onKeydown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown);
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

/* ─── Bouton Jouer ─── */
.play-home {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}
.btn-play-main {
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: #fff;
  font-size: 1.3rem;
  font-weight: 800;
  padding: 18px 72px;
  border-radius: 50px;
  letter-spacing: 0.02em;
  box-shadow: 0 6px 28px rgba(249,115,22,0.45);
  transition: transform 0.15s, box-shadow 0.15s;
}
.btn-play-main:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 36px rgba(249,115,22,0.6);
  opacity: 1;
}

/* ─── En-tête vue Multi ─── */
.multi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.multi-header h2 { margin-bottom: 0; }
.btn-back {
  background: #1e2a45;
  color: #cbd5e1;
  border: 1px solid rgba(255,255,255,0.1);
}

/* ─── Modale de sélection de mode ─── */
.mode-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(8,8,20,0.8);
  backdrop-filter: blur(4px);
  animation: modeFade 0.18s ease;
}
.mode-modal {
  position: relative;
  width: 100%;
  max-width: 900px;
  background: #12122a;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 20px;
  padding: 28px;
  animation: modePop 0.2s ease;
}
.mode-modal-title {
  text-align: center;
  font-size: 1.5rem;
  font-weight: 800;
  color: #f1f5f9;
  margin: 0 0 22px;
}
.mode-modal-close {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 34px;
  height: 34px;
  padding: 0;
  border-radius: 50%;
  background: rgba(255,255,255,0.08);
  color: #cbd5e1;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mode-modal-close:hover { background: rgba(255,255,255,0.16); opacity: 1; }

.mode-cards {
  display: flex;
  gap: 16px;
  width: 100%;
}
.mode-card-btn {
  position: relative;
  flex: 1;
  height: 340px;
  padding: 0;
  border: none;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  display: flex;
  align-items: flex-end;
  background-color: #16213e;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  transition: transform 0.18s, box-shadow 0.18s;
}
.mode-card-btn:hover:not(:disabled) {
  transform: translateY(-6px);
  box-shadow: 0 14px 34px rgba(0,0,0,0.5);
  opacity: 1;
}
.mode-card-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, rgba(0,0,0,0.25) 45%, transparent 100%);
}
.mode-card-content {
  position: relative;
  z-index: 1;
  width: 100%;
  padding: 20px;
  text-align: left;
}
.mode-card-title {
  display: block;
  font-size: 1.4rem;
  font-weight: 800;
  color: #fff;
}
.mode-card-desc {
  display: block;
  margin-top: 4px;
  font-size: 0.85rem;
  font-weight: 500;
  color: rgba(255,255,255,0.82);
}
.mode-card-btn--soon {
  cursor: not-allowed;
  filter: grayscale(0.7) brightness(0.55);
}
.mode-card-soon {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 2;
  padding: 4px 10px;
  border-radius: 50px;
  background: rgba(15,15,35,0.85);
  border: 1px solid rgba(251,191,36,0.4);
  color: #fbbf24;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

@keyframes modeFade { from { opacity: 0; } to { opacity: 1; } }
@keyframes modePop {
  from { opacity: 0; transform: scale(0.96); }
  to   { opacity: 1; transform: scale(1); }
}

/* ─── Liste des salons ─── */
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
.room-actions { display: flex; gap: 10px; margin-bottom: 4px; flex-wrap: wrap; }
.btn-toggle { background: #1e2a45; color: #cbd5e1; border: 1px solid rgba(255,255,255,0.1); flex: 1; }
.btn-refresh { background: #3b82f6; margin-bottom: 14px; }
.btn-spectate { background: #6366f1; }
.btn-join-confirm { background: #f97316; }

.password-prompt { display: flex; gap: 8px; }
.password-prompt input { flex: 1; }

/* ─── Responsive ─── */
@media (max-width: 700px) {
  .mode-cards { flex-wrap: wrap; }
  .mode-card-btn { flex: 1 1 calc(50% - 8px); height: 200px; }
}
</style>
