#!/usr/bin/env sh
set -eu

backup_input="${BACKUP_INPUT:?BACKUP_INPUT is required for restore}"
restore_dsn="${RESTORE_DSN:?RESTORE_DSN is required for restore}"

if [ ! -s "$backup_input" ]; then
  echo "PostgreSQL backup is missing or empty" >&2
  exit 1
fi

pg_restore \
  --dbname="$restore_dsn" \
  --exit-on-error \
  --if-exists \
  --no-owner \
  --no-privileges \
  --no-password \
  --single-transaction \
  --clean \
  "$backup_input"
