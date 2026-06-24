#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# server-setup.sh — First-time production setup for a fresh Ubuntu 24.04 server
# (e.g. a Hetzner Cloud CX22). Installs Docker, generates secrets, configures the
# domain, issues the Let's Encrypt certificate, and starts the full stack.
#
# USAGE (run as root on the server):
#   DOMAIN=aniquiz.tondomaine.com EMAIL=toi@exemple.com bash server-setup.sh
#
# Optional vars:
#   REPO=...        git URL (default: this project's GitHub)
#   APP_DIR=...     install path (default: /opt/aniquiz)
#   GHCR_TOKEN=...  + GHCR_USER=...  → only needed if the GHCR image is PRIVATE
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

DOMAIN="${DOMAIN:?Set DOMAIN=aniquiz.tondomaine.com}"
EMAIL="${EMAIL:?Set EMAIL=toi@exemple.com (needed for the TLS certificate)}"
REPO="${REPO:-https://github.com/LeYapson/aniquiz.git}"
APP_DIR="${APP_DIR:-/opt/aniquiz}"

log() { echo -e "\n\033[1;36m=== $* ===\033[0m"; }

# ─── 1. Base packages ─────────────────────────────────────────────────────────
log "Installing base packages (git, curl, openssl)"
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq git curl openssl ca-certificates

# ─── 2. Docker (+ compose plugin) ─────────────────────────────────────────────
if ! command -v docker >/dev/null 2>&1; then
  log "Installing Docker"
  curl -fsSL https://get.docker.com | sh
else
  log "Docker already installed — skipping"
fi

# ─── 3. Firewall (allow SSH first, then web) ──────────────────────────────────
if command -v ufw >/dev/null 2>&1; then
  log "Configuring firewall (22, 80, 443)"
  ufw allow 22/tcp
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable
else
  echo "ufw not found — use Hetzner's Cloud firewall to allow 22/80/443 instead."
fi

# ─── 4. Optional: log in to GHCR (only if the image is private) ───────────────
if [ -n "${GHCR_TOKEN:-}" ]; then
  log "Logging in to GHCR"
  echo "${GHCR_TOKEN}" | docker login ghcr.io -u "${GHCR_USER:-LeYapson}" --password-stdin
fi

# ─── 5. Clone (or update) the project ─────────────────────────────────────────
if [ ! -d "${APP_DIR}/.git" ]; then
  log "Cloning ${REPO} into ${APP_DIR}"
  git clone "${REPO}" "${APP_DIR}"
else
  log "Repo already present — pulling latest"
  git -C "${APP_DIR}" pull --ff-only || true
fi
cd "${APP_DIR}"

# ─── 6. Generate .env with random secrets (first time only) ───────────────────
if [ ! -f .env ]; then
  log "Generating .env with random secrets"
  JWT="$(openssl rand -hex 64)"
  PGPASS="$(openssl rand -hex 24)"
  cat > .env <<EOF
# Généré automatiquement par server-setup.sh — NE PAS commiter
POSTGRES_USER=aniquiz
POSTGRES_PASSWORD=${PGPASS}
POSTGRES_DB=aniquiz
JWT_SECRET=${JWT}
FRONTEND_URL=https://${DOMAIN}
IMAGE_TAG=latest

# OAuth (optionnel — remplir si tu utilises AniList / MyAnimeList)
ANILIST_CLIENT_ID=
ANILIST_CLIENT_SECRET=
ANILIST_REDIRECT_URI=https://${DOMAIN}/api/auth/anilist/callback
MAL_CLIENT_ID=
MAL_CLIENT_SECRET=
MAL_REDIRECT_URI=https://${DOMAIN}/api/auth/mal/callback
EOF
  chmod 600 .env
else
  log ".env already exists — leaving it untouched"
fi

# ─── 7. Inject the domain into the nginx vhost ────────────────────────────────
log "Setting domain ${DOMAIN} in nginx config"
sed -i "s/aniquiz\.example\.com/${DOMAIN}/g" nginx/conf.d/aniquiz.conf

# ─── 8. Issue the TLS certificate (standalone; needs port 80 free) ────────────
# On first run nginx isn't started yet, so port 80 is free for the ACME challenge.
if ! docker run --rm -v aniquiz_letsencrypt_certs:/etc/letsencrypt alpine \
       test -d "/etc/letsencrypt/live/${DOMAIN}" 2>/dev/null; then
  log "Requesting Let's Encrypt certificate for ${DOMAIN}"
  docker run --rm -p 80:80 \
    -v aniquiz_letsencrypt_certs:/etc/letsencrypt \
    -v aniquiz_certbot_webroot:/var/www/certbot \
    certbot/certbot certonly --standalone \
    -d "${DOMAIN}" -d "www.${DOMAIN}" --email "${EMAIL}" --agree-tos --non-interactive
else
  log "Certificate already present for ${DOMAIN} — skipping"
fi

# ─── 9. Start the stack ───────────────────────────────────────────────────────
log "Pulling and starting the stack"
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d

log "Done! Check: https://${DOMAIN}"
echo "Health:   curl -k https://localhost/health"
echo "Status:   docker compose -f ${APP_DIR}/docker-compose.prod.yml ps"
echo "Logs:     docker compose -f ${APP_DIR}/docker-compose.prod.yml logs -f app"
