<template>
  <div id="app">
    <AppHeader
      v-if="authStore.user"
      :currentView="currentView"
      :inGame="isConnected"
      :isAdmin="isAdmin"
      @navigate="navigateTo"
      @logout="authStore.logout"
      @connect-anilist="connectAnilist"
      @connect-mal="connectMAL"
      @connect-discord="connectDiscord"
      @join-room="onJoinFromInvite"
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

        <AdminPage v-else-if="!isConnected && currentView === 'admin' && isAdmin" />

        <div v-else class="app-main">

        <div v-if="!isConnected" class="lobby-wrapper">
          <HomeDashboard @room-created="setupWebSocket" @room-joined="setupWebSocket" @navigate="navigateTo" />
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

                <!-- Inviter un ami dans ce salon -->
                <div class="invite-zone">
                  <button type="button" class="btn-invite-friend" @click="toggleInvitePicker">
                    👥 Inviter un ami
                  </button>
                  <div v-if="showInvitePicker" class="invite-picker">
                    <p v-if="friendsForInvite.length === 0" class="invite-empty">
                      Aucun ami à inviter. Ajoute-en depuis ton profil !
                    </p>
                    <button
                      v-for="f in friendsForInvite"
                      :key="f.user_id"
                      class="invite-friend-row"
                      @click="inviteFriend(f)"
                    >
                      <span class="invite-friend-dot">{{ f.username.charAt(0).toUpperCase() }}</span>
                      {{ f.username }}
                    </button>
                  </div>
                </div>

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
                <!-- Stats de vitesse de réponse -->
                <div v-if="speedStats.length > 0" class="speed-stats">
                  <h3>⚡ Vitesse de réponse</h3>
                  <table class="speed-table">
                    <thead>
                      <tr><th>Joueur</th><th>Trouvés</th><th>Moyenne</th><th>Plus rapide</th></tr>
                    </thead>
                    <tbody>
                      <tr v-for="(s, i) in speedStats" :key="s.username" :class="{ me: s.username === authStore.user.username }">
                        <td>{{ i === 0 ? '⚡ ' : '' }}{{ s.username }}</td>
                        <td>{{ s.found }}</td>
                        <td>{{ (s.avgMs / 1000).toFixed(1) }}s</td>
                        <td>{{ (s.bestMs / 1000).toFixed(1) }}s</td>
                      </tr>
                    </tbody>
                  </table>
                </div>

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

                  <div v-if="!isSpectator" class="skip-zone" style="margin-top: 12px;">
                    <button
                      @click="sendRevealSkipVote"
                      :disabled="hasVotedRevealSkip"
                      class="btn-skip"
                      :title="hasVotedRevealSkip ? 'Vous avez déjà voté' : 'Voter pour passer la révélation'"
                    >
                      ⏭ Passer la révélation
                      <span v-if="revealSkipVotes.votes > 0" class="skip-count">
                        {{ revealSkipVotes.votes }}/{{ revealSkipVotes.needed }}
                      </span>
                    </button>
                  </div>
                </div>

                <div v-else>
                  <p>🎵 Écoutez et devinez <strong>{{ guessLabel }}</strong>…</p>

                  <GameTimer :duration="roundDuration" :key="currentAudioUrl" />

                  <!-- Élément média piloté par le lecteur maison (pas de contrôles natifs) -->
                  <audio
                    v-if="currentAudioUrl"
                    ref="audioEl"
                    :src="playbackSrc"
                    :aria-label="`Extrait audio — trouvez le nom de l'anime`"
                    @loadedmetadata="onAudioLoaded"
                    @error="onAudioError"
                    @timeupdate="onTimeUpdate"
                    @durationchange="onDurationChange"
                    @play="onPlay"
                    @pause="onPause"
                    @ended="onPause"
                  ></audio>

                  <!-- Lecteur maison : lecture/pause, progression, volume général -->
                  <div v-if="currentAudioUrl && !audioFailed" class="audio-player">
                    <button
                      type="button"
                      class="player-btn"
                      @click="togglePlay"
                      :aria-label="isPlaying ? 'Mettre en pause' : 'Lire'"
                    >{{ isPlaying ? '⏸' : '▶' }}</button>

                    <span class="player-time">{{ formatMediaTime(currentTime) }}</span>
                    <input
                      type="range"
                      class="player-seek"
                      min="0"
                      :max="duration || 0"
                      step="0.1"
                      :value="currentTime"
                      @input="onSeek"
                      aria-label="Position dans l'extrait"
                    />
                    <span class="player-time">{{ formatMediaTime(duration) }}</span>

                    <button
                      type="button"
                      class="volume-btn"
                      @click="toggleMute"
                      :aria-label="volume === 0 ? 'Rétablir le son' : 'Couper le son'"
                    >{{ volumeIcon }}</button>
                    <input
                      type="range"
                      class="volume-slider"
                      min="0"
                      max="1"
                      step="0.01"
                      v-model.number="volume"
                      aria-label="Volume général"
                    />
                  </div>

                  <div v-if="audioFailed" class="audio-failed" role="alert">
                    ⚠️ L'extrait audio n'a pas pu être chargé.
                    <button type="button" class="audio-retry" @click="retryAudio">Réessayer</button>
                  </div>

                  <div v-else-if="audioBlocked" class="audio-blocked" role="alert">
                    🔇 Le son est bloqué par ton navigateur.
                    <button type="button" class="audio-retry" @click="resumeAudio">🔊 Activer le son</button>
                  </div>

                  <div v-if="isSpectator" class="spectator-watching" role="status">
                    👁 Vous regardez la partie en tant que spectateur
                  </div>

                  <div v-else class="answer-zone">
                    <button
                      v-if="buzzerMode && !hasBuzzed"
                      type="button"
                      class="btn-buzzer"
                      @click="onBuzz"
                    >🔔 BUZZER</button>

                    <div v-else-if="guessMode === 'multiple' && roundChoices.length > 0" class="choices-grid">
                      <button
                        v-for="choice in roundChoices"
                        :key="choice"
                        type="button"
                        class="choice-btn"
                        @click="submitAnswer(choice)"
                      >{{ choice }}</button>
                    </div>

                    <AnimeAutocomplete
                      v-else
                      v-model="userGuess"
                      :dictionary="guessMode === 'anime' ? animeDictionary : []"
                      :placeholder="`Devine ${guessLabel}…`"
                      input-id="anime-guess"
                      @submit="submitAnswer"
                    />

                    <p v-if="buzzerMode && buzzedUsers.length > 0" class="buzzed-list" aria-live="polite">
                      🔔 Ont buzzé : {{ buzzedUsers.join(', ') }}
                    </p>
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
import { ref, watch, onMounted, onUnmounted } from "vue";
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
import AdminPage from "./components/AdminPage.vue";
import ToastContainer from "./components/ToastContainer.vue";
import { authStore, apiFetch, consumeSessionExpired } from "./authStore";
import { useToast } from "./composables/useToast";
import { useGameState } from "./composables/useGameState";
import { useAudioPlayer } from "./composables/useAudioPlayer";
import { useGameSocket } from "./composables/useGameSocket";
import { useInvites } from "./composables/useInvites";
import { useOAuth } from "./composables/useOAuth";
import { API_URL } from "./config";

