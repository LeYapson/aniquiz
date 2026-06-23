<template>
  <div class="speedrun-page">

    <!-- Écran de démarrage -->
    <div v-if="phase === 'idle'" class="center-card">
      <h1>Mode Speed Run</h1>
      <p class="subtitle">
        Retrouve un maximum d'animes en <strong>5 minutes</strong>.<br>
        Dès que tu trouves, la piste suivante arrive immédiatement.
      </p>
      <button class="btn-primary btn-large" @click="startGame">Lancer la partie</button>
      <router-link to="/" class="btn-ghost btn-full">Retour au lobby</router-link>
    </div>

    <!-- Écran de jeu -->
    <div v-else-if="phase === 'playing'" class="game-wrap">

      <!-- Barre supérieure -->
      <header class="game-topbar">
        <div class="topbar-left">
          <div class="timer" :class="{ danger: timeLeft <= 30 }">{{ formattedTime }}</div>
          <div class="timer-bar-track">
            <div class="timer-bar-fill" :style="{ width: timerPercent + '%' }" :class="{ danger: timeLeft <= 30 }" />
          </div>
        </div>
        <div class="topbar-center">
          <div class="score-block">
            <span class="score-num">{{ score }}</span>
            <span class="score-lbl">anime{{ score > 1 ? 's' : '' }}</span>
          </div>
          <div class="streak-block" v-if="streak >= 2">
            <span class="streak-fire">🔥</span>
            <span class="streak-num">{{ streak }}</span>
          </div>
        </div>
        <div class="topbar-right">
          <button
            v-if="!confirmQuit"
            class="btn-quit"
            @click="confirmQuit = true"
          >Quitter la partie</button>
          <div v-else class="quit-confirm">
            <span>Abandonner ?</span>
            <button class="btn-quit-yes" @click="abandonSession">Oui, quitter</button>
            <button class="btn-quit-no" @click="confirmQuit = false">Annuler</button>
          </div>
        </div>
      </header>

      <!-- Corps du jeu : 2 colonnes sur desktop -->
      <div class="game-body">

        <!-- Colonne gauche : visualiseur + stats -->
        <div class="col-visual">
          <div class="audio-visual">
            <audio ref="audioEl" :src="currentTrack.audio_url" autoplay @ended="onAudioEnded" />
            <div class="rings">
              <div class="ring ring-1" :class="{ active: isPlaying }" />
              <div class="ring ring-2" :class="{ active: isPlaying }" />
              <div class="ring ring-3" :class="{ active: isPlaying }" />
            </div>
            <div class="pulse-core" />
          </div>
          <p class="audio-hint">🎵 Écoute et trouve l'anime !</p>

          <div class="side-stats">
            <div class="stat-row">
              <span class="stat-lbl">Pistes jouées</span>
              <span class="stat-val">{{ tracksPlayed }}</span>
            </div>
            <div class="stat-row">
              <span class="stat-lbl">Pistes skippées</span>
              <span class="stat-val">{{ tracksSkipped }}</span>
            </div>
            <div class="stat-row">
              <span class="stat-lbl">Précision</span>
              <span class="stat-val">{{ accuracyDisplay }}</span>
            </div>
          </div>
        </div>

        <!-- Colonne droite : formulaire réponse -->
        <div class="col-form">
          <div class="answer-form">
            <label class="form-label">Nom de l'anime</label>
            <AnimeAutocomplete
              ref="answerInput"
              v-model="answer"
              :dictionary="animeDictionary"
              :show-submit="false"
              placeholder="Tape ta réponse..."
              input-id="speedrun-guess"
              @submit="submitAnswer"
            />
            <div v-if="feedback" class="feedback" :class="feedbackClass">
              {{ feedback }}
            </div>
            <button type="button" class="btn-primary btn-full" @click="submitAnswer" :disabled="!answer.trim() || isSubmitting">
              Valider
            </button>
            <button type="button" class="btn-skip btn-full" @click="skipTrack" :disabled="isSubmitting">
              Abandonner cette piste →
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Écran de fin -->
    <div v-else-if="phase === 'finished'" class="center-card">
      <div class="result-icon">🎌</div>
      <h1>{{ abandoned ? 'Partie terminée' : 'Temps écoulé !' }}</h1>
      <div class="final-score">
        <span class="score-big">{{ score }}</span>
        <span class="score-unit">anime{{ score > 1 ? 's' : '' }} trouvé{{ score > 1 ? 's' : '' }}</span>
      </div>
      <div class="end-stats">
        <div class="end-stat">
          <strong>{{ tracksPlayed }}</strong><span>pistes jouées</span>
        </div>
        <div class="end-stat-sep" />
        <div class="end-stat">
          <strong>{{ tracksSkipped }}</strong><span>skippées</span>
        </div>
        <div class="end-stat-sep" />
        <div class="end-stat">
          <strong>{{ accuracyDisplay }}</strong><span>précision</span>
        </div>
      </div>
      <div v-if="rank" class="rank-info">
        Tu es <strong>#{{ rank }}</strong> au classement speed run !
      </div>
      <div class="end-actions">
        <button class="btn-primary btn-large" @click="startGame">Rejouer</button>
        <router-link to="/leaderboard" class="btn-ghost btn-full">Voir le classement</router-link>
        <router-link to="/" class="btn-ghost btn-full">Retour au lobby</router-link>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { API_URL } from '../config.ts'
