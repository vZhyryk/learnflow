#!/bin/bash

set -euo pipefail

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
FILENAME="${POSTGRES_DB:-learnflow}_${TIMESTAMP}.sql.gz"
BUCKET="${BACKUP_BUCKET:-learnflow-backups}"

echo "[backup] $(date): starting → ${FILENAME}"

PGPASSWORD="${POSTGRES_PASSWORD:-postgres}" pg_dump \
    -h "${POSTGRES_HOST:-postgres}" \
    -p "${POSTGRES_PORT:-5432}" \
    -U "${POSTGRES_USER:-postgres}" \
    -d "${POSTGRES_DB:-learnflow}" \
    --no-owner \
    --no-acl \
    | gzip \
    | mc pipe "local/${BUCKET}/${FILENAME}"

echo "[backup] $(date): done → ${FILENAME}"
