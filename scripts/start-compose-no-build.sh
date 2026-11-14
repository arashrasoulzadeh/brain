#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:?COMPOSE_FILE is required for deployment}"
compose_project="${COMPOSE_PROJECT_NAME:-brain}"

docker compose -p "$compose_project" -f "$compose_file" up -d --no-build
