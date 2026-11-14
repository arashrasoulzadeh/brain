#!/usr/bin/env sh
set -eu

postgres_dsn="${POSTGRES_DSN:?POSTGRES_DSN is required for backup}"
backup_output="${BACKUP_OUTPUT:?BACKUP_OUTPUT is required for backup}"

umask 077
pg_dump \
  --dbname="$postgres_dsn" \
  --format=custom \
  --file="$backup_output" \
  --no-owner \
  --no-privileges \
  --serializable-deferrable \
  --no-password
