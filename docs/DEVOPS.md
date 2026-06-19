# Documentation DevOps — AniQuiz

Ce document couvre la pipeline CI/CD, la conteneurisation, le reverse proxy, le monitoring, le déploiement (Docker Compose et Kubernetes) et les procédures opérationnelles. Il s'adresse à tout développeur qui reprend le projet.

> Pour la mise en production pas à pas, voir aussi [`PRODUCTION_CHECKLIST.md`](PRODUCTION_CHECKLIST.md).

---

## Sommaire

1. [Pipeline CI/CD](#1-pipeline-cicd)
2. [Docker](#2-docker)
3. [Reverse proxy Nginx](#3-reverse-proxy-nginx)
4. [Monitoring & logging](#4-monitoring--logging)
5. [Déploiement Kubernetes](#5-déploiement-kubernetes)
6. [Scripts opérationnels](#6-scripts-opérationnels)
7. [Schéma d'architecture](#7-schéma-darchitecture)
8. [Variables d'environnement](#8-variables-denvironnement)
9. [Fiabilité & résilience](#9-fiabilité--résilience)
10. [Protocole WebSocket](#10-protocole-websocket)
11. [API REST](#11-api-rest)
12. [Ajouter un test](#12-ajouter-un-test)
13. [Procédure de mise en production](#13-procédure-de-mise-en-production)
14. [Dépannage fréquent](#14-dépannage-fréquent)

---

## 1. Pipeline CI/CD

Fichier : `.github/workflows/go.yml`

La pipeline se déclenche sur :
- tout push sur `main`, une branche `feature/**` ou `working/**`
- tout push d'un tag `v*` (déclenche le déploiement de production)
- toute Pull Request vers `main`

Les runs concurrents sur la même branche sont annulés automatiquement (`cancel-in-progress: true`).

Une variable d'environnement au niveau du workflow fixe le nom d'image **en minuscules** (obligatoire pour les références OCI/Docker, alors que `github.repository` conserve la casse `LeYapson/aniquiz`) :

```yaml
env:
  IMAGE_NAME: ghcr.io/leyapson/aniquiz
```

### Ordre des jobs

```
security ──┐
           ├──► backend ──┐
           │              ├──► docker (build + push + scan) ──► deploy-staging  (push main)
           └──► frontend ─┤                                 └──► deploy-prod     (tag v*)
                          └──► e2e ──┘
```

### Détail de chaque job

#### `security` — Analyse de vulnérabilités
- `actions/setup-go@v5` avec cache du module Go (lit la version dans `go.mod`)
- **govulncheck** : CVE connues dans le code Go réellement atteint (bloquant)
- **Trivy (scan filesystem)** : CVE `CRITICAL`/`HIGH` dans les dépendances Go + npm
  - `exit-code: 0` → **rapport seulement**. Une CVE publiée demain ne doit pas faire passer un pipeline vert au rouge sans changement de code. Repasser à `exit-code: 1` une fois l'arbre de dépendances figé et trié.

#### `backend` — Qualité et tests Go
- `actions/setup-go@v5` (cache activé)
1. **Lint** : `go vet ./...` + `gofmt -l .` (bloque si code mal formaté)
2. **Build** : `go build -v ./...`
3. **Tests** : `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...`
   - `-race` : détecte les race conditions (données partagées sans mutex)
   - `-covermode=atomic` : coverage thread-safe
4. L'artefact `coverage.out` est conservé 7 jours

> **Note lint** : `golangci-lint` n'est volontairement pas utilisé. Sa v2 (installée par `version: latest`) refuse le fichier `.golangci.yml` au format v1 du projet et exposerait des erreurs `errcheck`/`goimports` pré-existantes. On garde le couple `go vet` + `gofmt`, déjà éprouvé. Migration possible plus tard (`golangci-lint migrate` + passage de l'action en `@v7/@v8`).

#### `frontend` — Qualité et build Vue
1. **Tests unitaires** : Vitest (`npm run test:unit`)
2. **Build** : Vite (`npm run build`)
3. L'artefact `frontend-dist` est conservé 1 jour

#### `e2e` — Tests de bout en bout (Playwright)
- Dépend de `frontend`
- Navigateur : Chromium uniquement (réduit les coûts CI)
- Rapport HTML uploadé comme artefact en cas d'échec (7 jours)

#### `docker` — Build, publication et scan de l'image
- Dépend de `backend` + `frontend` + `e2e` (les 3 doivent être verts)
- Permissions : `contents: read`, `packages: write`, `security-events: write`, `actions: read`
- Build avec cache GitHub Actions (`type=gha`), **provenance** et **SBOM** activés
- **Push uniquement sur `main` ou un tag `v*`** — les branches feature buildent sans pousser
- `build-args` injectés : `BUILD_VERSION`, `BUILD_COMMIT`, `BUILD_DATE` (voir Dockerfile + `main.go`)
- Tags générés par `docker/metadata-action` :
  - `sha-<court>` : SHA court immuable (`format=short`)
  - `<branche>` : nom de branche
  - `<version>` et `<major>.<minor>` : sur tag semver (le `v` est retiré → `v1.2.3` produit `1.2.3`)
  - `latest` : uniquement sur `main`
- **Scan d'image Trivy** par **digest** (exact, non ambigu) → rapport SARIF
- **Upload SARIF** vers l'onglet Security (`github/codeql-action/upload-sarif`)
  - `continue-on-error: true` : l'upload nécessite le *code scanning*, gratuit sur dépôt **public** mais soumis à GitHub Advanced Security sur dépôt **privé**. S'il est indisponible, le pipeline reste vert (le résultat du scan reste visible dans les logs du job).

#### `deploy-staging` — Déploiement automatique (staging)
- Dépend de `docker`
- Condition : `github.ref == 'refs/heads/main' && vars.ENABLE_DEPLOY == 'true'`
- **Tant que `ENABLE_DEPLOY` n'existe pas, le job est ignoré (gris, pas rouge)** — c'est l'état normal d'un projet sans infrastructure live.
- Étapes : résout le tag image (`sha-<court>`), SSH sur le serveur, `docker compose pull` + `up -d`, health check via l'ingress Nginx (`https://localhost/health`), **rollback automatique** sur l'image précédente en cas d'échec.

#### `deploy-prod` — Déploiement de production
- Dépend de `docker`
- Condition : `startsWith(github.ref, 'refs/tags/v') && vars.ENABLE_DEPLOY == 'true'`
- Déclenché en poussant un tag : `git tag v1.0.0 && git push --tags`
- Étapes : sauvegarde DB préalable (`scripts/backup-db.sh`), résout le tag image (semver sans `v`), `compose up -d`, health check via l'ingress, création d'une **GitHub Release** avec notes générées.

### Secrets et variables GitHub

**Secrets** (Settings → Secrets and variables → Actions → *Secrets*) :

| Secret | Usage |
|--------|-------|
| `GITHUB_TOKEN` | Automatique — push vers GHCR (`packages: write`) |
| `DEPLOY_SSH_KEY` | Clé privée SSH pour les jobs de déploiement |
| `DEPLOY_USER` | Utilisateur SSH sur le serveur cible |
| `STAGING_HOST` | Hôte du serveur de staging |
| `PROD_HOST` | Hôte du serveur de production |

**Variables** (onglet *Variables*) :

| Variable | Usage |
|----------|-------|
| `ENABLE_DEPLOY` | `true` pour activer les jobs de déploiement (sinon ignorés) |
| `STAGING_URL` / `PROD_URL` | URL affichée dans l'environnement GitHub |
| `DEPLOY_PATH` | Chemin du projet sur le serveur (défaut `/opt/aniquiz`) |

> Pour la CI seule (build + test + scan), seul `GITHUB_TOKEN` est nécessaire. Les autres secrets ne servent qu'au déploiement.

---

## 2. Docker

### Dockerfile (multi-stage)

```
Stage 1 : frontend-builder  (node:22-alpine)
  └─► npm ci && npm run build → /app/frontend/dist

Stage 2 : backend-builder   (golang:1.26-alpine)
  └─► go build -ldflags="-s -w -X main.Version=… -X main.Commit=… -X main.BuildDate=…"
      → /app/main

Stage 3 : final             (alpine:3.22)
  ├─► Labels OCI (version, revision, source…)
  ├─► ca-certificates + tzdata
  ├─► Copie /app/main           → ./main          (--chown appuser)
  ├─► Copie /app/frontend/dist  → ./static        (servi comme SPA fallback)
  ├─► Utilisateur non-root : appuser
  ├─► EXPOSE 8080
  └─► HEALTHCHECK : wget /health toutes les 30s
```

Image finale ~30 Mo (Alpine + binaire Go strippé + assets Vue statiques). Les bases `node:22-alpine`/`alpine:3.22` peuvent remonter des CVE dans le scan Trivy (avertissements) ; le gate étant en mode rapport, elles ne bloquent pas la CI. Pour une image au scan propre : épingler les digests ou viser une base *distroless*.

### docker-compose.yml (développement local)

```bash
docker compose up          # démarre db + backend + frontend
docker compose down -v     # arrête et supprime les volumes
```

- **db** : PostgreSQL 16 avec healthcheck, données persistées dans un volume nommé
- **backend** : `go run ./cmd/api`, code source monté
- **frontend** : `npm run dev` avec HMR

### docker-compose.prod.yml (production)

Services : **db**, **app**, **nginx**, **certbot**.

- **app** : image GHCR `ghcr.io/leyapson/aniquiz:${IMAGE_TAG:-latest}`, `GIN_MODE=release`, **port 8080 non publié** (seul Nginx est exposé), branché sur deux réseaux : `app_net` (interne) et `proxy_net`.
- **db** : réseau `app_net` **interne uniquement** (`internal: true`) — la base n'est jamais joignable depuis l'extérieur.
- **nginx** : expose `80`/`443`, terminaison TLS, monte `nginx/` et les volumes de certificats.
- **certbot** : renouvellement automatique Let's Encrypt (boucle toutes les 12 h).
- Toutes les variables sensibles sont **obligatoires** (`${JWT_SECRET:?…}`) — le compose refuse de démarrer si elles manquent.
- Chaque service a des **limites de ressources** (`deploy.resources.limits`) et une **rotation des logs** (`json-file`, `max-size`/`max-file`).

```bash
# Sur le serveur
cp .env.example .env       # remplir toutes les variables
IMAGE_TAG=sha-<court> docker compose -f docker-compose.prod.yml up -d
```

### docker-compose.monitoring.yml

Stack d'observabilité séparée — voir [§4](#4-monitoring--logging).

---

## 3. Reverse proxy Nginx

Fichiers : `nginx/nginx.conf` (global) et `nginx/conf.d/aniquiz.conf` (vhost).

Nginx est **le seul point d'entrée** ; l'application n'expose pas son port directement.

Fonctions assurées :
- **Terminaison TLS** (HTTP/2), redirection `80 → 443`, challenge ACME Let's Encrypt
- **En-têtes de sécurité** : HSTS, `X-Frame-Options: DENY`, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`
- **Rate limiting** par zones :
  - `auth` : 5 req/min (endpoints `/api/auth/login|register`)
  - `api` : 30 req/s (autres `/api/*`)
  - `ws` : 10 req/s (`/ws`)
  - limite de connexions simultanées par IP
- **Proxy WebSocket** sur `/ws` (`Upgrade`/`Connection`, timeouts longs 3600 s)
- **Cache agressif** des assets statiques (`/assets/`, `immutable`, 1 an)
- **Logs JSON** (`json_combined`) prêts pour l'ingestion Loki

> ⚠ Le fichier `conf.d/aniquiz.conf` est monté **directement** dans `/etc/nginx/conf.d` : Nginx n'y applique **pas** `envsubst`. Le domaine est donc **codé en dur** (`aniquiz.example.com`) — à remplacer par votre domaine réel avant déploiement. Alternative : déplacer le fichier dans `/etc/nginx/templates/*.template` et définir `NGINX_ENVSUBST_FILTER=DOMAIN` pour ne substituer que `$DOMAIN` (sans toucher à `$host`, `$remote_addr`, etc.).

### Premier certificat TLS

```bash
docker run --rm -p 80:80 \
  -v letsencrypt_certs:/etc/letsencrypt \
  -v certbot_webroot:/var/www/certbot \
  certbot/certbot certonly --standalone \
  -d aniquiz.example.com --email you@example.com --agree-tos
```

---

## 4. Monitoring & logging

Fichier : `docker-compose.monitoring.yml`. Configs sous `monitoring/`.

| Service | Rôle | Port (lié à `127.0.0.1`) |
|---------|------|--------------------------|
| **Prometheus** | Collecte des métriques (rétention 30 j) | 9090 |
| **Grafana** | Dashboards (datasources auto-provisionnées) | 3000 |
| **cAdvisor** | Métriques par conteneur | 8081 |
| **node-exporter** | Métriques hôte (CPU, RAM, disque) | host |
| **postgres-exporter** | Métriques PostgreSQL | 9187 |
| **Loki** | Agrégation des logs (rétention 30 j) | 3100 |
| **Promtail** | Collecteur de logs Docker → Loki | — |

- Tous les ports d'admin sont liés à `127.0.0.1` (jamais exposés publiquement — passer par un tunnel SSH pour y accéder).
- Grafana est provisionné automatiquement avec les datasources **Prometheus** et **Loki** (`monitoring/grafana/provisioning/`).
- Promtail lit les logs Docker et parse le format JSON de Nginx (status, méthode, durée).

```bash
# Démarrer la stack (après la stack prod, qui crée les réseaux partagés)
docker compose -f docker-compose.monitoring.yml up -d

# Accès Grafana depuis un poste local
ssh -L 3000:127.0.0.1:3000 user@serveur   # puis http://localhost:3000
```

> Les réseaux `aniquiz_proxy_net` et `aniquiz_app_net` sont déclarés `external` : la stack prod doit être démarrée **avant** la stack monitoring.

### Sondes de santé applicatives

- `GET /health` vérifie la connexion PostgreSQL (`200 ok` / `503 unhealthy`).
- Le `HEALTHCHECK` Docker et les `livenessProbe`/`readinessProbe` Kubernetes s'appuient dessus.

---

## 5. Déploiement Kubernetes

Manifestes sous `k8s/`. Alternative au Docker Compose pour un cluster.

```
k8s/
├── namespace.yml          # namespace "aniquiz"
├── configmap.yml          # GIN_MODE, PORT, POSTGRES_DB/USER
├── ingress.yml            # Ingress Nginx + TLS cert-manager + annotations WebSocket
├── postgres/
│   ├── statefulset.yml    # PostgreSQL + PVC 10Gi + probes pg_isready
│   └── service.yml        # ClusterIP headless (jamais exposé)
└── app/
    ├── deployment.yml     # rolling update maxUnavailable:0, probes, securityContext
    ├── service.yml        # ClusterIP :80 → 8080
    └── hpa.yml            # autoscaling 1→5 (CPU 70 % / RAM 80 %)
```

Points clés :
- **RollingUpdate `maxUnavailable: 0`** → zéro downtime au déploiement.
- **Probes** `liveness`/`readiness` sur `/health`.
- **securityContext** : non-root (uid 1001), `allowPrivilegeEscalation: false`, `capabilities: drop ALL`.
- Un secret `aniquiz-secrets` (non versionné) doit fournir `DATABASE_URL`, `JWT_SECRET`, `FRONTEND_URL`, `POSTGRES_PASSWORD`. Un `imagePullSecrets` `ghcr-credentials` est requis pour tirer l'image GHCR.

```bash
kubectl apply -f k8s/namespace.yml
kubectl apply -f k8s/configmap.yml
# créer les secrets aniquiz-secrets + ghcr-credentials
kubectl apply -f k8s/postgres/ -f k8s/app/ -f k8s/ingress.yml
kubectl rollout status deployment/aniquiz -n aniquiz
```

> ⚠ **Limite de scaling** : voir [§9](#9-fiabilité--résilience). `replicas` et `minReplicas` sont volontairement à **1**.

---

## 6. Scripts opérationnels

Dossier `scripts/`.

### `deploy.sh` — Déploiement zéro-downtime avec rollback

```bash
./scripts/deploy.sh sha-<court>
```
- Capture l'image courante (pour rollback), tire la nouvelle, redémarre **uniquement le service `app`** (la DB reste up).
- Health check via l'ingress (`https://localhost/health`) avec retries.
- **Rollback automatique** sur l'image précédente si le health check échoue.
- Variables surchargables : `HEALTH_URL`, `CURL_OPTS`, `COMPOSE_FILE`.

### `backup-db.sh` — Sauvegarde PostgreSQL

```bash
./scripts/backup-db.sh
```
- `pg_dump` compressé gzip, horodaté, dans `BACKUP_DIR` (défaut `/opt/aniquiz/backups`).
- Vérifie que le dump n'est pas vide, applique une **rétention de 14 jours**.
- À planifier en cron : `0 3 * * * /opt/aniquiz/scripts/backup-db.sh`.

> Recommandation : copier les sauvegardes hors-serveur (rsync/S3/Backblaze) et **tester une restauration** avant le premier déploiement.

---

## 7. Schéma d'architecture

```
Internet
    │ 80 / 443
    ▼
┌─────────────────────────────────────────────┐
│  Nginx (reverse proxy)                        │
│  • TLS (Let's Encrypt + renouvellement auto)  │
│  • En-têtes sécurité (HSTS, X-Frame-Options…) │
│  • Rate limiting (auth / api / ws)            │
│  • Logs JSON → Promtail → Loki                │
└───────────────┬───────────────────────────────┘
                │ proxy_net
┌───────────────▼───────────────────────────────┐
│  App (Go + Gin)  — port 8080 non publié        │
│  • /api/*  REST   • /ws  WebSocket  • SPA       │
│  • Graceful shutdown (drain 30 s sur SIGTERM)   │
│  • CheckOrigin WebSocket = FRONTEND_URL         │
└───────────────┬───────────────────────────────┘
                │ app_net (interne — aucun accès externe)
┌───────────────▼───────────────────────────────┐
│  PostgreSQL 16                                  │
└─────────────────────────────────────────────────┘

Observabilité (ports liés à 127.0.0.1) :
  Prometheus ← cAdvisor + node-exporter + postgres-exporter
  Grafana    ← Prometheus + Loki
  Promtail   → logs conteneurs → Loki
```

Le backend sert l'API REST, les WebSockets **et** le frontend buildé (SPA fallback). Nginx est le seul composant exposé.

---

## 8. Variables d'environnement

### Application (obligatoires en production)

| Variable | Exemple | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:pass@db:5432/aniquiz?sslmode=disable` | Connexion PostgreSQL (pgx v5) |
| `JWT_SECRET` | Chaîne aléatoire 64+ caractères (`openssl rand -hex 64`) | Signature des tokens JWT (HS256) |
| `FRONTEND_URL` | `https://aniquiz.example.com` | CORS, redirections OAuth **et** validation d'origine WebSocket |

### Application (optionnelles)

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8080` | Port d'écoute du serveur |
| `GIN_MODE` | `release` (prod) | Mode Gin ; en `release`, les routes debug renvoient 404 |
| `IMAGE_TAG` | `latest` | Tag d'image GHCR à exécuter (compose prod) |

### PostgreSQL (docker-compose.prod.yml)

| Variable | Défaut | Description |
|----------|--------|-------------|
| `POSTGRES_USER` | `aniquiz` | Utilisateur PostgreSQL |
| `POSTGRES_PASSWORD` | — (obligatoire) | Mot de passe |
| `POSTGRES_DB` | `aniquiz` | Nom de la base |

### Monitoring (docker-compose.monitoring.yml)

| Variable | Défaut | Description |
|----------|--------|-------------|
| `GRAFANA_USER` | `admin` | Login admin Grafana |
| `GRAFANA_PASSWORD` | — (obligatoire) | Mot de passe admin Grafana |
| `GRAFANA_URL` | `http://localhost:3000` | Root URL Grafana |

### OAuth (optionnels)

| Variable | Description |
|----------|-------------|
| `ANILIST_CLIENT_ID` / `ANILIST_CLIENT_SECRET` / `ANILIST_REDIRECT_URI` | Application AniList (https://anilist.co/settings/developer) |
| `MAL_CLIENT_ID` / `MAL_CLIENT_SECRET` / `MAL_REDIRECT_URI` | Application MAL (https://myanimelist.net/apiconfig) |

> Sans les variables OAuth, le site fonctionne normalement. Seule la fonctionnalité « Filtrer par ma liste » est désactivée.

---

## 9. Fiabilité & résilience

- **Arrêt propre (graceful shutdown)** : sur `SIGTERM`/`SIGINT`, le serveur arrête d'accepter de nouvelles connexions et draine les requêtes en cours pendant 30 s max (`http.Server.Shutdown`). Indispensable pour les rolling updates Docker/K8s.
- **Timeouts HTTP** : `ReadTimeout`, `WriteTimeout` (60 s pour l'upgrade WebSocket), `IdleTimeout` configurés sur le serveur.
- **Validation d'origine WebSocket** : `CheckOrigin` n'accepte que l'origine `FRONTEND_URL` (permissif uniquement si la variable est vide, en dev).
- **Routes debug désactivées en production** : `/test-audio` et `/anime/:id` renvoient `404` quand `GIN_MODE=release` (évite la fuite d'informations et les écritures non authentifiées).
- **Health checks** à tous les niveaux : Docker `HEALTHCHECK`, probes K8s, et health check applicatif dans les scripts de déploiement.
- **Rollback automatique** : `deploy.sh` et le job `deploy-staging` reviennent à l'image précédente si le health check post-déploiement échoue.
- **Sauvegardes** : `backup-db.sh` (rétention 14 j), exécuté avant chaque déploiement de production.

### ⚠ Limite de scaling — salons WebSocket en mémoire

Les salons de jeu (`internal/game/room.go`, map `ActiveRooms`) sont stockés **en mémoire dans le process**. Conséquences :

- **Une seule réplique** tant que Redis n'est pas intégré. Ne pas augmenter `replicas` / `minReplicas` au-delà de 1 — des joueurs sur des pods différents seraient dans des salons isolés.
- **Pour activer le scaling horizontal** :
  1. Ajouter Redis à la stack.
  2. Remplacer la map `ActiveRooms` par un registre de salons backé par Redis.
  3. Remplacer le broadcast in-process par du **Redis pub/sub** par salon.
  4. Passer alors `minReplicas: 2` dans le HPA.

---

## 10. Protocole WebSocket

Endpoint : `GET /ws?room=<room_id>&token=<jwt>&password=<password>`

La connexion est refusée si le token JWT est invalide, si le salon n'existe pas, ou si l'origine ne correspond pas à `FRONTEND_URL`.

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

## 11. API REST

### Publiques (sans authentification)

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/ping` | Healthcheck simple (`{ "message": "pong" }`) |
| GET | `/health` | Vérifie la connexion à la base de données |
| GET | `/rooms` | Liste des salons publics actifs |
| GET | `/api/leaderboard` | Top 50 joueurs (XP décroissant) |
| GET | `/api/leaderboard/speedrun` | Classement speedrun |
| GET | `/animes` | Noms d'animes distincts |
| POST | `/api/auth/register` | Inscription (rate-limité : 5 req/min/IP) |
| POST | `/api/auth/login` | Connexion → JWT + user (rate-limité) |
| GET | `/api/auth/anilist` + `/callback` | Flux OAuth AniList |
| GET | `/api/auth/mal` + `/callback` | Flux OAuth MAL (PKCE) |

### Protégées (Bearer JWT requis)

| Méthode | Route | Description |
|---------|-------|-------------|
| POST | `/rooms` | Créer un salon |
| GET | `/quiz/next` | Piste aléatoire (debug) |
| POST | `/quiz/answer` | Vérifier une réponse (debug) |
| GET | `/api/profile` | Profil de l'utilisateur connecté |
| GET | `/api/history` | Dernières parties |
| GET | `/api/me/anime-ids` | MAL IDs de la liste AniList + MAL de l'utilisateur |
| GET | `/api/anime/search` | Recherche d'anime (Jikan API) |
| POST | `/api/admin/import` | Import batch de pistes |
| POST | `/api/speedrun/start\|answer\|skip\|finish` | Mode speedrun |

### Debug (désactivées si `GIN_MODE=release`)

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/anime/:id` | Sourcing d'un anime (écritures DB) — **404 en prod** |
| GET | `/test-audio` | Page de test listant toutes les pistes — **404 en prod** |

---

## 12. Ajouter un test

### Test unitaire Go

Créer un fichier `*_test.go` dans le package concerné :

```go
func TestMonFeature(t *testing.T) {
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

Ajouter un `test()` dans `frontend/e2e/game.spec.js`, puis : `cd frontend && npm run test:e2e`.

---

## 13. Procédure de mise en production

> Voir [`PRODUCTION_CHECKLIST.md`](PRODUCTION_CHECKLIST.md) pour la checklist complète (sécurité, TLS, vérifications post-déploiement).

### Première installation

```bash
ssh user@serveur.example.com
git clone https://github.com/LeYapson/aniquiz.git && cd aniquiz

cp .env.example .env            # remplir JWT_SECRET, POSTGRES_PASSWORD, FRONTEND_URL, OAuth…
chmod 600 .env

# Remplacer aniquiz.example.com dans nginx/conf.d/aniquiz.conf, puis émettre le certificat TLS (voir §3)

IMAGE_TAG=latest docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.monitoring.yml up -d

curl -k https://localhost/health
```

### Activer le déploiement automatique

1. Ajouter les secrets GitHub (`DEPLOY_SSH_KEY`, `DEPLOY_USER`, `STAGING_HOST`/`PROD_HOST`).
2. Créer les environnements GitHub `staging` et `production`.
3. Définir la variable `ENABLE_DEPLOY=true` :
   ```bash
   gh variable set ENABLE_DEPLOY --body true
   ```
- Push sur `main` → déploiement **staging** automatique.
- Tag `v*` (`git tag v1.0.0 && git push --tags`) → déploiement **production** + Release.

### Mise à jour manuelle

```bash
./scripts/deploy.sh sha-<court>     # zéro-downtime + rollback automatique
```

---

## 14. Dépannage fréquent

### CI — `Unable to resolve action aquasecurity/trivy-action@0.28.0`
Les tags de cette action sont **préfixés par `v`** (`v0.36.0`). Référencer `aquasecurity/trivy-action@v0.36.0`, pas `@0.28.0`.

### CI — `could not parse reference: ghcr.io/LeYapson/aniquiz@sha256:…`
Les références d'image OCI doivent être **en minuscules**, or `github.repository` conserve la casse. Utiliser la variable `env.IMAGE_NAME: ghcr.io/leyapson/aniquiz` partout (déjà en place).

### CI — `Resource not accessible by integration` (upload SARIF)
`upload-sarif` a besoin de la permission **`actions: read`** (en plus de `security-events: write`). Un bloc `permissions:` est restrictif : seules les portées listées sont accordées. L'envoi vers l'onglet Security nécessite par ailleurs le *code scanning* (gratuit en public, GitHub Advanced Security en privé) — d'où le `continue-on-error: true`.

### Les jobs `deploy-staging`/`deploy-prod` sont **ignorés (gris)**
**Normal** tant que la variable `ENABLE_DEPLOY` n'est pas définie à `true`. Voir [§1](#1-pipeline-cicd) et [§13](#13-procédure-de-mise-en-production).

### Nginx ne démarre pas / `${DOMAIN}` littéral
`conf.d/` n'est pas traité par `envsubst`. Coder le domaine en dur ou utiliser `/etc/nginx/templates/` avec `NGINX_ENVSUBST_FILTER=DOMAIN`. Voir [§3](#3-reverse-proxy-nginx).

### Health check de déploiement en échec
Le port `8080` n'est **pas publié** en prod — interroger l'ingress Nginx (`https://localhost/health`, option `-k`), pas `localhost:8080`.

### `column "..." does not exist`
Migrations automatiques au démarrage via `database.Migrate()`. Vérifier que le serveur a démarré sans erreur fatale.

### `token invalide` à la connexion WebSocket
Le JWT est signé avec `JWT_SECRET`. Si la variable change entre deux démarrages, tous les tokens existants deviennent invalides. Définir un `JWT_SECRET` **stable** en production.

### WebSocket refusé en dev sur `127.0.0.1`
`CheckOrigin` compare l'origine exacte à `FRONTEND_URL`. `http://localhost:5173` et `http://127.0.0.1:5173` sont deux origines distinctes. Utiliser la même URL que `FRONTEND_URL`, ou laisser `FRONTEND_URL` vide en dev.

### `gofmt -l` liste des fichiers en local (Windows)
Artefact de fins de ligne CRLF. Git stocke en LF et la CI (Linux) voit du LF → `gofmt` y passe. Le fichier `.gitattributes` (`* text=auto eol=lf`) évite cette confusion.

### Race condition détectée par `-race`
Tout accès concurrent à une variable partagée sans mutex échoue en CI. Les structures `Room` et `Client` utilisent `sync.Mutex` — verrouiller avant lecture/écriture des champs partagés.
