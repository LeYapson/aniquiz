#!/usr/bin/env bash
# deploy.sh — Zero-downtime deploy with automatic rollback
# Usage: ./scripts/deploy.sh <image-tag>
# Example: ./scripts/deploy.sh sha-abc1234
set -euo pipefail

IMAGE_TAG="${1:?Usage: $0 <image-tag>}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
IMAGE_REPO="ghcr.io/leyapson/aniquiz"
# Hit the Nginx ingress — the app's 8080 port is not published in prod.
HEALTH_URL="${HEALTH_URL:-https://localhost/health}"
CURL_OPTS="${CURL_OPTS:--fsSk}"
MAX_RETRIES=12
RETRY_INTERVAL=5

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
info() { log "INFO  $*"; }
warn() { log "WARN  $*"; }
fail() { log "ERROR $*" >&2; exit 1; }

# ─── Capture current image for rollback ──────────────────────────────────────
PREVIOUS_IMAGE=$(docker inspect --format='{{.Config.Image}}' aniquiz-app 2>/dev/null || echo "")
info "Current image: ${PREVIOUS_IMAGE:-none}"
info "Deploying:     ${IMAGE_REPO}:${IMAGE_TAG}"

# ─── Pull new image ───────────────────────────────────────────────────────────
info "Pulling ${IMAGE_REPO}:${IMAGE_TAG}..."
docker pull "${IMAGE_REPO}:${IMAGE_TAG}"

# ─── Update and restart the app service only (db stays up) ───────────────────
info "Restarting app service..."
export IMAGE_TAG
docker compose -f "${COMPOSE_FILE}" up -d --no-deps --remove-orphans app

# ─── Health check ─────────────────────────────────────────────────────────────
info "Waiting for health check at ${HEALTH_URL}..."
for i in $(seq 1 "${MAX_RETRIES}"); do
    sleep "${RETRY_INTERVAL}"
    if curl ${CURL_OPTS} "${HEALTH_URL}" > /dev/null 2>&1; then
        info "Health check passed (attempt ${i}/${MAX_RETRIES})"
        info "Deploy successful: ${IMAGE_REPO}:${IMAGE_TAG}"
        docker image prune -f --filter "label=org.opencontainers.image.title=AniQuiz" \
                                --filter "until=24h" 2>/dev/null || true
        exit 0
    fi
    warn "Health check failed (attempt ${i}/${MAX_RETRIES})"
done

# ─── Rollback ─────────────────────────────────────────────────────────────────
fail_msg="Health check failed after $((MAX_RETRIES * RETRY_INTERVAL))s"
if [ -n "${PREVIOUS_IMAGE}" ]; then
    warn "${fail_msg} — rolling back to ${PREVIOUS_IMAGE}"
    export IMAGE_TAG="${PREVIOUS_IMAGE##*:}"
    docker compose -f "${COMPOSE_FILE}" up -d --no-deps app
    sleep "${RETRY_INTERVAL}"
    if curl ${CURL_OPTS} "${HEALTH_URL}" > /dev/null 2>&1; then
        info "Rollback successful — service restored"
    else
        warn "Rollback health check also failed — manual intervention required"
    fi
fi
fail "${fail_msg}"
