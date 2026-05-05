import { test, expect } from '@playwright/test'

// Messages que le serveur enverrait normalement après connexion
const LOBBY_STATE = JSON.stringify({ type: 'GAME_STATE', payload: 'LOBBY' })
const PLAYER_LIST = JSON.stringify({
  type: 'PLAYER_LIST',
  payload: [{ id: '1', username: 'Alice', score: 0 }],
})
const PLAYING_STATE = JSON.stringify({ type: 'GAME_STATE', payload: 'PLAYING' })
const NEW_QUESTION = JSON.stringify({
  type: 'NewQuestion',
  payload: { audio_url: 'https://example.com/track.webm', room_id: 'general' },
})

// Helper : mock WebSocket + rejoindre la room
async function joinRoom(page, username = 'Alice', roomId = 'general') {
  await page.routeWebSocket(/\/ws/, (ws) => {
    ws.send(LOBBY_STATE)
    ws.send(PLAYER_LIST)

    ws.onMessage((message) => {
      const data = JSON.parse(message)
      if (data.type === 'START_GAME') {
        ws.send(PLAYING_STATE)
        ws.send(NEW_QUESTION)
      }
    })
  })

  await page.goto('/')
  await page.getByPlaceholder('Ton pseudo').fill(username)
  await page.getByPlaceholder('Nom du salon (ex: general)').fill(roomId)
  await page.getByRole('button', { name: 'Rejoindre' }).click()
}

// ─── Formulaire de connexion ──────────────────────────────────────────────────

test.describe('Formulaire de connexion', () => {
  test('affiche le formulaire au chargement', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByPlaceholder('Ton pseudo')).toBeVisible()
    await expect(page.getByPlaceholder('Nom du salon (ex: general)')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Rejoindre' })).toBeVisible()
  })

  test("ne rejoint pas si les deux champs sont vides", async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'Rejoindre' }).click()

    await expect(page.getByPlaceholder('Ton pseudo')).toBeVisible()
  })

  test("ne rejoint pas si seul le pseudo est rempli", async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('Ton pseudo').fill('Alice')
    await page.getByRole('button', { name: 'Rejoindre' }).click()

    await expect(page.getByPlaceholder('Ton pseudo')).toBeVisible()
  })

  test("ne rejoint pas si seul le salon est rempli", async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('Nom du salon (ex: general)').fill('general')
    await page.getByRole('button', { name: 'Rejoindre' }).click()

    await expect(page.getByPlaceholder('Ton pseudo')).toBeVisible()
  })
})

// ─── Lobby ───────────────────────────────────────────────────────────────────

test.describe('Lobby', () => {
  test('affiche la vue de jeu après connexion', async ({ page }) => {
    await joinRoom(page)

    await expect(page.getByRole('button', { name: 'Lancer la partie' })).toBeVisible()
  })

  test('affiche le nom du salon et le pseudo dans la barre de statut', async ({ page }) => {
    await joinRoom(page)

    await expect(page.getByText('general')).toBeVisible()
    await expect(page.getByText('Alice')).toBeVisible()
  })

  test('affiche la liste des joueurs dans la sidebar', async ({ page }) => {
    await joinRoom(page)

    await expect(page.getByText('Joueurs')).toBeVisible()
    await expect(page.getByText('Alice')).toBeVisible()
  })

  test('affiche le bouton Quitter', async ({ page }) => {
    await joinRoom(page)

    await expect(page.getByRole('button', { name: 'Quitter' })).toBeVisible()
  })

  test('retourne au formulaire après avoir quitté', async ({ page }) => {
    await joinRoom(page)

    await page.getByRole('button', { name: 'Quitter' }).click()

    await expect(page.getByPlaceholder('Ton pseudo')).toBeVisible()
  })
})

// ─── Déroulement d'une partie ─────────────────────────────────────────────────

test.describe('Partie', () => {
  test('passe en état PLAYING après "Lancer la partie"', async ({ page }) => {
    await joinRoom(page)

    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.getByText('Écoutez attentivement')).toBeVisible()
  })

  test('affiche le lecteur audio après le démarrage', async ({ page }) => {
    await joinRoom(page)
    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.locator('audio')).toBeAttached()
  })

  test('affiche le champ de réponse pendant la partie', async ({ page }) => {
    await joinRoom(page)
    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.getByPlaceholder('Quel est cet anime ?')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Envoyer' })).toBeVisible()
  })
})
