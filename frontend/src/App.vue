<template>
  <div id="app">
    <AppHeader
      v-if="authStore.user"
      :currentView="currentView"
      @navigate="navigateTo"
      @logout="authStore.logout"
      @connect-anilist="connectAnilist"
      @connect-mal="connectMAL"
    />

    <ToastContainer />

    <main>
      <LandingPage
        v-if="!authStore.user && showLanding"
        @play="showLanding = false"
        @leaderboard="showLanding = false; showPublicLeaderboard = true"
      />
      <div v-else-if="!authStore.user && showPublicLeaderboard" class="public-leaderboard">
        <button class="btn-back-landing" @click="showLanding = true; showPublicLeaderboard = false">
          ← Retour
        </button>
        <LeaderboardPage :ownUsername="''" />
      </div>
      <AuthForm v-else-if="!authStore.user" />

      <template v-else>
        <LeaderboardPage v-if="!isConnected && currentView === 'leaderboard'" :ownUsername="authStore.user?.username" />

        <ProfilePage v-else-if="!isConnected && currentView === 'profile'" />

        <NewsPage v-else-if="!isConnected && currentView === 'news'" />

        <div v-else class="app-main">

        <div v-if="!isConnected" class="lobby-wrapper">
          <HomeDashboard @room-created="setupWebSocket" @room-joined="setupWebSocket" />
        </div>

        <div v-else class="game-layout" :data-mobile-tab="mobileTab">
          <aside class="sidebar" aria-label="Liste des joueurs">
            <div v-if="isSpectator" class="spectator-badge" role="status">👁 Mode spectateur</div>
            <h3 id="players-heading">Joueurs ({{ players.length }})</h3>
            <ul aria-labelledby="players-heading">
              <li
                v-for="p in players"
                :key="p.id"
                :aria-current="p.username === authStore.user.username ? 'true' : undefined"
              >
                {{ p.username }}
                <span v-if="p.username === authStore.user.username" aria-label="Vous">⭐</span>
                <small aria-label="`${p.score} points`">({{ p.score }} pts)</small>
                <button
                  v-if="isCreator && p.username !== authStore.user.username"
                  @click="kickPlayer(p.username)"
                  class="btn-kick"
                  :aria-label="`Expulser ${p.username}`"
                  title="Expulser"
                >✕</button>
              </li>
            </ul>
            <div v-if="spectatorCount > 0" class="spectator-count" aria-live="polite">
              👁 {{ spectatorCount }} spectateur{{ spectatorCount > 1 ? 's' : '' }}
            </div>
          </aside>

          <div class="game-area">
            <div v-if="reconnectMsg" class="reconnect-banner" role="alert">
              🔄 {{ reconnectMsg }}
            </div>
            <div class="status-bar">
              <p>
                Salon&nbsp;: <strong>{{ room }}</strong> &mdash; Joueur&nbsp;:
                <strong>{{ authStore.user.username }}</strong>
              </p>
              <button @click="disconnect" class="btn-quit" aria-label="Quitter le salon">Quitter</button>
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
                <ol class="final-scores" aria-label="Classement final">
                  <li
                    v-for="(p, i) in finalScores"
                    :key="p.id"
                    :class="{ 'me': p.username === authStore.user.username, 'gold': i === 0 }"
                  >
                    <span class="rank" aria-hidden="true">{{ i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `#${i + 1}` }}</span>
                    <span class="pname">{{ p.username }}</span>
                    <span class="pts">{{ p.score }} pts</span>
                  </li>
                </ol>
                <!-- Résumé round par round -->
                <div v-if="roundHistory.length > 0" class="round-history">
                  <h3>Récapitulatif</h3>
                  <div v-for="r in roundHistory" :key="r.round" class="round-item">
                    <div class="round-item-header">
                      <span class="round-num">Round {{ r.round }}</span>
                      <span v-if="r.track_type" class="round-tag">{{ r.track_type }}</span>
                      <span class="round-anime">{{ r.anime_name }}</span>
                    </div>
                    <div class="round-item-track">{{ r.title }}<span v-if="r.artist"> — {{ r.artist }}</span></div>
                    <div v-if="r.found_by && r.found_by.length > 0" class="round-finders">
                      <span v-for="(f, i) in r.found_by" :key="f.username" class="round-finder">
                        {{ i === 0 ? '🥇' : `#${i + 1}` }} {{ f.username }}
                        <em>{{ (f.time_ms / 1000).toFixed(1) }}s</em>
                        <span v-if="f.bonus > 0" class="round-bonus">+{{ f.bonus }}</span>
                      </span>
                    </div>
                    <div v-else class="round-nobody">😅 Personne n'a trouvé</div>
                  </div>
                </div>

                <button @click="backToLobby" class="btn-start" style="margin-top:20px">
                  Retour au lobby
                </button>
              </div>

              <div v-if="state === 'PLAYING'">
                <div v-if="isRevealing" class="reveal-zone" aria-live="assertive">
                  <div class="reveal-header">
                    <span v-if="currentAnswerInfo.trackType" class="reveal-tag">{{ currentAnswerInfo.trackType }}</span>
                    <span v-if="currentAnswerInfo.difficulty > 0" class="reveal-difficulty" :data-level="difficultyLabel(currentAnswerInfo.difficulty)">
                      {{ difficultyLabel(currentAnswerInfo.difficulty) }}
                    </span>
                  </div>
                  <h2>🎉 <span class="answer-name">{{ currentAnswerInfo.animeName }}</span></h2>
                  <p class="reveal-track-info">
                    <strong>{{ currentAnswerInfo.title }}</strong>
                    <span v-if="currentAnswerInfo.artist"> &mdash; {{ currentAnswerInfo.artist }}</span>
                  </p>

                  <!-- Qui a trouvé ? -->
                  <div v-if="currentAnswerInfo.foundBy.length > 0" class="found-by">
                    <h4>Ont trouvé :</h4>
                    <ul>
                      <li
                        v-for="(f, i) in currentAnswerInfo.foundBy"
                        :key="f.username"
                        :class="{ 'found-first': i === 0 }"
                      >
                        <span class="found-rank">{{ i === 0 ? '🥇' : `#${i + 1}` }}</span>
                        <span class="found-name">{{ f.username }}</span>
                        <span class="found-time">{{ (f.time_ms / 1000).toFixed(1) }}s</span>
                        <span v-if="f.bonus > 0" class="found-bonus">+{{ f.bonus }} bonus</span>
                      </li>
                    </ul>
                  </div>
                  <p v-else class="found-nobody">Personne n'a trouvé cette fois ! 😅</p>

                  <video
                    v-if="currentAnswerInfo.videoUrl"
                    ref="videoEl"
                    :src="currentAnswerInfo.videoUrl"
                    controls
                    :aria-label="`Générique de ${currentAnswerInfo.animeName}`"
                    style="width: 100%; max-width: 600px; border-radius: 8px; margin-top: 14px;"
                  ></video>
                  <p v-else class="no-video">Pas de vidéo disponible pour cette piste.</p>
                </div>

                <div v-else>
                  <p>🎵 Écoutez attentivement...</p>

                  <GameTimer :duration="roundDuration" :key="currentAudioUrl" />

                  <audio
                    v-if="currentAudioUrl"
                    v-show="!audioFailed"
                    ref="audioEl"
                    :src="currentAudioUrl"
                    :aria-label="`Extrait audio — trouvez le nom de l'anime`"
                    controls
                    @error="onAudioError"
                  ></audio>

                  <div v-if="audioFailed" class="audio-failed" role="alert">
                    ⚠️ L'extrait audio n'a pas pu être chargé.
                    <button type="button" class="audio-retry" @click="retryAudio">Réessayer</button>
                  </div>

                  <div v-if="isSpectator" class="spectator-watching" role="status">
                    👁 Vous regardez la partie en tant que spectateur
                  </div>

                  <div v-else class="answer-zone">
                    <AnimeAutocomplete
                      v-model="userGuess"
                      :dictionary="animeDictionary"
                      input-id="anime-guess"
                      @submit="submitAnswer"
                    />
                  </div>

                  <!-- Vote pour passer -->
                  <div v-if="!isSpectator" class="skip-zone">
                    <button
                      @click="sendSkipVote"
                      :disabled="hasVotedSkip"
                      class="btn-skip"
                      :title="hasVotedSkip ? 'Vous avez déjà voté' : 'Voter pour passer cette piste'"
                    >
                      ⏭ Passer
                      <span v-if="skipVotes.votes > 0" class="skip-count">
                        {{ skipVotes.votes }}/{{ skipVotes.needed }}
                      </span>
                    </button>
                  </div>

                  <!-- Contrôles hôte -->
                  <div v-if="isCreator" class="host-controls">
                    <span class="host-badge">🎮 Hôte</span>
                    <button @click="forceSkip" class="btn-force-skip">⏭ Forcer la piste</button>
                  </div>
                </div>

                <ReactionOverlay
                  ref="reactionOverlay"
                  :connected="isConnected"
                  @react="sendReaction"
                />

                <div class="leaderboard" style="margin-top: 20px;" aria-label="Scores en cours">
                  <h3>Classement</h3>
                  <ul>
                    <li v-for="p in players" :key="p.id">
                      {{ p.username }}: {{ p.score }} pts
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <aside class="chat-aside" aria-label="Chat de la partie">
            <ChatPanel
              :messages="chatMessages"
              :ownUsername="authStore.user?.username"
              :connected="isConnected"
              @send="sendChat"
            />
          </aside>

          <!-- Barre de navigation mobile (visible uniquement < 768px) -->
          <nav class="mobile-tabs" aria-label="Navigation du jeu">
            <button
              :class="{ active: mobileTab === 'players' }"
              @click="mobileTab = 'players'"
              :aria-pressed="mobileTab === 'players'"
            >
              <span aria-hidden="true">👥</span> Joueurs
            </button>
            <button
              :class="{ active: mobileTab === 'game' }"
              @click="mobileTab = 'game'"
              :aria-pressed="mobileTab === 'game'"
            >
              <span aria-hidden="true">🎮</span> Jeu
            </button>
            <button
              :class="{ active: mobileTab === 'chat' }"
              @click="mobileTab = 'chat'"
              :aria-pressed="mobileTab === 'chat'"
            >
              <span aria-hidden="true">💬</span> Chat
            </button>
          </nav>
        </div>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from "vue";
