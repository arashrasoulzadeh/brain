#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:?COMPOSE_FILE is required for restore}"
restore_project="${RESTORE_PROJECT:?RESTORE_PROJECT is required for restore}"
case "$restore_project" in
  ''|*[!A-Za-z0-9_-]*)
  echo "RESTORE_PROJECT must contain a valid Compose project name" >&2
  exit 1
  ;;
esac

docker compose -p "$restore_project" -f "$compose_file" up -d --no-build
scripts/restore-postgres-backup.sh
scripts/verify-object-manifest.sh
