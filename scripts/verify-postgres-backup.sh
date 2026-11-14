#!/usr/bin/env sh
set -eu

backup_input="${BACKUP_INPUT:?BACKUP_INPUT is required for verification}"

if [ ! -s "$backup_input" ]; then
  echo "PostgreSQL backup is missing or empty" >&2
  exit 1
fi

pg_restore \
  --list \
  --no-password \
  "$backup_input" >/dev/null