const toast = useToast();

// ── State réactif du jeu ─────────────────────────────────────────────────────
const {
  isConnected, room, players, state, currentAudioUrl, userGuess,
  roundDuration, roundStartFraction, animeDictionary, isRevealing, currentAnswerInfo,
  finalScores, roundHistory, speedStats, skipVotes, hasVotedSkip,
  revealSkipVotes, hasVotedRevealSkip, reconnectMsg, isCreator, roomSettings,
  buzzerMode, guessMode, guessLabel, hasBuzzed, buzzedUsers, chatMessages,
  isSpectator, spectatorCount, mobileTab, roundChoices,
} = useGameState();

// ── Lecteur audio ────────────────────────────────────────────────────────────
const {
  audioEl, videoEl, volume, volumeIcon, isPlaying, currentTime, duration,
  formatMediaTime, playbackSrc, audioFailed, audioBlocked,
  togglePlay, toggleMute, onSeek, onAudioLoaded, onAudioError,
  onPlay, onPause, onTimeUpdate, onDurationChange,
  resumeAudio, retryAudio, releaseMedia,
} = useAudioPlayer({ currentAudioUrl, currentAnswerInfo, roundStartFraction });

// ── Dictionnaire d'autocomplétion ────────────────────────────────────────────
const loadAnimeDictionary = async () => {
  try {
    const response = await fetch(`${API_URL}/animes`);
    if (response.ok) animeDictionary.value = await response.json();
  } catch (err) {
    console.error("Erreur lors du chargement du dictionnaire :", err);
  }
};

