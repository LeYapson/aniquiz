import { describe, it, expect } from 'vitest'
import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import App from '../../App.vue'

describe('App Quiz Logic', () => {
  it('affiche le bouton de lancement quand on est en LOBBY', async () => {
    const wrapper = mount(App, {
      global: { stubs: { JoinRoom: true } },
    })

    wrapper.vm.user = 'TestPlayer'
    wrapper.vm.isConnected = true
    wrapper.vm.state = 'LOBBY'
    await nextTick()

    expect(wrapper.text()).toContain('Lancer la partie')
  })

  it('affiche le message de lecture quand on est en PLAYING', async () => {
    const wrapper = mount(App, {
      global: { stubs: { JoinRoom: true } },
    })

    wrapper.vm.user = 'TestPlayer'
    wrapper.vm.isConnected = true
    wrapper.vm.state = 'PLAYING'
    await nextTick()

    expect(wrapper.text()).toContain('Écoutez attentivement')
  })
})
