#!/usr/bin/env sh
set -eu

compose_file="${COMPOSE_FILE:?COMPOSE_FILE is required}"
project="${COMPOSE_PROJECT_NAME:-brain-clean-rehearsal}"
restore_project="${RESTORE_PROJECT:-brain-clean-restore}"
backup_input="${BACKUP_INPUT:?BACKUP_INPUT is required}"
restore_dsn="${RESTORE_DSN:?RESTORE_DSN is required}"
object_manifest="${OBJECT_MANIFEST:?OBJECT_MANIFEST is required}"
brain_bin="${BRAIN_BIN:-./brain}"

cleanup() {
  docker compose -p "$project" -f "$compose_file" down --volumes --remove-orphans || true
  docker compose -p "$restore_project" -f "$compose_file" down --volumes --remove-orphans || true
}
trap cleanup EXIT

docker compose -p "$project" -f "$compose_file" up -d --build
curl --fail --silent --show-error http://127.0.0.1:18080/livez
docker compose -p "$project" -f "$compose_file" down --volumes --remove-orphans

COMPOSE_FILE="$compose_file" COMPOSE_PROJECT_NAME="$project" scripts/start-compose-no-build.sh
curl --fail --silent --show-error http://127.0.0.1:18080/livez
docker compose -p "$project" -f "$compose_file" down --volumes --remove-orphans

COMPOSE_FILE="$compose_file" RESTORE_PROJECT="$restore_project" BACKUP_INPUT="$backup_input" RESTORE_DSN="$restore_dsn" OBJECT_MANIFEST="$object_manifest" BRAIN_BIN="$brain_bin" scripts/restore-compose.sh
