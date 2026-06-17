# AniQuiz

Blindtest anime multijoueur en temps réel. Les joueurs écoutent des openings/endings et doivent deviner l'anime le plus vite possible.

## Stack technique

| Couche | Technologie |
|--------|-------------|
| Backend | Go 1.26 + Gin + Gorilla WebSocket |
| Frontend | Vue 3.5 (Composition API) + Vite |
| Base de données | PostgreSQL 16 |
| Auth | JWT (HS256) + OAuth AniList + OAuth MAL (PKCE) |
| Sourcing audio | Jikan API (métadonnées) + scraping |
| CI/CD | GitHub Actions |
| Conteneurisation | Docker multi-stage + GHCR |

## Prérequis

- Go 1.26+
- Node.js 20+
- PostgreSQL 16 (ou Docker)
- Comptes OAuth AniList et MyAnimeList (optionnel, pour le filtrage par liste perso)

## Lancer le projet en local

### Avec Docker (recommandé)

```bash
cp .env.example .env
# Remplir les variables OAuth dans .env
docker compose up
```

- Frontend : http://localhost:5173
- Backend : http://localhost:8080

### Sans Docker

**Backend**
```bash
cp .env.example .env
# Adapter DATABASE_URL dans .env
go run ./cmd/api
```

**Frontend** (dans un second terminal)
```bash
cd frontend
npm install
npm run dev
```

## Variables d'environnement

Copier `.env.example` en `.env` et renseigner :

| Variable | Description | Obligatoire |
|----------|-------------|-------------|
| `DATABASE_URL` | URL de connexion PostgreSQL | Oui |
| `JWT_SECRET` | Clé de signature des tokens JWT | Oui |
| `FRONTEND_URL` | URL du frontend (CORS + redirections OAuth) | Oui |
| `PORT` | Port du serveur HTTP (défaut : `8080`) | Non |
| `ANILIST_CLIENT_ID` | ID client OAuth AniList | Non |
| `ANILIST_CLIENT_SECRET` | Secret client OAuth AniList | Non |
| `ANILIST_REDIRECT_URI` | URI de redirection AniList | Non |
| `MAL_CLIENT_ID` | ID client OAuth MyAnimeList | Non |
| `MAL_CLIENT_SECRET` | Secret client OAuth MyAnimeList | Non |
| `MAL_REDIRECT_URI` | URI de redirection MAL | Non |

> Les variables OAuth sont optionnelles pour jouer, mais nécessaires pour le filtrage par liste personnelle.

## Schéma de base de données

```sql
-- Exécuter une fois à la création
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    username        TEXT UNIQUE NOT NULL,
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    xp              INT DEFAULT 0,
    level           INT DEFAULT 1,
    anilist_username TEXT DEFAULT '',
    anilist_user_id  INT DEFAULT 0,
    anilist_token    TEXT DEFAULT '',
    mal_username     TEXT DEFAULT '',
    mal_user_id      INT DEFAULT 0,
    mal_token        TEXT DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tracks (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    anime_name  TEXT NOT NULL,
    artist      TEXT DEFAULT '',
    audio_url   TEXT NOT NULL,
    difficulty  INT DEFAULT 1,
    mal_id      INT DEFAULT 0,
    track_type  TEXT DEFAULT '',   -- 'OP', 'ED', 'OST'
    anime_year  INT DEFAULT 0,
    CONSTRAINT tracks_unique_track UNIQUE (mal_id, title, track_type)
);

CREATE TABLE game_results (
    id         SERIAL PRIMARY KEY,
    user_id    INT REFERENCES users(id),
    score      INT NOT NULL,
    xp_gained  INT NOT NULL,
    played_at  TIMESTAMPTZ DEFAULT NOW()
);
```

> Les index et évolutions de schéma ultérieures sont appliqués automatiquement au démarrage via `database.Migrate()`.

## Architecture

```
aniquiz/
├── cmd/api/          # Point d'entrée du serveur (main.go)
├── internal/
│   ├── database/     # Connexion PostgreSQL, requêtes, migrations
│   ├── game/         # Logique de jeu, salons WebSocket, clients
│   ├── handlers/     # Routes HTTP Gin, middlewares, JWT, rate limiter
│   ├── models/       # Structures de données partagées
│   └── sourcing/     # OAuth AniList/MAL, récupération des pistes audio
└── frontend/
    └── src/
        ├── components/   # Composants Vue (AuthForm, GameSettings, …)
        ├── router/       # Vue Router (SPA, catch-all 404)
        ├── App.vue       # Composant principal (jeu + WebSocket)
        ├── AppRoot.vue   # Racine SPA (RouterView)
        └── authStore.js  # Store réactif (user, token, localStorage)
```

Voir [docs/DEVOPS.md](docs/DEVOPS.md) pour la pipeline CI/CD, Docker et le guide de contribution.

## Tests

```bash
# Backend (Go) — avec détection des race conditions
go test -v -race ./...

# Frontend — tests unitaires (Vitest)
cd frontend && npm run test:unit

# Frontend — tests E2E (Playwright, nécessite un navigateur)
cd frontend && npx playwright install chromium
cd frontend && npm run test:e2e
```

## Contribuer

1. Créer une branche `feature/<nom>` depuis `main`
2. Pousser → la CI tourne automatiquement (lint, tests, build, Docker)
3. Ouvrir une Pull Request vers `main`
4. Merge uniquement si tous les jobs sont verts

La branche `main` est protégée : aucun push direct.
