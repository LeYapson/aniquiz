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
const ROUND_ENDED = JSON.stringify({
  type: 'ROUND_ENDED',
  payload: {
    answer: 'Naruto',
    title: 'Rocks',
    artist: 'Hound Dog',
    video_url: 'https://example.com/video.webm',
    track_type: 'OP',
    difficulty: 75,
    found_by: [],
  },
})
const SPECTATOR_ON = JSON.stringify({ type: 'SPECTATOR_STATUS', payload: true })
const PLAYER_LIST_PLAYING = JSON.stringify({
  type: 'PLAYER_LIST',
  payload: { players: [{ id: '2', username: 'Bob', score: 10 }], spectator_count: 1 },
})

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Remplace les méthodes media du navigateur par des no-ops.
 * Doit être appelé AVANT page.goto() pour s'exécuter avant le JS de la page.
 * Sans ça, load() + play() déclenchent de vraies requêtes réseau vers des URL
 * fictives (https://example.com/...) et peuvent faire échouer les tests en CI.
 */
async function mockMedia(page) {
  await page.addInitScript(() => {
    HTMLMediaElement.prototype.load  = function () {}
    HTMLMediaElement.prototype.play  = function () { return Promise.resolve() }
    HTMLMediaElement.prototype.pause = function () {}
  })
}

async function mockAllApis(page, username = 'Alice', roomId = 'general') {
  // Empêche les vraies requêtes réseau pour les médias
  await mockMedia(page)

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
  await expect(page.getByRole('main').getByRole('button', { name: 'Jouer', exact: true })).toBeVisible()
}

// Ouvre la modale de modes et choisit Multi pour atteindre la liste des salons.
async function openRoomList(page) {
  await page.getByRole('main').getByRole('button', { name: 'Jouer', exact: true }).click()
  await page.locator('.mode-card-btn', { hasText: 'Multi' }).click()
  await expect(page.getByText('Salons disponibles')).toBeVisible()
}

async function joinRoomAsSpectator(page, username = 'Alice') {
  // Empêche les vraies requêtes réseau pour les médias
  await mockMedia(page)

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
    if (route.request().method() !== 'GET') return route.continue()
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        { id: 'general', state: 'PLAYING', players_count: 2, max_rounds: 5, is_private: false },
      ]),
    })
  })
  await page.route('**/animes', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  )
  await page.routeWebSocket(/\/ws/, (ws) => {
    ws.send(PLAYING_STATE)
    ws.send(SPECTATOR_ON)
    ws.send(PLAYER_LIST_PLAYING)
    ws.send(NEW_QUESTION)
  })
  await page.goto('/')
  const playBtn = page.getByRole('button', { name: 'Jouer maintenant' })
  if (await playBtn.isVisible()) await playBtn.click()
  await page.getByPlaceholder('Votre pseudo').fill(username)
  await page.locator('input[type="password"]').fill('password123')
  await page.getByRole('button', { name: 'Se connecter' }).click()
  await expect(page.getByRole('main').getByRole('button', { name: 'Jouer', exact: true })).toBeVisible()
  await openRoomList(page)
  await page.getByRole('button', { name: /Regarder/ }).first().click()
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
  await openRoomList(page)
  await page.getByRole('button', { name: 'Rejoindre' }).first().click()
}

// ─── Formulaire de connexion ──────────────────────────────────────────────────

test.describe('Formulaire de connexion', () => {
  test('affiche la landing page au chargement', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('button', { name: 'Jouer maintenant' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Classement' }).first()).toBeVisible()
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
    await openRoomList(page)

    await expect(page.getByText('Salons disponibles')).toBeVisible()
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

    await expect(page.getByRole('main').getByRole('button', { name: 'Jouer', exact: true })).toBeVisible()
  })
})

// ─── Déroulement d'une partie ─────────────────────────────────────────────────

test.describe('Partie', () => {
  test('passe en état PLAYING après "Lancer la partie"', async ({ page }) => {
    await joinRoom(page)

    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.getByText(/Écoutez et devinez/)).toBeVisible()
  })

  test('affiche le lecteur audio après le démarrage', async ({ page }) => {
    await joinRoom(page)
    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.locator('audio')).toBeAttached()
  })

  test('lance automatiquement la lecture audio à la réception d\'une question', async ({ page }) => {
    // Espionne play() pour vérifier qu'il est bien appelé par le watcher
    await mockAllApis(page)
    await page.addInitScript(() => {
      window.__playCallCount = 0
      HTMLMediaElement.prototype.load  = function () {}
      HTMLMediaElement.prototype.play  = function () {
        window.__playCallCount++
        return Promise.resolve()
      }
      HTMLMediaElement.prototype.pause = function () {}
    })

    let wsServer
    await page.routeWebSocket(/\/ws/, (ws) => {
      wsServer = ws
      ws.send(LOBBY_STATE)
      ws.send(PLAYER_LIST)
    })

    await page.goto('/')
    await loginAs(page)
    await openRoomList(page)
    await page.getByRole('button', { name: 'Rejoindre' }).first().click()
    await expect(page.locator('.sidebar').getByText('Alice')).toBeVisible()

    wsServer.send(PLAYING_STATE)
    wsServer.send(NEW_QUESTION)

    // L'élément audio doit apparaître
    await expect(page.locator('audio')).toBeAttached()

    // play() doit avoir été appelé automatiquement par le watcher
    const playCallCount = await page.evaluate(() => window.__playCallCount)
    expect(playCallCount).toBeGreaterThanOrEqual(1)
  })

  test('affiche le champ de réponse pendant la partie', async ({ page }) => {
    await joinRoom(page)
    await page.getByRole('button', { name: 'Lancer la partie' }).click()

    await expect(page.getByPlaceholder(/nom de l'anime/i)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Envoyer ma réponse' })).toBeVisible()
  })

  test('affiche la vidéo de révélation en fin de round', async ({ page }) => {
    await mockAllApis(page)

    let wsServer
    await page.routeWebSocket(/\/ws/, (ws) => {
      wsServer = ws
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
    await loginAs(page)
    await openRoomList(page)
    await page.getByRole('button', { name: 'Rejoindre' }).first().click()
    await page.getByRole('button', { name: 'Lancer la partie' }).click()
    await expect(page.locator('audio')).toBeAttached()

    wsServer.send(ROUND_ENDED)

    await expect(page.locator('video')).toBeAttached()
    await expect(page.getByText('Naruto')).toBeVisible()
  })
})

