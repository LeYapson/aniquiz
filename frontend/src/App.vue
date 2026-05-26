<template>
  <div class="app-container">
    <h1>AniQuiz 🎵</h1>

    <div v-if="!user" class="pseudo-setup">
      <div class="quiz-box" style="max-width: 400px; margin: 40px auto; text-align: center;">
        <h3>Choisissez votre Pseudo pour commencer :</h3>
        <input v-model="pseudoInput" @keyup.enter="savePseudo" type="text" placeholder="Ex: Gon, Goku..." style="padding: 10px; width: 80%; margin-bottom: 15px;" />
        <button @click="savePseudo" :disabled="!pseudoInput" class="btn-start" style="width: auto; padding: 10px 20px;">Continuer</button>
      </div>
    </div>

    <div v-else-if="!isConnected">
      <div style="max-width: 800px; margin: 0 auto; display: flex; justify-content: space-between; align-items: center; padding: 0 20px;">
        <p>Joueur : <strong>{{ user }}</strong></p>
        <button @click="user = ''" class="btn-quit" style="background: #777;">Changer de pseudo</button>
      </div>
      <RoomSelection @room-created="setupWebSocket" @room-joined="setupWebSocket" />
    </div>

    <div v-else class="game-layout">
      <aside class="sidebar">
        <h3>Joueurs ({{ players.length }})</h3>
        <ul>
          <li v-for="p in players" :key="p.id">
            {{ p.username }} <span v-if="p.username === user">⭐</span>
            <small>({{ p.score }} pts)</small>
          </li>
        </ul>
      </aside>

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
            <div v-if="isRevealing" class="reveal-zone">
              <h2>🎉 Réponse : <span style="color: #e91e63;">{{ currentAnswerInfo.animeName }}</span></h2>
              <p><strong>{{ currentAnswerInfo.title }}</strong> - {{ currentAnswerInfo.artist }}</p>
              
              <video
                v-if="currentAnswerInfo.videoUrl"
                :src="currentAnswerInfo.videoUrl"
                autoplay
                controls
                style="width: 100%; max-width: 600px; border-radius: 8px; margin-top: 10px;"
              ></video>
              <p v-else style="font-style: italic; margin-top: 20px;">Pas de vidéo disponible pour cette piste.</p>
            </div>

            <div v-else>
              <p>🎵 Écoutez attentivement...</p>
              
              <GameTimer :duration="roundDuration" :key="currentAudioUrl" />

              <audio
                v-if="currentAudioUrl"
                :src="currentAudioUrl"
                autoplay
              ></audio>

              <div class="answer-zone">
                <input
                  v-model="userGuess"
                  @keyup.enter="submitAnswer"
                  placeholder="Nom de l'anime..."
                />
                <button @click="submitAnswer">Envoyer</button>
              </div>
            </div>

            <div class="leaderboard" style="margin-top: 20px;">
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
import RoomSelection from "./components/RoomSelection.vue"; // Remplacement de JoinRoom
import GameTimer from "./components/GameTimer.vue";

// 1. Variables d'état GLOBALES
const isConnected = ref(false);
const pseudoInput = ref("");
const user = ref("");
const room = ref("");
const players = ref([]);
const state = ref("LOBBY");
const currentAudioUrl = ref("");
const userGuess = ref("");
const roundDuration = ref(0); // Stockera la durée dynamique reçue du Back
let socket = null;

// Validation du pseudo initial
const savePseudo = () => {
  if (pseudoInput.value.trim()) {
    user.value = pseudoInput.value.trim();
  }
};

const isRevealing = ref(false);
const currentAnswerInfo = ref({
  animeName: "",
  title: "",
  artist: "",
  videoUrl: ""
});

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
  userGuess.value = "";
};

// Modifié pour accepter l'objet de données provenant de RoomSelection
const setupWebSocket = ({ room_id, password }) => {
  room.value = room_id;

  // Ciblage explicite du port 8080 en local pour éviter les décalages de ports entre front et back
  const wsUrl = `ws://localhost:8080/ws?username=${user.value}&room=${room_id}&password=${password || ''}`;

  socket = new WebSocket(wsUrl);

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
          isRevealing.value = false; 
          
          currentAudioUrl.value = data.payload.audio_url;
          roundDuration.value = data.payload.duration;
          
          const audio = document.querySelector("audio");
          if (audio) {
            audio.load();
            audio
              .play()
              .catch((e) => console.warn("Autoplay bloqué par le navigateur"));
          }
          break;

        case "ROUND_ENDED":
          console.log("Fin du round, affichage de la réponse :", data.payload);
          isRevealing.value = true;
          
          // --- ADAPTATION DES COMPOSANTS EN FONCTION DES LOGS REÇUS ---
          currentAnswerInfo.value = {
            animeName: data.payload.answer, // Le back envoie "answer" et non "anime_name"
            title: "Générique",             // Valeur par défaut car non fournie par le back
            artist: "Inconnu",              // Valeur par défaut car non fournie par le back
            
            // On réutilise l'URL de l'audio reçue au début du round puisque c'est un .webm (vidéo)
            videoUrl: currentAudioUrl.value 
          };
          
          currentAudioUrl.value = ""; // Coupe l'audio caché pour laisser la vidéo jouer avec le son
          break;
        case "GAME_OVER":
          alert("Partie terminée !");
          state.value = "LOBBY"; // Protection locale
          break;
      }
    } catch (err) {
      console.error("Erreur message:", err);
    }
  };

  socket.onclose = () => {
    isConnected.value = false;
    players.value = [];
    state.value = "LOBBY";

    isRevealing.value = false;
  };
};

const disconnect = () => {
  if (socket) socket.close();
};

defineExpose({ state, isConnected });
</script>

<style>
/* Tes styles d'origine conservés */
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
/* Styles rapides pour l'intégration du bouton de base d'origine */
.quiz-box { background: #fafafa; padding: 20px; border-radius: 8px; border: 1px dashed #ccc; }
.btn-start { background: #2ed573; color: white; border: none; padding: 10px 15px; border-radius: 4px; font-weight: bold; cursor: pointer; width: 100%; }
.answer-zone { margin: 20px 0; display: flex; gap: 10px; }
.answer-zone input { flex: 1; padding: 10px; border: 1px solid #ccc; border-radius: 4px; }
.answer-zone button { background: #1e90ff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
</style>