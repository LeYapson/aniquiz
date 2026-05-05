import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import App from '../../App.vue' // Ajuste le chemin selon ta structure

describe('App Quiz Logic', () => {
  it('affiche le bouton de lancement quand on est en LOBBY', () => {
    const wrapper = mount(App)
    // On force l'état interne pour le test
    wrapper.vm.state = 'LOBBY'
    
    expect(wrapper.text()).toContain('Lancer la partie')
  })

  it('affiche le message de lecture quand on est en PLAYING', async () => {
    const wrapper = mount(App)
    wrapper.vm.state = 'PLAYING'
    
    // On attend que Vue mette à jour le DOM
    await wrapper.nextTick()
    
    expect(wrapper.text()).toContain('Écoutez attentivement')
  })
})