// ─── Chat ────────────────────────────────────────────────────────────────────

test.describe('Chat', () => {
  test('affiche le panneau de chat en jeu', async ({ page }) => {
    await joinRoom(page)
    await expect(page.locator('.chat-panel')).toBeVisible()
  })

  test('affiche un message reçu dans le chat', async ({ page }) => {
    await mockAllApis(page)

    let wsServer
    await page.routeWebSocket(/\/ws/, (ws) => {
      wsServer = ws
      ws.send(LOBBY_STATE)
      ws.send(PLAYER_LIST)
    })

    await page.goto('/')
    await loginAs(page)
    await openRoomList(page)
    await page.getByRole('button', { name: 'Rejoindre' }).first().click()
    await expect(page.locator('.sidebar').getByText('Alice')).toBeVisible()

    wsServer.send(JSON.stringify({
      type: 'CHAT_MESSAGE',
      payload: { username: 'Bob', message: 'Bonjour tout le monde !' },
    }))

    await expect(page.getByText('Bonjour tout le monde !')).toBeVisible()
  })

  test('affiche le nom de l\'expéditeur dans le chat', async ({ page }) => {
    await mockAllApis(page)

    let wsServer
    await page.routeWebSocket(/\/ws/, (ws) => {
      wsServer = ws
      ws.send(LOBBY_STATE)
      ws.send(PLAYER_LIST)
    })

    await page.goto('/')
    await loginAs(page)
    await openRoomList(page)
    await page.getByRole('button', { name: 'Rejoindre' }).first().click()
    await expect(page.locator('.sidebar').getByText('Alice')).toBeVisible()

    wsServer.send(JSON.stringify({
      type: 'CHAT_MESSAGE',
      payload: { username: 'Bob', message: 'Salut !' },
    }))

    await expect(page.locator('.chat-username').getByText('Bob')).toBeVisible()
  })
})

// ─── Mode spectateur ─────────────────────────────────────────────────────────

test.describe('Spectateur', () => {
  test('affiche le badge spectateur en rejoignant une partie en cours', async ({ page }) => {
    await joinRoomAsSpectator(page)
    await expect(page.locator('.spectator-badge')).toBeVisible()
  })

  test('n\'affiche pas le champ de réponse en mode spectateur', async ({ page }) => {
    await joinRoomAsSpectator(page)
    await expect(page.getByPlaceholder(/nom de l'anime/i)).not.toBeVisible()
  })

  test('affiche le message de spectateur pendant la partie', async ({ page }) => {
    await joinRoomAsSpectator(page)
    await expect(page.locator('.spectator-watching')).toBeVisible()
  })
})

// ─── Classement ──────────────────────────────────────────────────────────────

test.describe('Classement', () => {
  test('affiche le classement depuis la landing page', async ({ page }) => {
    await page.route('**/api/leaderboard', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { rank: 1, user_id: 2, username: 'TopPlayer', level: 10, xp: 5000, total_games: 50, best_score: 300 },
          { rank: 2, user_id: 1, username: 'Alice', level: 5, xp: 1200, total_games: 20, best_score: 150 },
        ]),
      })
    )

    await page.goto('/')
    await page.getByRole('button', { name: 'Classement' }).first().click()

    await expect(page.getByRole('heading', { name: 'Classement Global' })).toBeVisible()
    await expect(page.getByText('TopPlayer')).toBeVisible()
  })

  test('affiche le bouton retour depuis le classement public', async ({ page }) => {
    await page.route('**/api/leaderboard', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
    )

    await page.goto('/')
    await page.getByRole('button', { name: 'Classement' }).first().click()
    await page.getByRole('button', { name: /Retour/ }).click()

    await expect(page.getByRole('button', { name: 'Jouer maintenant' })).toBeVisible()
  })
})

// ─── Page 404 ────────────────────────────────────────────────────────────────

test.describe('Page 404', () => {
  test('affiche la page 404 pour une URL inconnue', async ({ page }) => {
    await page.goto('/cette-page-nexiste-pas')

    await expect(page.getByText('404')).toBeVisible()
    await expect(page.getByRole('link', { name: "Retour à l'accueil" })).toBeVisible()
  })

  test("redirige vers l'accueil depuis la page 404", async ({ page }) => {
    await page.goto('/cette-page-nexiste-pas')
    await page.getByRole('link', { name: "Retour à l'accueil" }).click()

    await expect(page.getByRole('button', { name: 'Jouer maintenant' })).toBeVisible()
  })
})