import RoomSelection from "./components/RoomSelection.vue";
import GameTimer from "./components/GameTimer.vue";
import AuthForm from "./components/AuthForm.vue";
import ProfilePage from "./components/ProfilePage.vue";
import NewsPage from "./components/NewsPage.vue";
import GameSettings from "./components/GameSettings.vue";
import LeaderboardPage from "./components/LeaderboardPage.vue";
import ChatPanel from "./components/ChatPanel.vue";
import LandingPage from "./components/LandingPage.vue";
import ReactionOverlay from "./components/ReactionOverlay.vue";
import AnimeAutocomplete from "./components/AnimeAutocomplete.vue";
import AppHeader from "./components/AppHeader.vue";
import HomeDashboard from "./components/HomeDashboard.vue";
import ToastContainer from "./components/ToastContainer.vue";
import { authStore } from "./authStore";
import { useToast } from "./composables/useToast";
import { API_URL, WS_URL } from "./config";

const toast = useToast();

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
  trackType: "",
  difficulty: 0,
  foundBy: [],
});
const finalScores = ref([]);
const roundHistory = ref([]);
const skipVotes = ref({ votes: 0, needed: 1 });
const hasVotedSkip = ref(false);
const reconnectMsg = ref("");
const showLanding = ref(true);
const showPublicLeaderboard = ref(false);
const currentView = ref("home");
const isCreator = ref(false);

