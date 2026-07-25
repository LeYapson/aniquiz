import { ref } from 'vue'
import { WS_URL } from '../config'

export function useGameSocket({
  room, isConnected, players, state, currentAudioUrl, roundDuration, roundStartFraction,
  isRevealing, currentAnswerInfo, finalScores, roundHistory, skipVotes, hasVotedSkip,
  revealSkipVotes, hasVotedRevealSkip, hasBuzzed, buzzedUsers, chatMessages,
  isSpectator, spectatorCount, reconnectMsg, isCreator, roomSettings,
  mobileTab, reactionOverlay, audioEl, videoEl, releaseMedia,
  authStore, toast, loadAnimeDictionary,
}) {
  // socket exposé en ref pour que le template auto-unwrappe vers l'instance WebSocket.
  const socket = ref(null)
  let reconnectAttempts = 0
  let intentionalClose = false

  const send = (type, payload = null) => {
    if (socket.value?.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify({ type, payload }))
    }
  }

  const startGame = () => send('START_GAME')

  const backToLobby = () => {
    finalScores.value = []
    roundHistory.value = []
    skipVotes.value = { votes: 0, needed: 1 }
    hasVotedSkip.value = false
    revealSkipVotes.value = { votes: 0, needed: 1 }
    hasVotedRevealSkip.value = false
    state.value = 'LOBBY'
  }

  const sendSkipVote = () => {
    if (!hasVotedSkip.value) { send('VOTE_SKIP'); hasVotedSkip.value = true }
  }

  const forceSkip = () => send('FORCE_SKIP')

  const sendRevealSkipVote = () => {
    if (!hasVotedRevealSkip.value) { send('VOTE_SKIP_REVEAL'); hasVotedRevealSkip.value = true }
  }

  const kickPlayer = (username) => send('KICK_PLAYER', username)

  const onBuzz = () => {
    if (hasBuzzed.value) return
    send('BUZZ')
    hasBuzzed.value = true
    if (audioEl.value) audioEl.value.muted = true
  }

  // sendAnswer reçoit la valeur directement — App.vue garde le v-model userGuess.
  const sendAnswer = (guess) => send('SUBMIT_ANSWER', guess)

  const sendReaction = (emoji) => send('REACTION', emoji)

  const sendChat = (text) => send('CHAT', text)

  const disconnect = () => {
    intentionalClose = true
    chatMessages.value = []
    isSpectator.value = false
    spectatorCount.value = 0
    mobileTab.value = 'game'
    socket.value?.close()
  }

  const connectWebSocket = (room_id, password) => {
    const wsUrl = `${WS_URL}/ws?room=${room_id}&password=${password || ''}&token=${authStore.token}`
    socket.value = new WebSocket(wsUrl)

    socket.value.onopen = () => {
      isConnected.value = true
      reconnectAttempts = 0
      reconnectMsg.value = ''
    }

    socket.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        switch (data.type) {
          case 'PLAYER_LIST':
            players.value = data.payload.players ?? []
            spectatorCount.value = data.payload.spectator_count ?? 0
            break
          case 'SPECTATOR_STATUS':
            isSpectator.value = data.payload
            break
          case 'GAME_STATE':
            state.value = data.payload
            break
          case 'NewQuestion':
            // Relâche la vidéo du reveal précédent avant que v-if ne la démonte.
            releaseMedia(videoEl.value)
            isRevealing.value = false
            currentAudioUrl.value = data.payload.audio_url
            roundDuration.value = data.payload.duration
            roundStartFraction.value = data.payload.start_fraction ?? 0
            hasVotedSkip.value = false
            skipVotes.value = { votes: 0, needed: 1 }
            hasVotedRevealSkip.value = false
            revealSkipVotes.value = { votes: 0, needed: 1 }
            hasBuzzed.value = false
            buzzedUsers.value = []
            if (audioEl.value) audioEl.value.muted = false
            break
          case 'ROUND_ENDED':
            isRevealing.value = true
            currentAnswerInfo.value = {
              animeName: data.payload.answer,
              title: data.payload.title || '',
              artist: data.payload.artist || '',
              videoUrl: data.payload.video_url || '',
              trackType: data.payload.track_type || '',
              difficulty: data.payload.difficulty || 0,
              foundBy: data.payload.found_by || [],
            }
            currentAudioUrl.value = ''
            break
          case 'SETTINGS_UPDATED':
            roomSettings.value = {
              maxRounds: data.payload.max_rounds,
              roundDuration: data.payload.round_duration,
              filterType: data.payload.filter_type,
              isPrivate: data.payload.is_private,
              buzzerMode: data.payload.buzzer_mode === true,
              guessMode: data.payload.guess_mode || 'anime',
            }
            break
          case 'PLAYER_BUZZED':
            if (data.payload.username && !buzzedUsers.value.includes(data.payload.username)) {
              buzzedUsers.value.push(data.payload.username)
            }
            break
          case 'PLAYER_WRONG':
            buzzedUsers.value = buzzedUsers.value.filter((u) => u !== data.payload.username)
            if (data.payload.username === authStore.user?.username) {
              toast.error('Mauvaise réponse — éliminé pour ce round !', { title: '🔔 Buzzer' })
            }
            break
          case 'NOTICE':
            toast.info(data.payload, { title: 'Info partie' })
            break
          case 'SKIP_VOTE_UPDATE':
            skipVotes.value = { votes: data.payload.votes, needed: data.payload.needed }
            break
          case 'REVEAL_SKIP_VOTE_UPDATE':
            revealSkipVotes.value = { votes: data.payload.votes, needed: data.payload.needed }
            break
          case 'HOST_CHANGED':
            isCreator.value = data.payload === authStore.user?.username
            break
          case 'KICKED':
            disconnect()
            toast.error(data.payload ?? 'Vous avez été expulsé de la partie.', { title: 'Expulsé' })
            break
          case 'GAME_OVER':
            finalScores.value = [...players.value].sort((a, b) => b.score - a.score)
            roundHistory.value = data.payload.history ?? []
            state.value = 'GAME_OVER'
            break
          case 'CHAT_MESSAGE':
            chatMessages.value.push({ username: data.payload.username, message: data.payload.message })
            if (chatMessages.value.length > 200) chatMessages.value.shift()
            break
          case 'REACTION_BROADCAST':
            reactionOverlay.value?.addParticle(data.payload.emoji)
            break
          case 'XP_GAINED': {
            const oldLevel = authStore.user?.level ?? 1
            const levelUp = data.payload.new_level > oldLevel
            toast.xp({
              xpGained: data.payload.xp_gained,
              newXP: data.payload.new_xp,
              newLevel: data.payload.new_level,
              levelUp,
            })
            if (authStore.user) {
              authStore.setUser(
                { ...authStore.user, xp: data.payload.new_xp, level: data.payload.new_level },
                authStore.token
              )
            }
            break
          }
        }
      } catch (err) {
        console.error('Erreur message:', err)
      }
    }

    socket.value.onclose = () => {
      if (intentionalClose) {
        isConnected.value = false
        players.value = []
        state.value = 'LOBBY'
        isRevealing.value = false
        reconnectMsg.value = ''
        return
      }
      // Déconnexion involontaire : backoff exponentiel (1s, 2s, 4s… max 30s).
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
      reconnectAttempts++
      reconnectMsg.value = `Connexion perdue. Reconnexion dans ${Math.round(delay / 1000)}s… (tentative ${reconnectAttempts})`
      setTimeout(() => connectWebSocket(room.value, ''), delay)
    }
  }

  const setupWebSocket = ({ room_id, password, isCreator: creator }) => {
    room.value = room_id
    isCreator.value = !!creator
    intentionalClose = false
    reconnectAttempts = 0
    loadAnimeDictionary()
    connectWebSocket(room_id, password)
  }

  return {
    socket, setupWebSocket, disconnect,
    startGame, backToLobby, sendSkipVote, forceSkip, sendRevealSkipVote,
    kickPlayer, onBuzz, sendAnswer, sendReaction, sendChat,
  }
}
