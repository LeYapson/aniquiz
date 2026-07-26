<template>
  <div class="daily-page">

    <div class="daily-header">
      <div class="daily-title-row">
        <h1>📅 Quiz du jour</h1>
        <div v-if="!loading && !error" class="daily-countdown">
          Prochain quiz dans <strong>{{ countdown }}</strong>
        </div>
      </div>
      <p class="daily-sub">Une piste par jour, identique pour tous. Qui la trouvera le plus vite ?</p>
    </div>

    <div v-if="loading" class="daily-loading">Chargement…</div>
    <div v-else-if="error" class="daily-error">{{ error }}</div>

    <div v-else class="daily-content">

      <!-- ── Déjà joué ──────────────────────────────────────────── -->
      <template v-if="alreadyPlayed">
        <div class="result-card" :class="result?.found ? 'result-found' : 'result-missed'">
          <div class="result-icon">{{ result?.found ? '🎉' : '😅' }}</div>
          <h2>{{ result?.found ? 'Bien joué !' : 'Pas trouvé aujourd\'hui' }}</h2>
          <p v-if="result?.found">
            Tu as trouvé en <strong>{{ formatMs(result.time_ms) }}</strong>
          </p>
          <div class="result-reveal">
            <span class="reveal-tag" v-if="dailyData.track_type">{{ dailyData.track_type }}</span>
            <strong class="reveal-anime">{{ dailyData.answer }}</strong>
            <span class="reveal-track">{{ dailyData.title }}<span v-if="dailyData.artist"> — {{ dailyData.artist }}</span></span>
          </div>
        </div>
      </template>

      <!-- ── Phase de jeu ───────────────────────────────────────── -->
      <template v-else-if="!submitted">
        <div class="player-zone">
          <p class="player-hint">🎵 Écoute et trouve l'anime !</p>

          <audio
            v-if="dailyData.audio_url"
            ref="audioEl"
            :src="playbackSrc"
            @loadedmetadata="onAudioLoaded"
            @error="onAudioError"
            @timeupdate="onTimeUpdate"
            @durationchange="onDurationChange"
            @play="onPlay"
            @pause="onPause"
            @ended="onPause"
            style="display:none"
          ></audio>

          <div v-if="!audioFailed" class="audio-player">
            <button type="button" class="player-btn" @click="togglePlay" :aria-label="isPlaying ? 'Pause' : 'Lire'">
              {{ isPlaying ? '⏸' : '▶' }}
            </button>
            <span class="player-time">{{ formatMediaTime(currentTime) }}</span>
            <input type="range" class="player-seek" min="0" :max="duration || 0" step="0.1" :value="currentTime" @input="onSeek" />
            <span class="player-time">{{ formatMediaTime(duration) }}</span>
          </div>
          <div v-if="audioFailed" class="audio-failed" role="alert">
            ⚠️ Extrait indisponible.
          </div>

          <div class="timer-display" :class="{ running: timerRunning }">
            ⏱ {{ formatMs(elapsed) }}
          </div>
        </div>

        <!-- Choix QCM -->
        <div class="choices-grid">
          <button
            v-for="choice in dailyData.choices"
            :key="choice"
            type="button"
            class="choice-btn"
            :disabled="submitting"
            @click="submit(choice)"
          >{{ choice }}</button>
        </div>
      </template>

      <!-- ── Révélation après soumission ────────────────────────── -->
      <template v-else>
        <div class="result-card" :class="submitResult?.correct ? 'result-found' : 'result-missed'">
          <div class="result-icon">{{ submitResult?.correct ? '🎉' : '😅' }}</div>
          <h2>{{ submitResult?.correct ? `Trouvé en ${formatMs(elapsed)} !` : 'Pas cette fois…' }}</h2>
          <div class="result-reveal">
            <strong class="reveal-anime">{{ submitResult?.answer }}</strong>
            <span class="reveal-track">
              {{ submitResult?.title }}<span v-if="submitResult?.artist"> — {{ submitResult?.artist }}</span>
            </span>
          </div>
          <video
            v-if="submitResult?.video_url"
            :src="submitResult.video_url"
            controls
            class="reveal-video"
          ></video>
        </div>
      </template>

      <!-- ── Classement du jour ─────────────────────────────────── -->
      <div class="leaderboard-section">
        <h3>🏆 Classement du jour</h3>
        <div v-if="leaderboard.length === 0" class="lb-empty">
          Personne n'a encore joué aujourd'hui — sois le premier !
        </div>
        <ol v-else class="lb-list">
          <li
            v-for="(entry, i) in leaderboard"
            :key="entry.username"
            :class="{ me: entry.username === ownUsername }"
          >
            <span class="lb-rank">{{ i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `#${i + 1}` }}</span>
            <span class="lb-name">{{ entry.username }}</span>
            <span class="lb-time" v-if="entry.found">{{ formatMs(entry.time_ms) }}</span>
            <span class="lb-miss" v-else>—</span>
          </li>
        </ol>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { authStore, apiFetch } from '../authStore'
