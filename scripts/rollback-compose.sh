#!/usr/bin/env sh
set -eu

compose_project="${COMPOSE_PROJECT_NAME:-brain}"
rollback_file="${ROLLBACK_COMPOSE_FILE:?ROLLBACK_COMPOSE_FILE is required for rollback}"
case "$compose_project" in
  ''|*[!A-Za-z0-9_-]*)
  echo "COMPOSE_PROJECT_NAME must contain a valid Compose project name" >&2
  exit 1
  ;;
esac

docker compose -p "$compose_project" -f "$rollback_file" down --volumes --remove-orphans || true
docker compose -p "$compose_project" -f "$rollback_file" up -d --no-build
