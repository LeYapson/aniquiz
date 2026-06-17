import { test, expect } from '@playwright/test'

// ─── Messages WebSocket simulés ───────────────────────────────────────────────

const LOBBY_STATE = JSON.stringify({ type: 'GAME_STATE', payload: 'LOBBY' })
const PLAYER_LIST = JSON.stringify({
  type: 'PLAYER_LIST',
  payload: { players: [{ id: '1', username: 'Alice', score: 0 }], spectator_count: 0 },
})
const PLAYING_STATE = JSON.stringify({ type: 'GAME_STATE', payload: 'PLAYING' })
const NEW_QUESTION = JSON.stringify({
  type: 'NewQuestion',
  payload: { audio_url: 'https://example.com/track.webm', duration: 30 },
})

// ─── Helpers ──────────────────────────────────────────────────────────────────

async function mockAllApis(page, username = 'Alice', roomId = 'general') {
  await page.route('**/api/auth/login', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'test-jwt-token',
        user: { id: 1, username, level: 1, xp: 0, email: `${username.toLowerCase()}@test.com` },
      }),
    })
  )

  await page.route('**/rooms', (route) => {
    if (route.request().method() === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: roomId, state: 'LOBBY', players_count: 0, max_rounds: 5, is_private: false },
        ]),
      })
    }
    return route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({ room_id: roomId, creator_id: username }),
    })
  })

  await page.route('**/animes', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  )
}

async function loginAs(page, username = 'Alice') {
  // Passer la landing page si elle est affichée
  const playBtn = page.getByRole('button', { name: 'Jouer maintenant' })
  if (await playBtn.isVisible()) {
    await playBtn.click()
  }
  await page.getByPlaceholder('Votre pseudo').fill(username)
  await page.locator('input[type="password"]').fill('password123')
  await page.getByRole('button', { name: 'Se connecter' }).click()
  await expect(page.getByText('Salons Disponibles')).toBeVisible()
}

async function joinRoom(page, username = 'Alice', roomId = 'general') {
  await mockAllApis(page, username, roomId)

  await page.routeWebSocket(/\/ws/, (ws) => {
    ws.send(LOBBY_STATE)
    ws.send(PLAYER_LIST)
    ws.onMessage((msg) => {
      const data = JSON.parse(msg)
      if (data.type === 'START_GAME') {
        ws.send(PLAYING_STATE)
        ws.send(NEW_QUESTION)
      }
    })
  })

  await page.goto('/')
  await loginAs(page, username)
  await page.getByRole('button', { name: 'Rejoindre' }).first().click()
}

// ─── Formulaire de connexion ──────────────────────────────────────────────────

test.describe('Formulaire de connexion', () => {
  test('affiche la landing page au chargement', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('button', { name: 'Jouer maintenant' })).toBeVisible()
    await expect(page.getByRole('button', { name: '🏆 Classement' })).toBeVisible()
  })

  test('affiche le formulaire après "Jouer maintenant"', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'Jouer maintenant' }).click()

    await expect(page.getByRole('heading', { name: 'Connexion à AniQuiz' })).toBeVisible()
    await expect(page.getByPlaceholder('Votre pseudo')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Se connecter' })).toBeVisible()
  })

  test('affiche un message d\'erreur pour des identifiants incorrects', async ({ page }) => {
    await page.route('**/api/auth/login', (route) =>
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Identifiants incorrects' }),
      })
    )

    await page.goto('/')
    await page.getByRole('button', { name: 'Jouer maintenant' }).click()
    await page.getByPlaceholder('Votre pseudo').fill('inconnu')
    await page.locator('input[type="password"]').fill('wrongpassword')
    await page.getByRole('button', { name: 'Se connecter' }).click()

    await expect(page.getByText('Identifiants incorrects')).toBeVisible()
  })

  test('affiche la sélection de salon après la connexion', async ({ page }) => {
    await mockAllApis(page)
    await page.goto('/')
    await loginAs(page)

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
    await expect(page.getByRole('button', { name: 'Envoyer ma réponse' })).toBeVisible()
  })
})
