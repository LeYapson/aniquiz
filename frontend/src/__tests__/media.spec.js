import { describe, it, expect } from 'vitest'
import { audioOnlyUrl, isAudioOnly } from '../media.js'

describe('audioOnlyUrl', () => {
  it('convertit une vidéo WebM animethemes en audio-only Ogg sur l\'hôte a.', () => {
    expect(audioOnlyUrl('https://v.animethemes.moe/CowboyBebop-OP1.webm'))
      .toBe('https://a.animethemes.moe/CowboyBebop-OP1.ogg')
  })

  it('préserve une éventuelle query/fragment', () => {
    expect(audioOnlyUrl('https://v.animethemes.moe/X-OP1.webm?t=5'))
      .toBe('https://a.animethemes.moe/X-OP1.ogg?t=5')
  })

  it('laisse les URLs d\'un autre hôte intactes (repli sûr)', () => {
    const other = 'https://cdn.example.com/clip.webm'
    expect(audioOnlyUrl(other)).toBe(other)
  })

  it('gère les valeurs vides/non-string sans planter', () => {
    expect(audioOnlyUrl('')).toBe('')
    expect(audioOnlyUrl(null)).toBe(null)
    expect(audioOnlyUrl(undefined)).toBe(undefined)
  })
})

describe('isAudioOnly', () => {
  it('détecte l\'hôte audio-only', () => {
    expect(isAudioOnly('https://a.animethemes.moe/X-OP1.ogg')).toBe(true)
  })

  it('rejette la vidéo et les valeurs non pertinentes', () => {
    expect(isAudioOnly('https://v.animethemes.moe/X-OP1.webm')).toBe(false)
    expect(isAudioOnly('')).toBe(false)
    expect(isAudioOnly(null)).toBe(false)
  })
})