import { useAudioPlayer } from '../composables/useAudioPlayer'
import { API_URL } from '../config'

const loading    = ref(true)
const error      = ref('')
const dailyData  = ref({})
const alreadyPlayed = ref(false)
const result     = ref(null)
const submitted  = ref(false)
const submitting = ref(false)
const submitResult = ref(null)
const leaderboard  = ref([])

const ownUsername = computed(() => authStore.user?.username ?? '')

// ── Lecteur audio (réutilise le composable) ──────────────────────────────────
const currentAudioUrl  = ref('')
const currentAnswerInfo = ref({ videoUrl: '', foundBy: [] })
const roundStartFraction = ref(0)

const {
  audioEl, volume, isPlaying, currentTime, duration,
  formatMediaTime, playbackSrc, audioFailed,
  togglePlay, onSeek, onAudioLoaded, onAudioError,
  onPlay, onPause, onTimeUpdate, onDurationChange,
  releaseMedia,
} = useAudioPlayer({ currentAudioUrl, currentAnswerInfo, roundStartFraction })

// ── Timer de réponse ──────────────────────────────────────────────────────────
const elapsed      = ref(0)
const timerRunning = ref(false)
let timerStart = 0
let timerInterval = null

const startTimer = () => {
  if (timerRunning.value) return
  timerRunning.value = true
  timerStart = Date.now() - elapsed.value
  timerInterval = setInterval(() => {
    elapsed.value = Date.now() - timerStart
  }, 100)
}

const stopTimer = () => {
  timerRunning.value = false
  clearInterval(timerInterval)
}

// Démarre le timer dès que la lecture commence.
watch(isPlaying, (playing) => {
  if (playing && !submitted.value) startTimer()
})

const formatMs = (ms) => {
  if (!ms) return '—'
  const s = (ms / 1000).toFixed(1)
  return `${s}s`
}

// ── Compte à rebours jusqu'au prochain quiz ──────────────────────────────────
const countdown = ref('')
let countdownInterval = null

const updateCountdown = () => {
  const now = new Date()
  const next = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() + 1))
  const diff = next - now
  const h = Math.floor(diff / 3_600_000).toString().padStart(2, '0')
  const m = Math.floor((diff % 3_600_000) / 60_000).toString().padStart(2, '0')
  const s = Math.floor((diff % 60_000) / 1_000).toString().padStart(2, '0')
  countdown.value = `${h}:${m}:${s}`
}

// ── Chargement ────────────────────────────────────────────────────────────────
const loadLeaderboard = async () => {
  try {
    const res = await apiFetch(`${API_URL}/api/daily/leaderboard`)
    if (res.ok) leaderboard.value = await res.json()
  } catch { /* silencieux */ }
}

onMounted(async () => {
  try {
    const res = await apiFetch(`${API_URL}/api/daily`)
    if (!res.ok) { error.value = 'Impossible de charger le quiz du jour.'; return }
    const data = await res.json()
    dailyData.value = data
    alreadyPlayed.value = data.already_played
    if (data.already_played) result.value = data.result

    if (!data.already_played) {
      roundStartFraction.value = data.start_fraction ?? 0
      currentAudioUrl.value = data.audio_url
    }
  } catch {
    error.value = 'Erreur réseau.'
  } finally {
    loading.value = false
  }

  await loadLeaderboard()
  updateCountdown()
  countdownInterval = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  stopTimer()
  clearInterval(countdownInterval)
  releaseMedia(audioEl.value)
})

// ── Soumission ────────────────────────────────────────────────────────────────
const submit = async (answer) => {
  if (submitting.value) return
  stopTimer()
  submitting.value = true

  try {
    const res = await apiFetch(`${API_URL}/api/daily/submit`, {
      method: 'POST',
      body: JSON.stringify({ answer, time_ms: elapsed.value }),
    })
    if (res.ok) {
      submitResult.value = await res.json()
      submitted.value = true
      releaseMedia(audioEl.value)
      await loadLeaderboard()
    }
  } catch { /* silencieux */ } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.daily-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 32px 24px 64px;
}

