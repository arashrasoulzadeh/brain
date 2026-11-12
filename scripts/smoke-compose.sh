#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:-compose.test.yml}"
project="${COMPOSE_PROJECT_NAME:-brain-smoke}"

cleanup() {
  docker compose -p "$project" -f "$compose_file" down --volumes --remove-orphans
}
trap cleanup EXIT

docker compose -p "$project" -f "$compose_file" up -d --build
curl --fail --silent --show-error http://127.0.0.1:18080/livez