const navigateTo = (view) => {
  currentView.value = view;
};
const roomSettings = ref({ maxRounds: 5, roundDuration: 20, filterType: "", decade: 0, isPrivate: false, password: "" });
const chatMessages = ref([]);
const isSpectator = ref(false);
const spectatorCount = ref(0);
const mobileTab = ref("game");
const reactionOverlay = ref(null);
const audioEl = ref(null);
const videoEl = ref(null);
const audioFailed = ref(false);

// Play audio after Vue renders the new src into the DOM
watch(currentAudioUrl, async (url) => {
  if (!url) return;
  audioFailed.value = false;
  await nextTick();
  if (!audioEl.value) return;
  audioEl.value.load();
  audioEl.value.play().catch(() => {});
});

// Le clip est servi depuis un mirror externe : il peut être mort/indisponible.
// On le signale clairement plutôt que de laisser un lecteur muet.
const onAudioError = () => {
  if (currentAudioUrl.value) audioFailed.value = true;
};

const retryAudio = async () => {
  audioFailed.value = false;
  await nextTick();
  if (!audioEl.value) return;
  audioEl.value.load();
  audioEl.value.play().catch(() => {});
};

// Play video after Vue renders the reveal panel
watch(() => currentAnswerInfo.value.videoUrl, async (url) => {
  if (!url) return;
  await nextTick();
  videoEl.value?.play().catch(() => {});
});

