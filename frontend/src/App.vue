<template>
  <div class="app-container">
    <h1>AniQuiz 🎵</h1>

    <!-- ÉTAPE 1 : Si pas connecté, on montre le formulaire -->
    <div v-if="!isConnected">
      <JoinRoom @join="setupWebSocket" />
    </div>

    <!-- ÉTAPE 2 : SINON (on est connecté), on montre l'interface de jeu -->
    <div v-else class="game-layout">
      <!-- Barre latérale avec les joueurs -->
      <aside class="sidebar">
        <h3>Joueurs ({{ players.length }})</h3>
        <ul>
          <li v-for="p in players" :key="p.id">
            {{ p.username }} <span v-if="p.username === user">⭐</span>
            <small>({{ p.score }} pts)</small>
          </li>
        </ul>
      </aside>

      <!-- Zone principale -->
      <main class="game-area">
        <div class="status-bar">
          <p>
            Salon : <strong>{{ room }}</strong> | Joueur :
            <strong>{{ user }}</strong>
          </p>
          <button @click="disconnect" class="btn-quit">Quitter</button>
        </div>

        <div class="quiz-box">
          <div v-if="state === 'LOBBY'">
            <p v-if="players.length >= 1">Prêt à jouer ?</p>
            <button @click="startGame" class="btn-start">
              Lancer la partie
            </button>
          </div>

          <div v-if="state === 'PLAYING'">
            <!-- On affichera le lecteur ici après -->
            <p>🎵 Écoutez attentivement...</p>

            <!-- Le lecteur audio (en autoplay pour le quiz) -->
            <audio v-if="currentAudioUrl" :src="currentAudioUrl" autoplay controls></audio>

            <div class="answer-zone">
              <input v-model="userGuess" @keyup.enter="submitAnswer" placeholder="Nom de l'anime..." />
              <button @click="submitAnswer">Envoyer</button>
            </div>
            <div class="leaderboard">
              <h3>Classement</h3>
              <ul>
                <li v-for="p in players" :key="p.id">
                  {{ p.username }}: {{ p.score }} pts
                </li>
              </ul>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import JoinRoom from "./components/JoinRoom.vue";

// 1. Variables d'état GLOBALES au composant
const isConnected = ref(false);
const user = ref("");
const room = ref("");
const players = ref([]);
const state = ref("LOBBY"); // Sorti de la fonction
const currentAudioUrl = ref("");
let socket = null;
const userGuess = ref("");

// 2. Fonctions d'action
const startGame = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(
      JSON.stringify({
        type: "START_GAME",
        payload: null,
      }),
    );
  }
};

const submitAnswer = () => {
  if (!userGuess.value) return;

  socket.send(
    JSON.stringify({
      type: "SUBMIT_ANSWER",
      payload: userGuess.value,
    }),
  );
  userGuess.value = ""; // Clear input après envoi
};
const setupWebSocket = ({ username, roomId }) => {
  user.value = username;
  room.value = roomId;

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${protocol}//${window.location.host}/ws?username=${username}&room=${roomId}`;

  socket = new WebSocket(url);

  socket.onopen = () => {
    console.log("✅ Connecté au serveur Go !");
    isConnected.value = true;
  };

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case "PLAYER_LIST":
          players.value = data.payload;
          break;
        case "GAME_STATE":
          state.value = data.payload;
          break;
        case "NewQuestion":
          console.log("Nouvelle question reçue :", data.payload);
          currentAudioUrl.value = data.payload.audio_url;
          // Petit hack pour forcer le chargement si besoin
          const audio = document.querySelector("audio");
          if (audio) {
            audio.load();
            audio
              .play()
              .catch((e) => console.warn("Autoplay bloqué par le navigateur"));
          }
          break;
        case "GAME_OVER":
          alert("Partie terminée ! " + data.payload.message);
          // Ici tu pourrais stocker le gagnant dans une variable ref pour l'afficher proprement
          break;
      }
    } catch (err) {
      console.error("Erreur message:", err);
    }
  };

  socket.onclose = () => {
    isConnected.value = false;
    players.value = [];
    state.value = "LOBBY"; // On reset l'état à la déco
  };
};

const disconnect = () => {
  if (socket) socket.close();
};

defineExpose({ state, isConnected });
</script>

<style>
.game-layout {
  display: flex;
  gap: 20px;
  max-width: 1000px;
  margin: 20px auto;
  text-align: left;
}

.sidebar {
  width: 200px;
  background: #f4f4f4;
  padding: 15px;
  border-radius: 8px;
}

.game-area {
  flex: 1;
  background: #fff;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #eee;
  padding-bottom: 10px;
  margin-bottom: 20px;
}

.btn-quit {
  background: #ff4757;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}
</style>