// ── WebSocket ────────────────────────────────────────────────────────────────
const reactionOverlay = ref(null);

const {
  socket, setupWebSocket, disconnect,
  startGame, backToLobby, sendSkipVote, forceSkip, sendRevealSkipVote,
  kickPlayer, onBuzz, sendAnswer, sendReaction, sendChat,
} = useGameSocket({
  room, isConnected, players, state, currentAudioUrl, roundDuration, roundStartFraction,
  isRevealing, currentAnswerInfo, finalScores, roundHistory, skipVotes, hasVotedSkip,
  revealSkipVotes, hasVotedRevealSkip, hasBuzzed, buzzedUsers, chatMessages,
  isSpectator, spectatorCount, reconnectMsg, isCreator, roomSettings,
  mobileTab, roundChoices, reactionOverlay, audioEl, videoEl, releaseMedia,
  authStore, toast, loadAnimeDictionary,
});

// submitAnswer conserve le v-model userGuess dans App.vue et délègue l'envoi.
const submitAnswer = (valueOrEvent) => {
  const value = typeof valueOrEvent === 'string' ? valueOrEvent : userGuess.value;
  if (!value) return;
  sendAnswer(value);
  userGuess.value = "";
};

// ── Invitations ──────────────────────────────────────────────────────────────
const { friendsForInvite, showInvitePicker, toggleInvitePicker, inviteFriend } =
  useInvites({ room, authFetch: apiFetch, toast });

// Rejoindre depuis une invitation reçue via la cloche de notifications.
const onJoinFromInvite = (invite) => {
  setupWebSocket({ room_id: invite.room_id, password: invite.password, isCreator: false });
};

// ── OAuth ────────────────────────────────────────────────────────────────────
const { connectAnilist, connectMAL, connectDiscord, checkOAuthCallback } =
  useOAuth({ authStore });

// ── Navigation ───────────────────────────────────────────────────────────────
const showLanding = ref(true);
const showPublicLeaderboard = ref(false);
const currentView = ref("home");
const navigateTo = (view) => { currentView.value = view; };

// ── Statut admin ─────────────────────────────────────────────────────────────
const isAdmin = ref(authStore.user?.is_admin === true);

const loadAdminStatus = async () => {
  if (!authStore.user) { isAdmin.value = false; return; }
  try {
    const res = await apiFetch(`${API_URL}/api/me/admin`);
    if (res.ok) {
      const d = await res.json();
      isAdmin.value = d.is_admin === true;
    }
  } catch { /* silencieux */ }
};

watch(() => authStore.user, (u) => {
  if (u) loadAdminStatus();
  else { isAdmin.value = false; if (currentView.value === "admin") currentView.value = "home"; }
});

// ── Utilitaires template ─────────────────────────────────────────────────────
const difficultyLabel = (d) => {
  if (d >= 80) return "Facile";
  if (d >= 50) return "Moyen";
  if (d >= 20) return "Difficile";
  return "Expert";
};

// ── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(() => {
  if (consumeSessionExpired()) {
    toast.info("Ta session a expiré, reconnecte-toi.", { title: "Session expirée" });
  }
  loadAnimeDictionary();
  checkOAuthCallback();
  loadAdminStatus();
});

onUnmounted(() => {
  releaseMedia(audioEl.value);
  releaseMedia(videoEl.value);
});

defineExpose({ state, isConnected });
</script>

