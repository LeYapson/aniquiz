#!/usr/bin/env bash
# backup-db.sh — PostgreSQL dump with compression and retention policy
# Usage: ./scripts/backup-db.sh
# Schedule: 0 3 * * * /opt/aniquiz/scripts/backup-db.sh
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/aniquiz/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
DB_CONTAINER="${DB_CONTAINER:-aniquiz-db-1}"
POSTGRES_USER="${POSTGRES_USER:-aniquiz}"
POSTGRES_DB="${POSTGRES_DB:-aniquiz}"

TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_FILE="${BACKUP_DIR}/aniquiz_${TIMESTAMP}.sql.gz"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
info() { log "INFO  $*"; }
fail() { log "ERROR $*" >&2; exit 1; }

mkdir -p "${BACKUP_DIR}"

# ─── Dump ─────────────────────────────────────────────────────────────────────
info "Dumping ${POSTGRES_DB} to ${BACKUP_FILE}..."
docker exec "${DB_CONTAINER}" \
    pg_dump -U "${POSTGRES_USER}" "${POSTGRES_DB}" \
    | gzip -9 > "${BACKUP_FILE}"

# Verify the backup is not empty
BACKUP_SIZE=$(stat -c%s "${BACKUP_FILE}" 2>/dev/null || stat -f%z "${BACKUP_FILE}")
if [ "${BACKUP_SIZE}" -lt 100 ]; then
    rm -f "${BACKUP_FILE}"
    fail "Backup file is suspiciously small (${BACKUP_SIZE} bytes) — aborting"
fi

info "Backup complete: ${BACKUP_FILE} ($(numfmt --to=iec "${BACKUP_SIZE}" 2>/dev/null || echo "${BACKUP_SIZE} bytes"))"

# ─── Retention cleanup ────────────────────────────────────────────────────────
DELETED=$(find "${BACKUP_DIR}" -name "aniquiz_*.sql.gz" -mtime "+${RETENTION_DAYS}" -print -delete | wc -l)
[ "${DELETED}" -gt 0 ] && info "Removed ${DELETED} backup(s) older than ${RETENTION_DAYS} days"

# ─── Summary ─────────────────────────────────────────────────────────────────
TOTAL=$(find "${BACKUP_DIR}" -name "aniquiz_*.sql.gz" | wc -l)
info "Total backups retained: ${TOTAL}"
