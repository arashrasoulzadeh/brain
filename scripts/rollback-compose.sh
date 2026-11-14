#!/usr/bin/env sh
set -eu

compose_project="${COMPOSE_PROJECT_NAME:-brain}"
rollback_file="${ROLLBACK_COMPOSE_FILE:?ROLLBACK_COMPOSE_FILE is required for rollback}"

docker compose -p "$compose_project" -f "$rollback_file" down --volumes --remove-orphans || true
docker compose -p "$compose_project" -f "$rollback_file" up -d --no-build