let socket = null;
let reconnectAttempts = 0;
let intentionalClose = false;

const difficultyLabel = (d) => {
  if (d >= 80) return "Facile";
  if (d >= 50) return "Moyen";
  if (d >= 20) return "Difficile";
  return "Expert";
};

const startGame = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "START_GAME", payload: null }));
  }
};

const backToLobby = () => {
  finalScores.value = [];
  roundHistory.value = [];
  skipVotes.value = { votes: 0, needed: 1 };
  hasVotedSkip.value = false;
  state.value = "LOBBY";
};

const sendSkipVote = () => {
  if (socket && socket.readyState === WebSocket.OPEN && !hasVotedSkip.value) {
    socket.send(JSON.stringify({ type: "VOTE_SKIP", payload: null }));
    hasVotedSkip.value = true;
  }
};

const forceSkip = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "FORCE_SKIP", payload: null }));
  }
};

const kickPlayer = (username) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "KICK_PLAYER", payload: username }));
  }
};

const submitAnswer = () => {
  if (!userGuess.value) return;
  socket.send(JSON.stringify({ type: "SUBMIT_ANSWER", payload: userGuess.value }));
  userGuess.value = "";
};

const connectAnilist = () => {
  window.location.href = `${API_URL}/api/auth/anilist?token=${authStore.token}`
}

const connectMAL = () => {
  window.location.href = `${API_URL}/api/auth/mal?token=${authStore.token}`
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
    const response = await fetch(`${API_URL}/animes`);
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
  const wsUrl = `${WS_URL}/ws?room=${room_id}&password=${password || ""}&token=${authStore.token}`;
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
          players.value = data.payload.players ?? [];
          spectatorCount.value = data.payload.spectator_count ?? 0;
          break;
        case "SPECTATOR_STATUS":
          isSpectator.value = data.payload;
          break;
        case "GAME_STATE":
          state.value = data.payload;
          break;
        case "NewQuestion":
          isRevealing.value = false;
          currentAudioUrl.value = data.payload.audio_url;
          roundDuration.value = data.payload.duration;
          hasVotedSkip.value = false;
          skipVotes.value = { votes: 0, needed: 1 };
          // playback is driven by the watch(currentAudioUrl) watcher above
          break;
        case "ROUND_ENDED":
          isRevealing.value = true;
          currentAnswerInfo.value = {
            animeName: data.payload.answer,
            title: data.payload.title || "",
            artist: data.payload.artist || "",
            videoUrl: data.payload.video_url || "",
            trackType: data.payload.track_type || "",
            difficulty: data.payload.difficulty || 0,
            foundBy: data.payload.found_by || [],
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
        case "NOTICE":
          toast.info(data.payload, { title: "Info partie" });
          break;
        case "SKIP_VOTE_UPDATE":
          skipVotes.value = { votes: data.payload.votes, needed: data.payload.needed };
          break;
        case "HOST_CHANGED":
          isCreator.value = data.payload === authStore.user?.username;
          break;
        case "KICKED":
          disconnect();
          toast.error(data.payload ?? "Vous avez été expulsé de la partie.", { title: "Expulsé" });
          break;
        case "GAME_OVER":
          finalScores.value = [...players.value].sort((a, b) => b.score - a.score);
          roundHistory.value = data.payload.history ?? [];
          state.value = "GAME_OVER";
          break;
        case "CHAT_MESSAGE":
          chatMessages.value.push({
            username: data.payload.username,
            message: data.payload.message,
          });
          if (chatMessages.value.length > 200) chatMessages.value.shift();
          break;
        case "REACTION_BROADCAST":
          reactionOverlay.value?.addParticle(data.payload.emoji);
          break;
        case "XP_GAINED": {
          const oldLevel = authStore.user?.level ?? 1;
          const levelUp  = data.payload.new_level > oldLevel;
          toast.xp({
            xpGained: data.payload.xp_gained,
            newXP:    data.payload.new_xp,
            newLevel: data.payload.new_level,
            levelUp,
          });
          if (authStore.user) {
            authStore.setUser(
              { ...authStore.user, xp: data.payload.new_xp, level: data.payload.new_level },
              authStore.token
            );
          }
          break;
        }
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

const sendReaction = (emoji) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "REACTION", payload: emoji }));
  }
};

