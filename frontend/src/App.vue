<template>
  <div id="app">
    <header v-if="authStore.user" class="app-header">
      <span>Bienvenue, <strong>{{ authStore.user.username }}</strong> (Niv. {{ authStore.user.level }})</span>
      <div class="header-actions">
        <span v-if="authStore.user.anilist_username" class="anilist-badge">
          AniList : <strong>{{ authStore.user.anilist_username }}</strong> ✓
        </span>
        <button v-else @click="connectAnilist" class="btn-anilist">Connecter AniList</button>

        <span v-if="authStore.user.mal_username" class="mal-badge">
          MAL : <strong>{{ authStore.user.mal_username }}</strong> ✓
        </span>
        <button v-else @click="connectMAL" class="btn-mal">Connecter MAL</button>

        <button @click="showProfile = !showProfile" class="btn-profile">
          {{ showProfile ? '🎮 Jeu' : '👤 Profil' }}
        </button>
        <button @click="authStore.logout" class="btn-logout">Déconnexion</button>
      </div>
    </header>

    <Transition name="toast">
      <div v-if="xpToast" class="xp-toast">
        <span class="xp-icon">⭐</span>
        <div>
          <strong>+{{ xpToast.xpGained }} XP</strong>
          <div v-if="xpToast.levelUp" class="level-up">Niveau {{ xpToast.newLevel }} atteint ! 🎉</div>
          <div v-else class="xp-total">Total : {{ xpToast.newXP }} XP · Niv. {{ xpToast.newLevel }}</div>
        </div>
      </div>
    </Transition>

    <main>
      <AuthForm v-if="!authStore.user" />

      <template v-else>
        <ProfilePage v-if="showProfile" />

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
            <div v-if="reconnectMsg" class="reconnect-banner">
              🔄 {{ reconnectMsg }}
            </div>
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
                <GameSettings
                  v-if="isCreator"
                  :socket="socket"
                  :initialSettings="roomSettings"
                />
              </div>

              <div v-if="state === 'GAME_OVER'" class="game-over-screen">
                <h2>🏆 Partie terminée !</h2>
                <ol class="final-scores">
                  <li
                    v-for="(p, i) in finalScores"
                    :key="p.id"
                    :class="{ 'me': p.username === authStore.user.username, 'gold': i === 0 }"
                  >
                    <span class="rank">{{ i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `#${i + 1}` }}</span>
                    <span class="pname">{{ p.username }}</span>
                    <span class="pts">{{ p.score }} pts</span>
                  </li>
                </ol>
                <button @click="backToLobby" class="btn-start" style="margin-top:20px">
                  Retour au lobby
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
      </template>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import RoomSelection from "./components/RoomSelection.vue";
import GameTimer from "./components/GameTimer.vue";
import AuthForm from "./components/AuthForm.vue";
import ProfilePage from "./components/ProfilePage.vue";
import GameSettings from "./components/GameSettings.vue";
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
const xpToast = ref(null);
const finalScores = ref([]);
const reconnectMsg = ref("");
const showProfile = ref(false);
const isCreator = ref(false);
const roomSettings = ref({ maxRounds: 5, roundDuration: 20, filterType: "", decade: 0, isPrivate: false, password: "" });
let socket = null;
let reconnectAttempts = 0;
let intentionalClose = false;

const startGame = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "START_GAME", payload: null }));
  }
};

const backToLobby = () => {
  finalScores.value = [];
  state.value = "LOBBY";
};

const submitAnswer = () => {
  if (!userGuess.value) return;
  socket.send(JSON.stringify({ type: "SUBMIT_ANSWER", payload: userGuess.value }));
  userGuess.value = "";
};

const connectAnilist = () => {
  window.location.href = `http://localhost:8080/api/auth/anilist?token=${authStore.token}`
}

const connectMAL = () => {
  window.location.href = `http://localhost:8080/api/auth/mal?token=${authStore.token}`
}

// Gestion des retours OAuth (?anilist=success&username=xxx ou ?mal=success&username=xxx)
const checkOAuthCallback = () => {
  const params = new URLSearchParams(window.location.search)

  const anilistStatus = params.get('anilist')
  if (anilistStatus === 'success') {
    const username = params.get('username')
    if (username && authStore.user) {
      authStore.setUser({ ...authStore.user, anilist_username: username }, authStore.token)
    }
  }

  const malStatus = params.get('mal')
  if (malStatus === 'success') {
    const username = params.get('username')
    if (username && authStore.user) {
      authStore.setUser({ ...authStore.user, mal_username: username }, authStore.token)
    }
  }

  if (anilistStatus || malStatus) {
    window.history.replaceState({}, '', window.location.pathname)
  }
}

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
  checkOAuthCallback();
});

