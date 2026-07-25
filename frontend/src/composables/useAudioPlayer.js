import { ref, computed, watch, nextTick } from 'vue'
import { audioOnlyUrl, isAudioOnly } from '../media'

export function useAudioPlayer({ currentAudioUrl, currentAnswerInfo, roundStartFraction }) {
  const audioEl = ref(null)
  const videoEl = ref(null)
  const audioFailed = ref(false)
  // Autoplay bloqué par le navigateur (Firefox/Safari) faute de geste utilisateur.
  const audioBlocked = ref(false)

  const AUDIO_VOLUME_KEY = 'aniquiz_volume'
  const clampVolume = (v) => Math.min(1, Math.max(0, v))
  const storedVolume = parseFloat(localStorage.getItem(AUDIO_VOLUME_KEY))
  const volume = ref(Number.isFinite(storedVolume) ? clampVolume(storedVolume) : 0.7)
  const lastNonZeroVolume = ref(volume.value > 0 ? volume.value : 0.7)
  const volumeIcon = computed(() => (volume.value === 0 ? '🔇' : volume.value < 0.5 ? '🔉' : '🔊'))

  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)

  const formatMediaTime = (s) => {
    if (!isFinite(s) || s < 0) s = 0
    const m = Math.floor(s / 60)
    const sec = Math.floor(s % 60)
    return `${m}:${sec.toString().padStart(2, '0')}`
  }

  // Source réellement lue par le <audio> : on tente d'abord l'audio-only (.ogg),
  // avec repli automatique sur la vidéo WebM si le .ogg n'existe pas.
  const playbackSrc = ref('')
  const triedVideoFallback = ref(false)

  // Libère explicitement la ressource d'un élément média avant qu'il soit démonté
  // (v-if). Sans ça, le navigateur garde la connexion réseau ouverte et on sature
  // la limite de 6 connexions HTTP/1.1 par hôte au bout de quelques manches.
  const releaseMedia = (el) => {
    if (!el) return
    try {
      el.pause()
      el.removeAttribute('src')
      el.load()
    } catch { /* élément déjà détaché */ }
  }

  // Démarre la lecture à la fraction choisie par le serveur (identique pour tous).
  // Clampé pour laisser ≥10 s de musique après le point de départ.
  const seekToStart = () => {
    const el = audioEl.value
    if (!el || !el.duration || !isFinite(el.duration)) return
    let start = (roundStartFraction.value || 0) * el.duration
    if (el.duration - start < 10) start = Math.max(0, el.duration - 10)
    try { el.currentTime = start } catch { /* seek non supporté */ }
  }

  const onAudioLoaded = () => {
    seekToStart()
    if (audioEl.value) {
      audioEl.value.volume = volume.value
      duration.value = audioEl.value.duration || 0
    }
    const p = audioEl.value?.play()
    if (p) {
      p.then(() => { audioBlocked.value = false })
       .catch(() => { audioBlocked.value = true })
    }
  }

  const resumeAudio = () => {
    audioEl.value?.play()
      .then(() => { audioBlocked.value = false })
      .catch(() => {})
  }

  const togglePlay = () => {
    const el = audioEl.value
    if (!el) return
    if (el.paused) {
      el.play().then(() => { audioBlocked.value = false }).catch(() => { audioBlocked.value = true })
    } else {
      el.pause()
    }
  }

  const onPlay = () => { isPlaying.value = true }
  const onPause = () => { isPlaying.value = false }
  const onTimeUpdate = () => { if (audioEl.value) currentTime.value = audioEl.value.currentTime || 0 }
  const onDurationChange = () => { if (audioEl.value) duration.value = audioEl.value.duration || 0 }
  const onSeek = (e) => {
    const el = audioEl.value
    if (!el) return
    const t = Number(e.target.value)
    try { el.currentTime = t; currentTime.value = t } catch { /* seek non supporté */ }
  }

  // Applique et persiste le volume ; répercuté sur l'audio et la vidéo du reveal.
  watch(volume, (v) => {
    v = clampVolume(v)
    if (audioEl.value) audioEl.value.volume = v
    if (videoEl.value) videoEl.value.volume = v
    if (v > 0) lastNonZeroVolume.value = v
    localStorage.setItem(AUDIO_VOLUME_KEY, String(v))
  })

  const toggleMute = () => {
    volume.value = volume.value > 0 ? 0 : lastNonZeroVolume.value
  }

  // Quand l'URL change : relâche l'ancien lecteur, charge la nouvelle source
  // audio-only. Le watcher s'exécute en flush "pre" → audioEl pointe encore
  // sur l'élément courant au moment du releaseMedia.
  watch(currentAudioUrl, async (url) => {
    if (!url) {
      releaseMedia(audioEl.value)
      playbackSrc.value = ''
      return
    }
    audioFailed.value = false
    audioBlocked.value = false
    triedVideoFallback.value = false
    currentTime.value = 0
    duration.value = 0
    playbackSrc.value = audioOnlyUrl(url)
    await nextTick()
    if (!audioEl.value) return
    audioEl.value.load() // déclenche @loadedmetadata → seekToStart + play
  })

  // Repli ogg → webm si l'audio-only n'existe pas pour cette piste.
  const onAudioError = () => {
    if (!currentAudioUrl.value) return
    if (!triedVideoFallback.value && isAudioOnly(playbackSrc.value)) {
      triedVideoFallback.value = true
      playbackSrc.value = currentAudioUrl.value
      nextTick(() => audioEl.value?.load())
      return
    }
    audioFailed.value = true
  }

  const retryAudio = async () => {
    audioFailed.value = false
    triedVideoFallback.value = false
    playbackSrc.value = audioOnlyUrl(currentAudioUrl.value)
    await nextTick()
    if (!audioEl.value) return
    audioEl.value.load()
  }

  // Lance la vidéo du reveal dès qu'elle est disponible, avec le même volume.
  watch(() => currentAnswerInfo.value.videoUrl, async (url) => {
    if (!url) return
    await nextTick()
    if (videoEl.value) videoEl.value.volume = volume.value
    videoEl.value?.play().catch(() => {})
  })

  return {
    audioEl, videoEl, volume, volumeIcon, isPlaying, currentTime, duration,
    formatMediaTime, playbackSrc, audioFailed, audioBlocked,
    togglePlay, toggleMute, onSeek, onAudioLoaded, onAudioError,
    onPlay, onPause, onTimeUpdate, onDurationChange,
    resumeAudio, retryAudio, releaseMedia,
  }
}
