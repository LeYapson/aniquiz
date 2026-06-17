# Documentation DevOps — AniQuiz

Ce document couvre la pipeline CI/CD, la conteneurisation Docker et les procédures opérationnelles. Il s'adresse à tout développeur qui reprend le projet.

---

## Sommaire

1. [Pipeline CI/CD](#1-pipeline-cicd)
2. [Docker](#2-docker)
3. [Schéma de déploiement](#3-schéma-de-déploiement)
4. [Variables d'environnement](#4-variables-denvironnement)
5. [Protocole WebSocket](#5-protocole-websocket)
6. [API REST](#6-api-rest)
7. [Ajouter un test](#7-ajouter-un-test)
8. [Procédure de mise en production](#8-procédure-de-mise-en-production)
9. [Dépannage fréquent](#9-dépannage-fréquent)

---

## 1. Pipeline CI/CD

Fichier : `.github/workflows/go.yml`

La pipeline se déclenche sur :
- tout push sur `main` ou une branche `feature/**`
- toute Pull Request vers `main`

Les runs concurrents sur la même branche sont annulés automatiquement (`cancel-in-progress: true`).

### Ordre des jobs

```
security ──┐
           ├──► backend ──┐
           │              ├──► docker (build + push GHCR)
           └──► frontend ─┤
                          └──► e2e ──┘
```

### Détail de chaque job

#### `security` — Analyse de vulnérabilités
- Outil : `govulncheck` (base de données officielle Go)
- Analyse toutes les dépendances Go transitives
- Bloque si une CVE connue est détectée dans le code réellement atteint

#### `backend` — Qualité et tests Go
1. **Lint** : `go vet ./...` + `gofmt -l .` (bloque si code mal formaté)
2. **Build** : `go build -v ./...` (vérifie que tout compile)
3. **Tests** : `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...`
   - `-race` : détecte les race conditions (données partagées sans mutex)
   - `-covermode=atomic` : coverage thread-safe
4. L'artefact `coverage.out` est conservé 7 jours

#### `frontend` — Qualité et build Vue
1. **Tests unitaires** : Vitest (`npm run test:unit`)
2. **Build** : Vite (`npm run build`) — vérifie qu'il n'y a pas d'erreur TS/JS

#### `e2e` — Tests de bout en bout (Playwright)
- Dépend de `frontend`
- Navigateur : Chromium uniquement (réduit les coûts CI)
- 1 retry automatique en cas d'échec
- Rapport HTML uploadé comme artefact en cas d'échec (7 jours)

#### `docker` — Build et publication de l'image
- Dépend de `backend` + `frontend` + `e2e` (les 3 doivent être verts)
- Build avec cache GitHub Actions (`type=gha`) pour accélérer les rebuilds
- **Push uniquement sur `main`** — les branches feature buildent sans pousser
- Tags générés :
  - `sha-<commit_sha>` : identifiant immuable
  - `latest` : alias vers le dernier commit de main
- Registre : `ghcr.io/leyapson/aniquiz`

### Secrets GitHub requis

| Secret | Usage |
|--------|-------|
| `GITHUB_TOKEN` | Automatique — push vers GHCR (packages: write) |

> Aucun autre secret n'est nécessaire pour la CI. Les secrets OAuth/DB ne sont utilisés qu'en production.

---

## 2. Docker

### Dockerfile (multi-stage)

```
Stage 1 : frontend-builder  (node:20-alpine)
  └─► npm ci && npm run build → /app/dist

Stage 2 : backend-builder   (golang:1.26-alpine)
  └─► go build -ldflags="-s -w" → /app/server

Stage 3 : final             (alpine:3.21)
  ├─► Copie /app/dist  → /app/static  (servi comme SPA fallback)
  ├─► Copie /app/server → /app/server
  ├─► Utilisateur non-root : appuser (uid 1001)
  ├─► EXPOSE 8080
  └─► HEALTHCHECK : GET /health toutes les 30s
```

L'image finale pèse ~30 Mo (Alpine + binaire Go strippé + assets Vue statiques).

### docker-compose.yml (développement local)

```bash
docker compose up          # démarre db + backend + frontend
docker compose down -v     # arrête et supprime les volumes
```

- **db** : PostgreSQL 16 avec healthcheck, données persistées dans un volume nommé
- **backend** : exécute `go run ./cmd/api` avec le code source monté (rechargement manuel)
- **frontend** : `npm run dev` avec HMR (Hot Module Replacement)

### docker-compose.prod.yml (production)

```bash
# Sur le serveur de production
cp .env.example .env  # remplir toutes les variables
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

- Utilise l'image pré-compilée depuis GHCR (`ghcr.io/leyapson/aniquiz:latest`)
- Aucun code source monté — image immutable
- `restart: always` pour la résilience aux redémarrages

### Mettre à jour la production

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --no-build
```

---

## 3. Schéma de déploiement

```
Internet
    │
    ▼
[ Reverse Proxy / Load Balancer ]  (ex: Nginx, Caddy, Traefik)
    │
    ├──► :8080  [ Container app ]
    │              │
    │              ├── GET /assets/*  → fichiers Vue statiques (dist/)
    │              ├── GET /api/*     → handlers Gin (JSON)
    │              ├── GET /ws        → upgrade WebSocket
    │              └── GET /*         → SPA fallback (index.html)
    │
    └──► :5432  [ Container db ]  (accès interne uniquement)
```

Le backend sert à la fois l'API REST, les WebSockets **et** le frontend buildé (en production). Il n'y a qu'un seul port exposé.

---

## 4. Variables d'environnement

### Obligatoires en production

| Variable | Exemple | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:pass@db:5432/aniquiz?sslmode=disable` | Connexion PostgreSQL (pgx v5) |
| `JWT_SECRET` | Chaîne aléatoire 64+ caractères | Signature des tokens JWT (HS256) |
| `FRONTEND_URL` | `https://aniquiz.example.com` | Utilisé pour CORS et les redirections OAuth |

### PostgreSQL (docker-compose.prod.yml)

| Variable | Défaut | Description |
|----------|--------|-------------|
| `POSTGRES_USER` | `aniquiz` | Utilisateur PostgreSQL |
| `POSTGRES_PASSWORD` | — | Mot de passe (obligatoire) |
| `POSTGRES_DB` | `aniquiz` | Nom de la base |

### OAuth (optionnels)

| Variable | Description |
|----------|-------------|
| `ANILIST_CLIENT_ID` | Application créée sur https://anilist.co/settings/developer |
| `ANILIST_CLIENT_SECRET` | Secret de l'application AniList |
| `ANILIST_REDIRECT_URI` | Ex: `https://aniquiz.example.com/api/auth/anilist/callback` |
| `MAL_CLIENT_ID` | Application créée sur https://myanimelist.net/apiconfig |
| `MAL_CLIENT_SECRET` | Secret de l'application MAL |
| `MAL_REDIRECT_URI` | Ex: `https://aniquiz.example.com/api/auth/mal/callback` |

> Sans les variables OAuth, le site fonctionne normalement. Seule la fonctionnalité "Filtrer par ma liste" est désactivée.

---

## 5. Protocole WebSocket

Endpoint : `GET /ws?room=<room_id>&token=<jwt>&password=<password>`

La connexion est refusée si le token JWT est invalide ou si le salon n'existe pas.

### Messages client → serveur

```jsonc
{ "type": "START_GAME",     "payload": null }
{ "type": "SUBMIT_ANSWER",  "payload": "Naruto Shippuden" }
{ "type": "CHAT",           "payload": "Trop facile !" }
{ "type": "REACTION",       "payload": "🔥" }  // émojis autorisés : 🔥 🤔 😱 ✅ 😭 👏
{ "type": "UPDATE_SETTINGS", "payload": {
    "max_rounds": 10,
    "round_duration": 30,
    "filter_type": "OP",   // "" | "OP" | "ED"
    "min_year": 2010,
    "max_year": 2019,
    "is_private": false,
    "password": "",
    "filter_mal_ids": [20, 1535]  // [] = pas de filtre
}}
```

> `UPDATE_SETTINGS` n'est accepté que du créateur du salon, en état LOBBY.

### Messages serveur → client

```jsonc
{ "type": "SPECTATOR_STATUS",   "payload": false }
{ "type": "GAME_STATE",         "payload": "LOBBY" }  // "LOBBY" | "PLAYING"
{ "type": "PLAYER_LIST",        "payload": { "players": [...], "spectator_count": 2 }}
{ "type": "NewQuestion",        "payload": { "audio_url": "...", "room_id": "test", "duration": 20 }}
{ "type": "PLAYER_GUESS",       "payload": { "username": "Alice", "is_first": true }}
{ "type": "ROUND_ENDED",        "payload": {
    "reason": "Temps écoulé !",
    "answer": "Naruto Shippuden",
    "title": "GO!!!",
    "artist": "FLOW",
    "track_type": "OP",
    "difficulty": 2,
    "video_url": "<même url que audio_url, fichier WebM>",
    "found_by": [{ "username": "Alice", "time_ms": 4200, "bonus": 10 }]
}}
{ "type": "GAME_OVER",          "payload": { "message": "La partie est terminée !" }}
{ "type": "XP_GAINED",          "payload": { "xp_gained": 150, "new_xp": 3200, "new_level": 6 }}
{ "type": "CHAT_MESSAGE",       "payload": { "username": "Bob", "message": "GG !" }}
{ "type": "REACTION_BROADCAST", "payload": { "username": "Alice", "emoji": "🔥" }}
{ "type": "SETTINGS_UPDATED",   "payload": { "max_rounds": 10, ... }}
```

---

## 6. API REST

### Publiques (sans authentification)

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/ping` | Healthcheck simple (`{ "message": "pong" }`) |
| GET | `/health` | Vérifie la connexion à la base de données |
| GET | `/rooms` | Liste des salons publics actifs |
| GET | `/api/leaderboard` | Top 50 joueurs (XP décroissant) |
| POST | `/api/auth/register` | Inscription (rate-limité : 5 req/min/IP) |
| POST | `/api/auth/login` | Connexion → JWT + user (rate-limité) |
| GET | `/api/auth/anilist` | Démarre le flux OAuth AniList |
| GET | `/api/auth/anilist/callback` | Callback OAuth AniList |
| GET | `/api/auth/mal` | Démarre le flux OAuth MAL (PKCE) |
| GET | `/api/auth/mal/callback` | Callback OAuth MAL |

### Protégées (Bearer JWT requis)

| Méthode | Route | Description |
|---------|-------|-------------|
| POST | `/rooms` | Créer un salon |
| GET | `/quiz/next` | Piste aléatoire (debug) |
| POST | `/quiz/answer` | Vérifier une réponse (debug) |
| GET | `/api/profile` | Profil de l'utilisateur connecté |
| GET | `/api/history` | 20 dernières parties |
| GET | `/api/me/anime-ids` | MAL IDs de la liste AniList + MAL de l'utilisateur |
| GET | `/api/anime/search` | Recherche d'anime (Jikan API) |
| POST | `/api/admin/import` | Import batch de pistes |

---

## 7. Ajouter un test

### Test unitaire Go

Créer un fichier `*_test.go` dans le package concerné :

```go
func TestMonFeature(t *testing.T) {
    // ...
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

Les tests sont automatiquement inclus dans la CI via `go test ./...`.

### Test unitaire Vue (Vitest)

Créer un fichier `frontend/src/components/__tests__/MonComposant.spec.js` :

```js
import { mount } from '@vue/test-utils'
import MonComposant from '../MonComposant.vue'

test('rendu correct', () => {
  const wrapper = mount(MonComposant)
  expect(wrapper.text()).toContain('AniQuiz')
})
```

### Test E2E (Playwright)

Ajouter un `test()` dans `frontend/e2e/game.spec.js` :

```js
test('mon scénario', async ({ page }) => {
  await page.goto('http://localhost:5173')
  await expect(page.getByRole('heading')).toBeVisible()
})
```

Lancer localement : `cd frontend && npm run test:e2e`

---

## 8. Procédure de mise en production

> Prérequis : un serveur avec Docker + Docker Compose, accessible via SSH.

```bash
# 1. Se connecter au serveur
ssh user@mon-serveur.example.com

# 2. Cloner le repo (première fois seulement)
git clone https://github.com/LeYapson/aniquiz.git
cd aniquiz

# 3. Configurer les variables d'environnement
cp .env.example .env
nano .env  # remplir DATABASE_URL, JWT_SECRET, FRONTEND_URL, etc.

# 4. Démarrer
docker compose -f docker-compose.prod.yml up -d

# 5. Vérifier
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8080/health
```

**Mise à jour après un nouveau push sur main :**

```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --no-build
# Zéro downtime si le container précédent répond encore pendant le pull
```

**Voir les logs :**

```bash
docker compose -f docker-compose.prod.yml logs -f app
docker compose -f docker-compose.prod.yml logs -f db
```

---

## 9. Dépannage fréquent

### `column "..." does not exist`
Le schéma de la base est incomplet. Les migrations automatiques s'exécutent au démarrage via `database.Migrate()`. Vérifier que le serveur a démarré sans erreur fatale.

### `token invalide` à la connexion WebSocket
Le token JWT est signé avec `JWT_SECRET`. Si cette variable diffère entre les environnements (ex: restart du container sans persistence du secret), tous les tokens existants deviennent invalides. Définir un `JWT_SECRET` stable en production.

### Tests E2E en échec en CI
Les tests E2E Playwright nécessitent qu'un serveur de dev tourne sur `localhost:5173`. Vérifier que le job `frontend` a bien buildé et que `npm run test:e2e` démarre bien le serveur Vite en mode preview.

### `imported and not used` à la compilation Go
Chaque import Go doit être utilisé. Si un package est importé pour préparer une feature, s'assurer d'ajouter au moins un appel avant de pousser.

### Race condition détectée par `-race`
Le flag `-race` est activé en CI. Tout accès concurrent à une variable partagée sans mutex provoque un échec. Les structures `Room` et `Client` utilisent `sync.Mutex` — s'assurer de verrouiller avant toute lecture/écriture des champs partagés.