import { authStore } from '../authStore.js'
import AnimeAutocomplete from '../components/AnimeAutocomplete.vue'

const phase        = ref('idle')
const sessionId    = ref(null)
const score        = ref(0)
const timeLeft     = ref(300)
const currentTrack = ref({ id: null, audio_url: '' })
const answer       = ref('')
const feedback     = ref('')
const feedbackClass= ref('')
const isSubmitting = ref(false)
const isPlaying    = ref(false)
const rank         = ref(null)
const confirmQuit  = ref(false)
const abandoned    = ref(false)
const streak       = ref(0)
const tracksPlayed = ref(0)
const tracksSkipped= ref(0)
const animeDictionary = ref([])

const audioEl    = ref(null)
const answerInput= ref(null)
let timerInterval= null

const formattedTime = computed(() => {
  const m = Math.floor(timeLeft.value / 60)
  const s = timeLeft.value % 60
  return `${m}:${s.toString().padStart(2, '0')}`
})

const timerPercent = computed(() => (timeLeft.value / 300) * 100)

const accuracyDisplay = computed(() => {
  if (tracksPlayed.value === 0) return '—'
  return Math.round((score.value / tracksPlayed.value) * 100) + '%'
})

function authHeaders() {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${authStore.token}`,
  }
}

async function startGame() {
  feedback.value   = ''
  answer.value     = ''
  score.value      = 0
  rank.value       = null
  abandoned.value  = false
  confirmQuit.value= false
  streak.value     = 0
  tracksPlayed.value  = 0
  tracksSkipped.value = 0

  const res = await fetch(`${API_URL}/api/speedrun/start`, {
    method: 'POST',
    headers: authHeaders(),
  })
  if (!res.ok) return

  const data = await res.json()
  sessionId.value    = data.session_id
  currentTrack.value = data.track
  timeLeft.value     = 300
  phase.value        = 'playing'
  isPlaying.value    = true
  tracksPlayed.value = 1

  startTimer()
  focusInput()
}

function startTimer() {
  clearInterval(timerInterval)
  timerInterval = setInterval(async () => {
    timeLeft.value--
    if (timeLeft.value <= 0) {
      clearInterval(timerInterval)
      await finishGame()
    }
  }, 1000)
}

async function finishGame() {
  clearInterval(timerInterval)
  isPlaying.value = false
  if (audioEl.value) audioEl.value.pause()

  const res = await fetch(`${API_URL}/api/speedrun/finish`, {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId.value }),
  })

  phase.value = 'finished'
  if (!res.ok) return
  await fetchRank()
}

async function abandonSession() {
  abandoned.value   = true
  confirmQuit.value = false
  await finishGame()
}

async function fetchRank() {
  const res = await fetch(`${API_URL}/api/leaderboard/speedrun`)
  if (!res.ok) return
  const entries = await res.json()
  const entry = entries.find(e => e.user_id === authStore.userID)
  if (entry) rank.value = entry.rank
}

async function submitAnswer() {
  if (!answer.value.trim() || isSubmitting.value) return
  isSubmitting.value = true
  feedback.value = ''

  const res = await fetch(`${API_URL}/api/speedrun/answer`, {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId.value, answer: answer.value }),
  })
  isSubmitting.value = false
  if (!res.ok) return

  const data = await res.json()

  if (data.finished) {
    clearInterval(timerInterval)
    score.value = data.score
    phase.value = 'finished'
    isPlaying.value = false
    await fetchRank()
    return
  }

  if (data.correct) {
    score.value = data.score
    streak.value++
    answer.value = ''
    feedback.value = ''
    currentTrack.value = data.next_track
    tracksPlayed.value++
    focusInput()
  } else {
    streak.value = 0
    feedback.value = "Ce n'est pas ça, continue d'écouter !"
    feedbackClass.value = 'wrong'
    answer.value = ''
    setTimeout(() => { feedback.value = '' }, 1500)
  }
}

async function skipTrack() {
  if (isSubmitting.value) return
  isSubmitting.value = true
  answer.value  = ''
  feedback.value= ''
  streak.value  = 0

  const res = await fetch(`${API_URL}/api/speedrun/skip`, {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId.value }),
  })
  isSubmitting.value = false
  if (!res.ok) return

  const data = await res.json()

  if (data.finished) {
    clearInterval(timerInterval)
    score.value = data.score
    phase.value = 'finished'
    isPlaying.value = false
    await fetchRank()
    return
  }

  tracksSkipped.value++
  tracksPlayed.value++
  currentTrack.value = data.next_track
  focusInput()
}

function onAudioEnded() {
  if (audioEl.value) {
    audioEl.value.currentTime = 0
    audioEl.value.play()
  }
}

function focusInput() {
  setTimeout(() => { answerInput.value?.focus() }, 50)
}

// Charge le dictionnaire d'animes pour l'auto-complétion (aide à la saisie).
onMounted(async () => {
  try {
    const res = await fetch(`${API_URL}/animes`)
    if (res.ok) animeDictionary.value = await res.json()
  } catch (err) {
    console.error('Erreur chargement dictionnaire animes :', err)
  }
})

onUnmounted(() => { clearInterval(timerInterval) })
</script>

<style scoped>
.speedrun-page {
  min-height: 100vh;
  background: var(--navy-2);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

/* ── Écrans centrés (idle / finished) ── */
.center-card {
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 3rem 2.5rem;
  max-width: 480px;
  width: 100%;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
}
.center-card h1 { font-size: 2rem; }

.subtitle { color: var(--text-dim); line-height: 1.7; }
.subtitle strong { color: var(--orange); }

.result-icon { font-size: 3.5rem; }

.final-score {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
}
.score-big { font-size: 5rem; font-weight: 800; color: var(--orange); line-height: 1; }
.score-unit { color: var(--text-dim); font-size: 1.1rem; }

.end-stats {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1rem 1.5rem;
  background: var(--navy-4);
  border: 1px solid var(--border);
  border-radius: 10px;
  width: 100%;
  justify-content: center;
}
.end-stat { display: flex; flex-direction: column; align-items: center; gap: 2px; }
.end-stat strong { font-size: 1.3rem; font-weight: 800; color: var(--orange); }
.end-stat span { font-size: 0.72rem; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.04em; }
.end-stat-sep { width: 1px; height: 32px; background: var(--border); }

.rank-info {
  background: var(--navy-4);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.6rem 1.2rem;
  color: var(--text-dim);
  font-size: 0.95rem;
}
.rank-info strong { color: var(--blue); }

.end-actions { display: flex; flex-direction: column; gap: 0.75rem; width: 100%; }

/* ── Layout jeu pleine page ── */
.game-wrap {
  width: 100%;
  max-width: 1100px;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  align-self: flex-start;
  padding-top: 1.5rem;
}

/* ── Barre supérieure ── */
.game-topbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 1rem;
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1rem 1.5rem;
}

.topbar-left {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.timer {
  font-size: 2.8rem;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  color: var(--blue);
  line-height: 1;
  transition: color 0.3s;
}
.timer.danger { color: #ef4444; }

.timer-bar-track {
  height: 4px;
  background: var(--border);
  border-radius: 2px;
  overflow: hidden;
  width: 160px;
}
.timer-bar-fill {
  height: 100%;
  background: var(--blue);
  border-radius: 2px;
  transition: width 1s linear, background 0.3s;
}
.timer-bar-fill.danger { background: #ef4444; }

.topbar-center {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.score-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.score-num { font-size: 3rem; font-weight: 800; color: var(--orange); line-height: 1; }
.score-lbl { font-size: 0.72rem; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.05em; }

.streak-block {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(249,115,22,0.12);
  border: 1px solid rgba(249,115,22,0.25);
  border-radius: 20px;
  padding: 4px 10px;
}
.streak-fire { font-size: 1.1rem; }
.streak-num { font-size: 1rem; font-weight: 800; color: var(--orange); }

.topbar-right { display: flex; justify-content: flex-end; }

.btn-quit {
  padding: 0.5rem 1rem;
  background: transparent;
  color: #64748b;
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
}
.btn-quit:hover { color: #ef4444; border-color: rgba(239,68,68,0.4); }

.quit-confirm {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
}
.quit-confirm span { color: var(--text-dim); }
.btn-quit-yes {
  padding: 0.4rem 0.8rem;
  background: rgba(239,68,68,0.15);
  color: #f87171;
  border: 1px solid rgba(239,68,68,0.35);
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 600;
  transition: background 0.15s;
}
.btn-quit-yes:hover { background: rgba(239,68,68,0.25); }
.btn-quit-no {
  padding: 0.4rem 0.8rem;
  background: transparent;
  color: var(--text-dim);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.82rem;
  transition: color 0.15s;
}
.btn-quit-no:hover { color: var(--text); }

/* ── Corps 2 colonnes ── */
.game-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  align-items: start;
}

/* ── Colonne gauche : visuel ── */
.col-visual {
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 2.5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.audio-visual {
  position: relative;
  width: 160px;
  height: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
}

audio { display: none; }

.rings { position: absolute; inset: 0; }
.ring {
  position: absolute;
  border-radius: 50%;
  border: 2px solid var(--orange);
  opacity: 0;
  inset: 0;
}
.ring.active { animation: pulse 2s ease-out infinite; }
.ring-2.active { animation-delay: 0.6s; }
.ring-3.active { animation-delay: 1.2s; }

@keyframes pulse {
  0%   { transform: scale(0.3); opacity: 0.7; }
  100% { transform: scale(1);   opacity: 0; }
}

.pulse-core {
  width: 48px;
  height: 48px;
  background: var(--orange);
  border-radius: 50%;
  z-index: 1;
  box-shadow: 0 0 24px rgba(249,115,22,0.4);
}

.audio-hint { color: var(--text-dim); font-size: 0.95rem; }

.side-stats {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  border-top: 1px solid var(--border);
  padding-top: 1.25rem;
}
.stat-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.88rem;
}
.stat-lbl { color: var(--text-dim); }
.stat-val { font-weight: 700; color: var(--text); }

/* ── Colonne droite : formulaire ── */
.col-form {
  background: var(--navy-3);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 2.5rem 2rem;
}

.answer-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-label {
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-dim);
}

.answer-form input {
  width: 100%;
  padding: 1rem 1.1rem;
  background: var(--navy-4);
  border: 1px solid var(--border);
  border-radius: 10px;
  color: var(--text);
  font-size: 1.15rem;
  transition: border-color 0.2s;
}
.answer-form input:focus {
  outline: none;
  border-color: var(--orange);
  box-shadow: 0 0 0 3px rgba(249,115,22,0.1);
}

.feedback {
  font-size: 0.9rem;
  padding: 0.5rem 0.85rem;
  border-radius: 7px;
  margin-top: -0.25rem;
}
.feedback.wrong { background: rgba(239,68,68,0.1); color: #f87171; }

.btn-skip {
  padding: 0.85rem 1.25rem;
  background: transparent;
  color: var(--text-dim);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
  text-align: center;
}
.btn-skip:hover:not(:disabled) { color: var(--text); border-color: rgba(255,255,255,0.2); }
.btn-skip:disabled { opacity: 0.4; cursor: not-allowed; }

/* ── Boutons génériques ── */
.btn-primary {
  padding: 0.85rem 1.25rem;
  background: var(--orange);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-primary:hover:not(:disabled) { background: var(--orange-2); }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-primary.btn-large { font-size: 1.1rem; padding: 1rem 2rem; }

.btn-ghost {
  padding: 0.75rem 1.25rem;
  background: transparent;
  color: var(--text-dim);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  text-align: center;
  text-decoration: none;
  display: inline-block;
  transition: border-color 0.2s, color 0.2s;
}
.btn-ghost:hover { border-color: rgba(255,255,255,0.2); color: var(--text); text-decoration: none; }

.btn-full { width: 100%; }

/* ── Responsive ── */
@media (max-width: 768px) {
  .speedrun-page { align-items: flex-start; }
  .game-wrap { padding-top: 0.75rem; }
  .game-topbar { grid-template-columns: 1fr 1fr; grid-template-rows: auto auto; }
  .topbar-right { grid-column: 1 / -1; justify-content: flex-start; }
  .timer { font-size: 2rem; }
  .timer-bar-track { width: 120px; }
  .score-num { font-size: 2.2rem; }
  .game-body { grid-template-columns: 1fr; }
}
</style>
