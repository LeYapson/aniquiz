import { describe, it, expect, vi, beforeEach } from 'vitest'
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import App from '../../App.vue'

// Empêche l'appel réseau réel à /animes
global.fetch = vi.fn(() =>
  Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
)

// Simule un utilisateur connecté pour que le jeu soit visible
vi.mock('../../authStore', () => ({
  authStore: {
    user: { id: 1, username: 'TestPlayer', level: 1 },
    token: 'test-token',
    isAuthenticated: true,
    authHeaders: () => ({
      'Content-Type': 'application/json',
      Authorization: 'Bearer test-token',
    }),
    logout: vi.fn(),
  },
}))

describe('App Quiz Logic', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
    )
  })

  it('affiche le bouton de lancement quand on est en LOBBY', async () => {
    const wrapper = mount(App, {
      global: {
        stubs: { RoomSelection: true, GameTimer: true, AuthForm: true, LandingPage: true, LeaderboardPage: true, ProfilePage: true, ChatPanel: true },
      },
    })

    wrapper.vm.isConnected = true
    wrapper.vm.state = 'LOBBY'
    await nextTick()

    expect(wrapper.text()).toContain('Lancer la partie')
  })

  it('affiche le message de lecture quand on est en PLAYING', async () => {
    const wrapper = mount(App, {
      global: {
        stubs: { RoomSelection: true, GameTimer: true, AuthForm: true, LandingPage: true, LeaderboardPage: true, ProfilePage: true, ChatPanel: true },
      },
    })

    wrapper.vm.isConnected = true
    wrapper.vm.state = 'PLAYING'
    await nextTick()

    expect(wrapper.text()).toContain('Écoutez attentivement')
  })
})
