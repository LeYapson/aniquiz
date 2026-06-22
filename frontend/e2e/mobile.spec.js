import { test, expect } from '@playwright/test'

// Vue mobile (iPhone 12-ish). On vérifie l'absence de débordement horizontal —
// le bug d'affichage mobile le plus courant.
test.use({ viewport: { width: 390, height: 844 } })

/** Débordement horizontal du document, en px (0 = parfait). */
async function horizontalOverflow(page) {
  return page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth
  )
}

test.describe('Mobile', () => {
  test('la landing page ne déborde pas horizontalement', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('button', { name: 'Jouer maintenant' })).toBeVisible()
    expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1)
  })

  test('le classement public ne déborde pas horizontalement', async ({ page }) => {
    await page.route('**/api/leaderboard', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { rank: 1, user_id: 2, username: 'TopPlayer', level: 10, xp: 5000, total_games: 50, best_score: 300 },
        ]),
      })
    )
    await page.goto('/')
    await page.getByRole('button', { name: 'Classement' }).first().click()
    await expect(page.getByRole('heading', { name: 'Classement Global' })).toBeVisible()
    expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1)
  })
})
