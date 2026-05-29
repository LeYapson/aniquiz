import { test, expect } from '@playwright/test'

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

async function mockRoomsApi(page, roomId = 'general') {
  await page.route('**/rooms', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: roomId, state: 'LOBBY', players_count: 0, max_rounds: 5, is_private: false },
        ]),
      })
    } else {
      await route.continue()
    }
  })
}

// Helper : mock WebSocket + rejoindre la room
async function joinRoom(page, username = 'Alice', roomId = 'general') {
  await mockRoomsApi(page, roomId)

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
  await page.getByPlaceholder('Ex: Gon, Goku...').fill(username)
  await page.getByRole('button', { name: 'Continuer' }).click()
  await page.getByRole('button', { name: 'Rejoindre' }).first().click()
}

// ─── Formulaire de connexion ──────────────────────────────────────────────────

test.describe('Formulaire de connexion', () => {
  test('affiche le formulaire au chargement', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByPlaceholder('Ex: Gon, Goku...')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Continuer' })).toBeVisible()
  })

  test('le bouton Continuer est désactivé si le pseudo est vide', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('button', { name: 'Continuer' })).toBeDisabled()
  })

  test('affiche la sélection de salon après avoir entré un pseudo', async ({ page }) => {
    await mockRoomsApi(page)
    await page.goto('/')
    await page.getByPlaceholder('Ex: Gon, Goku...').fill('Alice')
    await page.getByRole('button', { name: 'Continuer' }).click()

    await expect(page.getByText('Salons Disponibles')).toBeVisible()
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

    await expect(page.locator('.status-bar').getByText('general')).toBeVisible()
    await expect(page.locator('.status-bar').getByText('Alice')).toBeVisible()
  })

  test('affiche la liste des joueurs dans la sidebar', async ({ page }) => {
    await joinRoom(page)

    await expect(page.locator('.sidebar').getByText('Joueurs')).toBeVisible()
    await expect(page.locator('.sidebar').getByText('Alice')).toBeVisible()
  })

  test('affiche le bouton Quitter', async ({ page }) => {
    await joinRoom(page)

    await expect(page.getByRole('button', { name: 'Quitter' })).toBeVisible()
  })

  test('retourne à la sélection de salon après avoir quitté', async ({ page }) => {
    await joinRoom(page)

    await page.getByRole('button', { name: 'Quitter' }).click()

    await expect(page.getByText('Salons Disponibles')).toBeVisible()
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

    await expect(page.getByPlaceholder("Nom de l'anime...")).toBeVisible()
    await expect(page.getByRole('button', { name: 'Envoyer' })).toBeVisible()
  })
})
