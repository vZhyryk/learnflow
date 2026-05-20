#!/bin/bash

set -euo pipefail


BUCKET="${BACKUP_BUCKET:-learnflow-backups}"
SCHEDULE="${BACKUP_SCHEDULE:-0 */6 * * *}"
echo "[backup] configuring mc alias → ${MINIO_ENDPOINT:-http://minio:9000}"
mc alias set local \
    "${MINIO_ENDPOINT:-http://minio:9000}" \
    "${MINIO_ACCESS_KEY:-minioadmin}" \
    "${MINIO_SECRET_KEY:-minioadmin}"

echo "[backup] ensuring bucket: ${BUCKET}"
mc mb --ignore-existing "local/${BUCKET}"

cat > /usr/local/bin/backup-cron.sh << EOF
#!/bin/bash
export POSTGRES_HOST="${POSTGRES_HOST:-postgres}"
export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
export POSTGRES_DB="${POSTGRES_DB:-learnflow}"
export POSTGRES_USER="${POSTGRES_USER:-postgres}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
export MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://minio:9000}"
export MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minioadmin}"
export MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minioadmin}"
export BACKUP_BUCKET="${BUCKET}"
/usr/local/bin/backup.sh
EOF
chmod +x /usr/local/bin/backup-cron.sh

echo "${SCHEDULE} /usr/local/bin/backup-cron.sh >> /var/log/backup.log 2>&1" > /etc/crontabs/root

echo "[backup] schedule: ${SCHEDULE}"

echo "[backup] running initial backup..."
/usr/local/bin/backup.sh

echo "[backup] starting crond..."
exec crond -f -l 6