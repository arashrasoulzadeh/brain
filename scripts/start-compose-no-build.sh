#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:?COMPOSE_FILE is required for deployment}"
compose_project="${COMPOSE_PROJECT_NAME:-brain}"
case "$compose_project" in
  ''|*[!A-Za-z0-9_-]*)
  echo "COMPOSE_PROJECT_NAME must contain a valid Compose project name" >&2
  exit 1
  ;;
esac

docker compose -p "$compose_project" -f "$compose_file" up -d --no-build