<style>
/* ── Layout principal ───────────────────────────────────── */
main { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.app-main { flex: 1; display: flex; flex-direction: column; min-width: 0; }

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
/* ── Inviter un ami (lobby) ── */
.invite-zone { margin-top: 12px; }
.btn-invite-friend {
  width: 100%;
  background: rgba(255,255,255,0.05);
  color: #cbd5e1;
  border: 1px solid rgba(255,255,255,0.12);
  padding: 10px; border-radius: 8px; font-weight: 600; font-size: 0.9rem; cursor: pointer;
  transition: background 0.15s;
}
.btn-invite-friend:hover { background: rgba(255,255,255,0.1); }
.invite-picker {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 8px;
  padding: 6px;
}
.invite-empty { color: #64748b; font-size: 0.82rem; font-style: italic; padding: 6px 8px; margin: 0; }
.invite-friend-row {
  display: flex; align-items: center; gap: 9px;
  background: none; border: none; cursor: pointer;
  color: #e2e8f0; font-size: 0.88rem; font-weight: 600;
  padding: 7px 8px; border-radius: 6px; text-align: left; width: 100%;
}
.invite-friend-row:hover { background: rgba(249,115,22,0.12); }
.invite-friend-dot {
  width: 24px; height: 24px; border-radius: 50%;
  background: linear-gradient(135deg, #f97316, #ea580c); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.78rem; font-weight: 700; flex-shrink: 0;
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
.speed-stats { margin-top: 24px; text-align: left; }
.speed-stats h3 { font-size: 0.9rem; color: #94a3b8; margin-bottom: 10px; text-transform: uppercase; letter-spacing: .05em; }
.speed-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; background: #16213e; border: 1px solid rgba(255,255,255,0.07); border-radius: 10px; overflow: hidden; }
.speed-table th { background: #0f0f23; color: #f97316; padding: 8px 12px; text-align: left; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; }
.speed-table td { padding: 8px 12px; border-bottom: 1px solid rgba(255,255,255,0.05); color: #cbd5e1; }
.speed-table tr:last-child td { border-bottom: none; }
.speed-table tr.me td { color: #fb923c; font-weight: 700; }

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

.audio-failed,
.audio-blocked {
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
.audio-blocked {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.3);
  color: #93c5fd;
}

/* ── Lecteur audio maison ── */
.audio-player {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  padding: 8px 12px;
  background: #16213e;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  flex-wrap: wrap;
}
.player-btn {
  background: #f97316;
  color: #fff;
  border: none;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.15s;
}
.player-btn:hover { background: #ea580c; }
.player-time {
  font-size: 0.75rem;
  color: #94a3b8;
  min-width: 34px;
  text-align: center;
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}
.player-seek {
  flex: 1 1 120px;
  accent-color: #f97316;
  cursor: pointer;
  height: 4px;
  min-width: 80px;
}
.volume-btn {
  background: transparent;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 2px 4px;
  line-height: 1;
  flex-shrink: 0;
}
.volume-slider {
  flex: 0 1 90px;
  accent-color: #f97316;
  cursor: pointer;
  height: 4px;
  min-width: 60px;
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

/* ── Mode QCM ── */
.choices-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 12px;
}
.choice-btn {
  padding: 14px 10px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.1);
  color: #e2e8f0;
  border-radius: 10px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  text-align: center;
  transition: background 0.15s, border-color 0.15s, transform 0.1s;
  line-height: 1.35;
}
.choice-btn:hover {
  background: rgba(249,115,22,0.12);
  border-color: rgba(249,115,22,0.4);
  color: #fff;
  transform: translateY(-1px);
}
.choice-btn:active { transform: scale(0.97); }

.btn-buzzer {
  margin-top: 8px;
  width: 100%;
  padding: 18px 0;
  font-size: 1.3rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  color: #fff;
  background: linear-gradient(135deg, #f97316, #ef4444);
  border: none;
  border-radius: 12px;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(249, 115, 22, 0.35);
  transition: transform 0.08s ease, box-shadow 0.15s ease;
}
.btn-buzzer:hover { box-shadow: 0 6px 22px rgba(249, 115, 22, 0.5); }
.btn-buzzer:active { transform: scale(0.97); }
.buzzed-list { margin-top: 10px; font-size: 0.82rem; color: #fbbf24; }

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
  min-width: 0;
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

  /* Stats de vitesse : éviter le débordement horizontal */
  .speed-stats { overflow-x: auto; -webkit-overflow-scrolling: touch; }
  .speed-table { font-size: 0.78rem; }
  .speed-table th, .speed-table td { padding: 7px 8px; }

  /* Récap des rounds plus compact */
  .round-history { max-height: 260px; }
}

/* La nav mobile fixe doit respecter la zone sûre (encoches iOS). */
@media (max-width: 768px) {
  .mobile-tabs { padding-bottom: env(safe-area-inset-bottom, 0); height: calc(56px + env(safe-area-inset-bottom, 0)); }
  .game-layout { padding-bottom: calc(56px + env(safe-area-inset-bottom, 0)); }
}
</style>
