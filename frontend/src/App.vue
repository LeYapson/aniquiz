<template>
  <div id="app">
    <header v-if="authStore.user" class="app-header">
      <span>Bienvenue, <strong>{{ authStore.user.username }}</strong> (Niv. {{ authStore.user.level }})</span>
      <button @click="authStore.logout" class="btn-logout">Déconnexion</button>
    </header>

    <main>
      <AuthForm v-if="!authStore.user" />

      <div v-else>
        <h1>AniQuiz 🎵</h1>

        <div v-if="!isConnected">
          <div style="max-width: 800px; margin: 0 auto; padding: 0 20px;">
            <p>Joueur : <strong>{{ authStore.user.username }}</strong></p>
          </div>
          <RoomSelection @room-created="setupWebSocket" @room-joined="setupWebSocket" />
        </div>

        <div v-else class="game-layout">
          <aside class="sidebar">
            <h3>Joueurs ({{ players.length }})</h3>
            <ul>
              <li v-for="p in players" :key="p.id">
                {{ p.username }} <span v-if="p.username === authStore.user.username">⭐</span>
                <small>({{ p.score }} pts)</small>
              </li>
            </ul>
          </aside>

          <main class="game-area">
            <div class="status-bar">
              <p>
                Salon : <strong>{{ room }}</strong> | Joueur :
                <strong>{{ authStore.user.username }}</strong>
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
                      list="anime-suggestions"
                    />

                    <datalist id="anime-suggestions">
                      <option v-for="anime in animeDictionary" :key="anime" :value="anime"></option>
                    </datalist>

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
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import RoomSelection from "./components/RoomSelection.vue";
import GameTimer from "./components/GameTimer.vue";
import AuthForm from "./components/AuthForm.vue";
import { authStore } from "./authStore";

const isConnected = ref(false);
const room = ref("");
const players = ref([]);
const state = ref("LOBBY");
const currentAudioUrl = ref("");
const userGuess = ref("");
const roundDuration = ref(0);
const animeDictionary = ref([]);
const isRevealing = ref(false);
const currentAnswerInfo = ref({
  animeName: "",
  title: "",
  artist: "",
  videoUrl: "",
});
let socket = null;

const startGame = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "START_GAME", payload: null }));
  }
};

const submitAnswer = () => {
  if (!userGuess.value) return;
  socket.send(JSON.stringify({ type: "SUBMIT_ANSWER", payload: userGuess.value }));
  userGuess.value = "";
};

const loadAnimeDictionary = async () => {
  try {
    const response = await fetch("http://localhost:8080/animes");
    if (response.ok) {
      animeDictionary.value = await response.json();
    }
  } catch (err) {
    console.error("Erreur lors du chargement du dictionnaire :", err);
  }
};

const authFetch = (url, options = {}) => {
  return fetch(url, {
    ...options,
    headers: {
      ...authStore.authHeaders(),
      ...(options.headers || {}),
    },
  });
};

onMounted(() => {
  loadAnimeDictionary();
});

const setupWebSocket = ({ room_id, password }) => {
  room.value = room_id;
  const wsUrl = `ws://localhost:8080/ws?room=${room_id}&password=${password || ""}&token=${authStore.token}`;
  socket = new WebSocket(wsUrl);

  socket.onopen = () => {
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
          isRevealing.value = false;
          currentAudioUrl.value = data.payload.audio_url;
          roundDuration.value = data.payload.duration;
          const audio = document.querySelector("audio");
          if (audio) {
            audio.load();
            audio.play().catch((e) => console.warn("Autoplay bloqué par le navigateur"));
          }
          break;
        case "ROUND_ENDED":
          isRevealing.value = true;
          currentAnswerInfo.value = {
            animeName: data.payload.answer,
            title: "Générique",
            artist: "Inconnu",
            videoUrl: currentAudioUrl.value,
          };
          currentAudioUrl.value = "";
          break;
        case "GAME_OVER":
          alert("Partie terminée ! " + (data.payload?.message ?? ""));
          state.value = "LOBBY";
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
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #1a1a2e;
  color: #fff;
}
.btn-logout {
  background: #ff4757;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}
main {
  flex: 1;
}
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
.quiz-box { background: #fafafa; padding: 20px; border-radius: 8px; border: 1px dashed #ccc; }
.btn-start { background: #2ed573; color: white; border: none; padding: 10px 15px; border-radius: 4px; font-weight: bold; cursor: pointer; width: 100%; }
.answer-zone { margin: 20px 0; display: flex; gap: 10px; }
.answer-zone input { flex: 1; padding: 10px; border: 1px solid #ccc; border-radius: 4px; }
.answer-zone button { background: #1e90ff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
</style>