const setupWebSocket = ({ room_id, password, isCreator: creator }) => {
  room.value = room_id;
  isCreator.value = !!creator;
  intentionalClose = false;
  reconnectAttempts = 0;
  connectWebSocket(room_id, password);
};

const connectWebSocket = (room_id, password) => {
  const wsUrl = `ws://localhost:8080/ws?room=${room_id}&password=${password || ""}&token=${authStore.token}`;
  socket = new WebSocket(wsUrl);

  socket.onopen = () => {
    isConnected.value = true;
    reconnectAttempts = 0;
    reconnectMsg.value = "";
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
        case "SETTINGS_UPDATED":
          roomSettings.value = {
            maxRounds: data.payload.max_rounds,
            roundDuration: data.payload.round_duration,
            filterType: data.payload.filter_type,
            isPrivate: data.payload.is_private,
          };
          break;
        case "GAME_OVER":
          finalScores.value = [...players.value].sort((a, b) => b.score - a.score);
          state.value = "GAME_OVER";
          break;
        case "XP_GAINED":
          const oldLevel = authStore.user?.level ?? 1;
          const levelUp = data.payload.new_level > oldLevel;
          xpToast.value = {
            xpGained: data.payload.xp_gained,
            newXP: data.payload.new_xp,
            newLevel: data.payload.new_level,
            levelUp,
          };
          // Mettre à jour le niveau dans le store
          if (authStore.user) {
            authStore.setUser(
              { ...authStore.user, xp: data.payload.new_xp, level: data.payload.new_level },
              authStore.token
            );
          }
          setTimeout(() => { xpToast.value = null; }, 5000);
          break;
      }
    } catch (err) {
      console.error("Erreur message:", err);
    }
  };

  socket.onclose = () => {
    if (intentionalClose) {
      isConnected.value = false;
      players.value = [];
      state.value = "LOBBY";
      isRevealing.value = false;
      reconnectMsg.value = "";
      return;
    }
    // Déconnexion involontaire : backoff exponentiel (1s, 2s, 4s… max 30s)
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
    reconnectAttempts++;
    reconnectMsg.value = `Connexion perdue. Reconnexion dans ${Math.round(delay / 1000)}s… (tentative ${reconnectAttempts})`;
    setTimeout(() => connectWebSocket(room.value, ""), delay);
  };
};

const disconnect = () => {
  intentionalClose = true;
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
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.btn-profile {
  background: #444;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}
.btn-logout {
  background: #ff4757;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
}
.btn-anilist {
  background: #02a9ff;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}
.anilist-badge {
  color: #02a9ff;
  font-size: 0.9rem;
}
.btn-mal {
  background: #2e51a2;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}
.mal-badge {
  color: #7db4de;
  font-size: 0.9rem;
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
.xp-toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  background: #1a1a2e;
  color: #fff;
  padding: 14px 20px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.4);
  border-left: 4px solid #ffd700;
  z-index: 1000;
}
.xp-icon { font-size: 1.8rem; }
.level-up { color: #ffd700; font-weight: bold; margin-top: 2px; }
.xp-total { color: #aaa; font-size: 0.85rem; margin-top: 2px; }
.toast-enter-active, .toast-leave-active { transition: all 0.4s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(20px); }
.game-over-screen { text-align: center; padding: 20px; }
.game-over-screen h2 { font-size: 1.8rem; margin-bottom: 20px; }
.final-scores { list-style: none; padding: 0; max-width: 400px; margin: 0 auto; }
.final-scores li {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 16px; margin-bottom: 8px;
  background: #f4f4f4; border-radius: 8px;
  font-size: 1rem;
}
.final-scores li.gold { background: #fff8e1; border-left: 4px solid #ffd700; }
.final-scores li.me { font-weight: bold; }
.final-scores .rank { font-size: 1.3rem; width: 32px; }
.final-scores .pname { flex: 1; }
.final-scores .pts { color: #666; font-size: 0.9rem; }
.reconnect-banner {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffc107;
  border-radius: 6px;
  padding: 8px 14px;
  margin-bottom: 12px;
  font-size: 0.9rem;
}
</style>
