#!/bin/bash

set -euo pipefail

BACKUP_FILE="${1:-}"
BUCKET="${BACKUP_BUCKET:-learnflow-backups}"

if [[ -z "$BACKUP_FILE" ]]; then
    echo "Usage: restore.sh <filename>"
    echo ""
    echo "Available backups:"
    mc ls "local/${BUCKET}"
    exit 1
fi

echo "[restore] $(date): starting ← ${BACKUP_FILE}"
echo "[restore] WARNING: ensure the application is stopped before restoring to avoid conflicts"

mc cat "local/${BUCKET}/${BACKUP_FILE}" \
    | gunzip \
    | PGPASSWORD="${POSTGRES_PASSWORD:-postgres}" psql \
        --single-transaction \
        -v ON_ERROR_STOP=1 \
        -h "${POSTGRES_HOST:-postgres}" \
        -p "${POSTGRES_PORT:-5432}" \
        -U "${POSTGRES_USER:-postgres}" \
        -d "${POSTGRES_DB:-learnflow}"

echo "[restore] $(date): done. Database fully restored from ${BACKUP_FILE}"