const sendChat = (text) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "CHAT", payload: text }));
  }
};

const disconnect = () => {
  intentionalClose = true;
  chatMessages.value = [];
  isSpectator.value = false;
  spectatorCount.value = 0;
  mobileTab.value = "game";
  if (socket) socket.close();
};

defineExpose({ state, isConnected });
</script>

<style>
/* ── Layout principal ───────────────────────────────────── */
main { flex: 1; display: flex; flex-direction: column; }
.app-main { flex: 1; display: flex; flex-direction: column; }

.lobby-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

/* ── Layout jeu ─────────────────────────────────────────── */
.game-layout {
  display: flex;
  gap: 0;
  flex: 1;
  min-height: calc(100vh - 64px);
  overflow: hidden;
}

/* ── Vue News (placeholder) ─────────────────────────────── */
.news-view {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f0f23;
  color: #94a3b8;
  font-size: 1rem;
  min-height: calc(100vh - 64px);
}

.sidebar {
  width: 200px;
  flex-shrink: 0;
  background: #0f0f23;
  border-right: 1px solid rgba(255,255,255,0.07);
  padding: 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}
.sidebar h3 {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #94a3b8;
  margin-bottom: 8px;
}
.sidebar li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #cbd5e1;
  background: rgba(255,255,255,0.03);
}
.sidebar li small { color: #f97316; font-weight: 600; font-size: 0.78rem; }

.game-area {
  flex: 1;
  background: #16213e;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.chat-aside {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid rgba(255,255,255,0.07);
}

/* ── Status bar ─────────────────────────────────────────── */
.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255,255,255,0.07);
  padding-bottom: 14px;
  margin-bottom: 20px;
  font-size: 0.88rem;
  color: #94a3b8;
}
.status-bar strong { color: #f1f5f9; }
.btn-quit {
  background: rgba(239,68,68,0.15);
  color: #f87171;
  border: 1px solid rgba(239,68,68,0.3);
  padding: 5px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 600;
  transition: background 0.15s;
}
.btn-quit:hover { background: rgba(239,68,68,0.25); }

/* ── Quiz box ───────────────────────────────────────────── */
.quiz-box {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 12px;
  padding: 24px;
  flex: 1;
}
.btn-start {
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: white;
  border: none;
  padding: 12px;
  border-radius: 8px;
  font-weight: 700;
  font-size: 1rem;
  cursor: pointer;
  width: 100%;
  box-shadow: 0 4px 14px rgba(249,115,22,0.3);
  transition: transform 0.15s, box-shadow 0.15s;
}
.btn-start:hover { transform: translateY(-1px); box-shadow: 0 6px 18px rgba(249,115,22,0.4); }

/* ── Réponse ────────────────────────────────────────────── */
.answer-zone { margin: 20px 0; }

/* ── Vote skip & Contrôles hôte ─────────────────────────── */
.skip-zone { margin: 10px 0 4px; display: flex; align-items: center; gap: 8px; }
.btn-skip {
  background: transparent;
  border: 1px solid rgba(255,255,255,0.15);
  color: #94a3b8;
  padding: 6px 14px;
  border-radius: 7px;
  font-size: 0.82rem;
  cursor: pointer;
  transition: all 0.15s;
  display: flex; align-items: center; gap: 6px;
}
.btn-skip:hover:not(:disabled) { border-color: #f97316; color: #f97316; }
.btn-skip:disabled { opacity: 0.4; cursor: not-allowed; }
.skip-count {
  background: rgba(249,115,22,0.15);
  color: #fb923c;
  padding: 1px 7px;
  border-radius: 99px;
  font-size: 0.75rem;
  font-weight: 700;
}
.host-controls {
  display: flex; align-items: center; gap: 8px;
  margin: 6px 0 12px;
  padding: 7px 12px;
  background: rgba(249,115,22,0.06);
  border: 1px solid rgba(249,115,22,0.2);
  border-radius: 8px;
}
.host-badge { font-size: 0.78rem; font-weight: 700; color: #fb923c; }
.btn-force-skip {
  background: rgba(249,115,22,0.15);
  border: 1px solid rgba(249,115,22,0.3);
  color: #f97316;
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-force-skip:hover { background: rgba(249,115,22,0.28); }

/* ── Kick button (sidebar) ──────────────────────────────── */
.btn-kick {
  background: transparent;
  border: none;
  color: #475569;
  cursor: pointer;
  font-size: 0.7rem;
  padding: 1px 4px;
  border-radius: 4px;
  margin-left: auto;
  transition: color 0.15s;
  line-height: 1;
}
.btn-kick:hover { color: #ef4444; }

/* ── Résumé de fin de partie ────────────────────────────── */
.round-history {
  margin-top: 24px;
  text-align: left;
  max-height: 320px;
  overflow-y: auto;
  padding-right: 4px;
}
.round-history h3 { font-size: 0.9rem; color: #94a3b8; margin-bottom: 10px; text-transform: uppercase; letter-spacing: .05em; }
.round-item {
  background: #16213e;
  border-radius: 8px;
  padding: 10px 14px;
  margin-bottom: 6px;
  border-left: 3px solid #1e2a45;
  transition: border-color 0.15s;
}
.round-item:hover { border-left-color: #f97316; }
.round-item-header { display: flex; align-items: center; gap: 8px; margin-bottom: 3px; }
.round-num { font-size: 0.72rem; color: #475569; font-weight: 700; text-transform: uppercase; }
.round-tag { font-size: 0.7rem; background: rgba(249,115,22,0.15); color: #fb923c; padding: 1px 6px; border-radius: 4px; font-weight: 600; }
.round-anime { font-weight: 700; color: #f1f5f9; font-size: 0.9rem; }
.round-item-track { font-size: 0.78rem; color: #64748b; margin-bottom: 5px; }
.round-finders { display: flex; flex-wrap: wrap; gap: 6px; }
.round-finder {
  font-size: 0.78rem; color: #94a3b8;
  background: rgba(255,255,255,0.04);
  padding: 2px 8px; border-radius: 5px;
  display: flex; align-items: center; gap: 4px;
}
.round-finder em { color: #475569; font-style: normal; }
.round-bonus { color: #34d399; font-size: 0.72rem; font-weight: 700; }
.round-nobody { font-size: 0.78rem; color: #475569; font-style: italic; }
/* styles input/button déplacés dans AnimeAutocomplete.vue */

/* Classe utilitaire accessibilité (label visually hidden) */
.sr-only {
  position: absolute;
  width: 1px; height: 1px;
  padding: 0; margin: -1px;
  overflow: hidden;
  clip: rect(0,0,0,0);
  white-space: nowrap;
  border: 0;
}

/* ── Reveal / Game Over ─────────────────────────────────── */
.reveal-zone {
  text-align: center;
  padding: 20px 0;
}
.reveal-zone h2 { font-size: 1.4rem; margin-bottom: 10px; }
.answer-name { color: #f9a8d4; font-weight: 700; }
.no-video { font-style: italic; margin-top: 20px; color: #94a3b8; }

.audio-failed {
  margin-top: 16px;
  padding: 10px 14px;
  background: rgba(251, 191, 36, 0.12);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: 8px;
  color: #fbbf24;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
  flex-wrap: wrap;
}
.audio-retry {
  background: #f97316;
  color: #fff;
  border: none;
  padding: 5px 12px;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.82rem;
  cursor: pointer;
}
.audio-retry:hover { opacity: 0.88; }

/* ── Reveal enrichi ─────────────────────────────────────── */
.reveal-header { display: flex; gap: 8px; justify-content: center; margin-bottom: 10px; }

.reveal-tag {
  font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em;
  padding: 3px 10px; border-radius: 20px;
  background: rgba(59,130,246,0.2); color: #93c5fd; border: 1px solid rgba(59,130,246,0.3);
}
.reveal-difficulty {
  font-size: 0.72rem; font-weight: 700;
  padding: 3px 10px; border-radius: 20px;
}
.reveal-difficulty[data-level="Facile"]   { background: rgba(34,197,94,0.15);  color: #86efac; border: 1px solid rgba(34,197,94,0.3); }
.reveal-difficulty[data-level="Moyen"]    { background: rgba(250,204,21,0.15); color: #fde68a; border: 1px solid rgba(250,204,21,0.3); }
.reveal-difficulty[data-level="Difficile"]{ background: rgba(249,115,22,0.15); color: #fdba74; border: 1px solid rgba(249,115,22,0.3); }
.reveal-difficulty[data-level="Expert"]   { background: rgba(239,68,68,0.15);  color: #fca5a5; border: 1px solid rgba(239,68,68,0.3); }

.reveal-track-info { color: #94a3b8; font-size: 0.95rem; margin: 4px 0 16px; }

.found-by { margin: 16px auto; max-width: 360px; text-align: left; }
.found-by h4 {
  font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.08em;
  color: #64748b; margin-bottom: 8px;
}
.found-by ul { display: flex; flex-direction: column; gap: 6px; }
.found-by li {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 14px; border-radius: 8px;
  background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.07);
  font-size: 0.88rem; color: #cbd5e1;
}
.found-by li.found-first { border-color: rgba(255,215,0,0.35); background: rgba(255,215,0,0.07); }
.found-rank { width: 28px; flex-shrink: 0; }
.found-name { flex: 1; font-weight: 600; color: #f1f5f9; }
.found-time { color: #64748b; font-size: 0.8rem; }
.found-bonus { color: #fbbf24; font-size: 0.78rem; font-weight: 700; }
.found-nobody { color: #64748b; font-style: italic; font-size: 0.9rem; margin: 14px 0; }
.game-over-screen { text-align: center; padding: 20px; }
.game-over-screen h2 { font-size: 1.8rem; margin-bottom: 20px; color: #f1f5f9; }
.final-scores { padding: 0; max-width: 420px; margin: 0 auto; }
.final-scores li {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 16px; margin-bottom: 8px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 10px;
  font-size: 1rem;
  color: #e2e8f0;
}
.final-scores li.gold { background: rgba(255,215,0,0.08); border-color: rgba(255,215,0,0.3); }
.final-scores li.me  { border-color: #f97316; }
.final-scores .rank  { font-size: 1.3rem; width: 32px; }
.final-scores .pname { flex: 1; }
.final-scores .pts   { color: #f97316; font-size: 0.9rem; font-weight: 600; }
.spectator-badge {
  background: rgba(249,115,22,0.12);
  color: #f97316;
  border: 1px solid rgba(249,115,22,0.3);
  font-size: 0.78rem;
  font-weight: 700;
  padding: 5px 10px;
  border-radius: 8px;
  margin-bottom: 12px;
  text-align: center;
}
.spectator-count {
  margin-top: 14px;
  font-size: 0.78rem;
  color: #475569;
  text-align: center;
}
.spectator-watching {
  margin: 20px 0;
  padding: 16px;
  background: rgba(59,130,246,0.08);
  border: 1px dashed rgba(59,130,246,0.3);
  border-radius: 10px;
  color: #93c5fd;
  font-size: 0.9rem;
  text-align: center;
}
.reconnect-banner {
  background: rgba(250,204,21,0.1);
  color: #fde68a;
  border: 1px solid rgba(250,204,21,0.25);
  border-radius: 8px;
  padding: 8px 14px;
  margin-bottom: 12px;
  font-size: 0.88rem;
}

/* XP / notification toasts are handled by ToastContainer.vue */

/* ── Leaderboard public (avant login) ───────────────────── */
.public-leaderboard {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #0f0f23;
  padding-top: 10px;
}
.btn-back-landing {
  display: inline-block;
  margin: 8px 0 0 24px;
  background: none;
  border: none;
  color: #f97316;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  padding: 6px 0;
}
.btn-back-landing:hover { text-decoration: underline; }

/* ── In-game leaderboard mini ───────────────────────────── */
.game-area .leaderboard { margin-top: 20px; }
.game-area .leaderboard h3 {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: #94a3b8;
  margin-bottom: 8px;
}
.game-area .leaderboard ul li {
  display: flex;
  justify-content: space-between;
  padding: 5px 0;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  font-size: 0.88rem;
  color: #cbd5e1;
}

/* ── Barre de navigation mobile ─────────────────────────── */
.mobile-tabs { display: none; }

/* ── Responsive ≤ 768px ─────────────────────────────────── */
@media (max-width: 768px) {
  .game-layout {
    flex-direction: column;
    min-height: unset;
    padding-bottom: 56px; /* espace pour la nav mobile */
    overflow: visible;
  }

  /* Sur mobile : seul le panneau actif est visible */
  .game-layout .sidebar,
  .game-layout .game-area,
  .game-layout .chat-aside { display: none; }

  .game-layout[data-mobile-tab="players"] .sidebar  { display: flex; flex-direction: column; width: 100%; border-right: none; border-bottom: 1px solid rgba(255,255,255,0.07); padding: 16px; min-height: calc(100vh - 64px - 56px); }
  .game-layout[data-mobile-tab="game"]    .game-area { display: flex; }
  .game-layout[data-mobile-tab="chat"]    .chat-aside { display: flex; flex-direction: column; width: 100%; border-left: none; height: calc(100vh - 64px - 56px); }

  .game-area { padding: 14px 16px; }
  .answer-zone { flex-direction: column; }
  .answer-zone input { width: 100%; }
  .answer-zone button { width: 100%; }

  /* Barre de navigation mobile */
  .mobile-tabs {
    display: flex;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 56px;
    background: #0f0f23;
    border-top: 1px solid rgba(255,255,255,0.1);
    z-index: 200;
  }
  .mobile-tabs button {
    flex: 1;
    background: none;
    border: none;
    border-top: 2px solid transparent;
    color: #64748b;
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    transition: color 0.15s, border-color 0.15s;
    padding: 0;
  }
  .mobile-tabs button span { font-size: 1.1rem; }
  .mobile-tabs button.active {
    color: #f97316;
    border-top-color: #f97316;
  }
  .mobile-tabs button:focus-visible {
    outline: 2px solid #f97316;
    outline-offset: -2px;
  }

  .sidebar { width: 100%; border-right: none; }
  .chat-aside { width: 100%; border-left: none; }
}

/* ── Responsive ≤ 480px ─────────────────────────────────── */
@media (max-width: 480px) {
  .status-bar { flex-direction: column; align-items: flex-start; gap: 8px; }
  .status-bar .btn-quit { align-self: flex-end; }
  .quiz-box { padding: 16px; }
  .final-scores { max-width: 100%; }
  .room-selection-container { padding: 20px 14px; }
}
</style>
