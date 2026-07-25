import { ref, computed } from 'vue'

export function useGameState() {
  const isConnected = ref(false)
  const room = ref('')
  const players = ref([])
  const state = ref('LOBBY')
  const currentAudioUrl = ref('')
  const userGuess = ref('')
  const roundDuration = ref(0)
  const roundStartFraction = ref(0)
  const animeDictionary = ref([])
  const isRevealing = ref(false)
  const currentAnswerInfo = ref({
    animeName: '',
    title: '',
    artist: '',
    videoUrl: '',
    trackType: '',
    difficulty: 0,
    foundBy: [],
  })
  const finalScores = ref([])
  const roundHistory = ref([])

  const speedStats = computed(() => {
    const map = {}
    for (const r of roundHistory.value) {
      for (const f of r.found_by || []) {
        const s = map[f.username] || (map[f.username] = { username: f.username, found: 0, totalMs: 0, bestMs: Infinity })
        s.found++
        s.totalMs += f.time_ms
        if (f.time_ms < s.bestMs) s.bestMs = f.time_ms
      }
    }
    return Object.values(map)
      .map((s) => ({ ...s, avgMs: Math.round(s.totalMs / s.found) }))
      .sort((a, b) => a.avgMs - b.avgMs)
  })

  const skipVotes = ref({ votes: 0, needed: 1 })
  const hasVotedSkip = ref(false)
  const revealSkipVotes = ref({ votes: 0, needed: 1 })
  const hasVotedRevealSkip = ref(false)
  const reconnectMsg = ref('')
  const isCreator = ref(false)
  const roomSettings = ref({
    maxRounds: 5, roundDuration: 20, filterType: '', decade: 0,
    isPrivate: false, password: '', buzzerMode: false, guessMode: 'anime',
  })
  const buzzerMode = computed(() => roomSettings.value.buzzerMode === true)
  const guessMode = computed(() => roomSettings.value.guessMode || 'anime')
  const guessLabel = computed(() => ({
    anime: "le nom de l'anime",
    title: 'le titre de la musique',
    artist: "l'artiste",
  }[guessMode.value] || "le nom de l'anime"))
  const hasBuzzed = ref(false)
  const buzzedUsers = ref([])
  const chatMessages = ref([])
  const isSpectator = ref(false)
  const spectatorCount = ref(0)
  const mobileTab = ref('game')

  return {
    isConnected, room, players, state, currentAudioUrl, userGuess,
    roundDuration, roundStartFraction, animeDictionary, isRevealing, currentAnswerInfo,
    finalScores, roundHistory, speedStats, skipVotes, hasVotedSkip,
    revealSkipVotes, hasVotedRevealSkip, reconnectMsg, isCreator, roomSettings,
    buzzerMode, guessMode, guessLabel, hasBuzzed, buzzedUsers, chatMessages,
    isSpectator, spectatorCount, mobileTab,
  }
}