/* ── Header ── */
.daily-header { margin-bottom: 28px; }
.daily-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 6px;
}
.daily-title-row h1 { font-size: 1.5rem; font-weight: 800; color: #f1f5f9; margin: 0; }
.daily-countdown {
  font-size: 0.82rem;
  color: #64748b;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  padding: 5px 12px;
  border-radius: 99px;
}
.daily-countdown strong { color: #f97316; }
.daily-sub { color: #64748b; font-size: 0.9rem; margin: 0; }

.daily-loading, .daily-error {
  text-align: center; padding: 60px; color: #64748b; font-size: 0.95rem;
}
.daily-error { color: #f87171; }

/* ── Player ── */
.player-zone {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 14px;
  padding: 24px;
  margin-bottom: 20px;
  text-align: center;
}
.player-hint { color: #94a3b8; font-size: 0.9rem; margin: 0 0 16px; }

.audio-player {
  display: flex; align-items: center; gap: 10px;
  background: #0f0f23; border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px; padding: 8px 12px; flex-wrap: wrap;
}
.player-btn {
  background: #f97316; color: #fff; border: none;
  width: 38px; height: 38px; border-radius: 50%;
  font-size: 1rem; cursor: pointer; flex-shrink: 0;
  transition: background 0.15s;
}
.player-btn:hover { background: #ea580c; }
.player-time { font-size: 0.75rem; color: #94a3b8; min-width: 34px; text-align: center; font-variant-numeric: tabular-nums; }
.player-seek { flex: 1 1 100px; accent-color: #f97316; cursor: pointer; height: 4px; }

.audio-failed { color: #fbbf24; font-size: 0.85rem; margin-top: 12px; }

.timer-display {
  margin-top: 16px;
  font-size: 1.4rem;
  font-weight: 800;
  color: #475569;
  font-variant-numeric: tabular-nums;
  transition: color 0.3s;
}
.timer-display.running { color: #f97316; }

/* ── Choix QCM ── */
.choices-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 28px;
}
.choice-btn {
  padding: 16px 12px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.1);
  color: #e2e8f0;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  text-align: center;
  line-height: 1.4;
  transition: background 0.15s, border-color 0.15s, transform 0.1s;
}
.choice-btn:hover:not(:disabled) {
  background: rgba(249,115,22,0.12);
  border-color: rgba(249,115,22,0.4);
  color: #fff;
  transform: translateY(-1px);
}
.choice-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ── Résultat ── */
.result-card {
  border-radius: 16px;
  padding: 28px 24px;
  text-align: center;
  margin-bottom: 28px;
  border: 1px solid;
}
.result-found {
  background: rgba(34,197,94,0.07);
  border-color: rgba(34,197,94,0.25);
}
.result-missed {
  background: rgba(239,68,68,0.07);
  border-color: rgba(239,68,68,0.2);
}
.result-icon { font-size: 2.5rem; margin-bottom: 10px; }
.result-card h2 { font-size: 1.3rem; font-weight: 800; color: #f1f5f9; margin: 0 0 12px; }
.result-card p { color: #94a3b8; font-size: 0.9rem; margin: 0 0 16px; }

.result-reveal {
  display: flex; flex-direction: column; gap: 4px; align-items: center;
  margin-top: 16px;
}
.reveal-tag {
  font-size: 0.7rem; font-weight: 700; text-transform: uppercase;
  background: rgba(59,130,246,0.15); color: #93c5fd;
  padding: 2px 8px; border-radius: 99px;
}
.reveal-anime { font-size: 1.15rem; font-weight: 800; color: #f9a8d4; }
.reveal-track { font-size: 0.85rem; color: #64748b; }
.reveal-video {
  width: 100%; max-width: 520px; border-radius: 10px;
  margin-top: 16px; display: block; margin-inline: auto;
}

/* ── Classement ── */
.leaderboard-section { margin-top: 8px; }
.leaderboard-section h3 {
  font-size: 0.85rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; color: #475569; margin-bottom: 14px;
}
.lb-empty { color: #475569; font-style: italic; font-size: 0.88rem; }
.lb-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 6px; }
.lb-list li {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 10px; font-size: 0.88rem; color: #cbd5e1;
}
.lb-list li.me { border-color: rgba(249,115,22,0.4); background: rgba(249,115,22,0.06); }
.lb-rank { width: 28px; flex-shrink: 0; font-size: 1rem; }
.lb-name { flex: 1; font-weight: 600; color: #f1f5f9; }
.lb-time { color: #4ade80; font-weight: 700; font-variant-numeric: tabular-nums; }
.lb-miss { color: #475569; }

@media (max-width: 500px) {
  .daily-page { padding: 20px 14px 40px; }
  .choices-grid { grid-template-columns: 1fr; }
  .daily-title-row { flex-direction: column; align-items: flex-start; }
}
</style>
