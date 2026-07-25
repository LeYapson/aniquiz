import { API_URL } from '../config'

export function useOAuth({ authStore }) {
  const connectAnilist = () => {
    window.location.href = `${API_URL}/api/auth/anilist?token=${authStore.token}`
  }

  const connectMAL = () => {
    window.location.href = `${API_URL}/api/auth/mal?token=${authStore.token}`
  }

  const connectDiscord = () => {
    window.location.href = `${API_URL}/api/auth/discord?token=${authStore.token}`
  }

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

    const discordStatus = params.get('discord')
    if (discordStatus === 'success') {
      const username = params.get('username')
      if (username && authStore.user) {
        authStore.setUser({ ...authStore.user, discord_username: username }, authStore.token)
      }
    }

    if (anilistStatus || malStatus || discordStatus) {
      window.history.replaceState({}, '', window.location.pathname)
    }
  }

  return { connectAnilist, connectMAL, connectDiscord, checkOAuthCallback }
}
