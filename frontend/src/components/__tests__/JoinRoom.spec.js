import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import JoinRoom from '../JoinRoom.vue'

describe('JoinRoom', () => {
  it('affiche le formulaire de connexion', () => {
    const wrapper = mount(JoinRoom)

    expect(wrapper.find('input[placeholder="Ton pseudo"]').exists()).toBe(true)
    expect(wrapper.find('input[placeholder="Nom du salon (ex: general)"]').exists()).toBe(true)
    expect(wrapper.find('button').text()).toBe('Rejoindre')
  })

  it("n'émet pas join si les champs sont vides", async () => {
    const wrapper = mount(JoinRoom)

    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('join')).toBeFalsy()
  })

  it("n'émet pas join si seul le pseudo est rempli", async () => {
    const wrapper = mount(JoinRoom)

    await wrapper.find('input[placeholder="Ton pseudo"]').setValue('Alice')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('join')).toBeFalsy()
  })

  it("n'émet pas join si seul le salon est rempli", async () => {
    const wrapper = mount(JoinRoom)

    await wrapper.find('input[placeholder="Nom du salon (ex: general)"]').setValue('general')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('join')).toBeFalsy()
  })

  it('émet join avec les bonnes données quand les deux champs sont remplis', async () => {
    const wrapper = mount(JoinRoom)

    await wrapper.find('input[placeholder="Ton pseudo"]').setValue('Alice')
    await wrapper.find('input[placeholder="Nom du salon (ex: general)"]').setValue('general')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('join')).toBeTruthy()
    expect(wrapper.emitted('join')[0][0]).toEqual({
      username: 'Alice',
      roomId: 'general',
    })
  })
})
