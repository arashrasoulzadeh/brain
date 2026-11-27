#!/usr/bin/env bash
set -euo pipefail

# Local development only. Production credentials must come from a secret manager.
export HTTP_ADDRESS="${HTTP_ADDRESS:-127.0.0.1:18082}"
export OPERATOR_USERNAME="${OPERATOR_USERNAME:-admin}"
export OPERATOR_SECRET="${OPERATOR_SECRET:-secret}"
export PROJECTS_FILE="${PROJECTS_FILE:-.local/brain.sqlite}"

cd "$(dirname "${BASH_SOURCE[0]}")/.."

die() { printf '%s\n' "$*" >&2; exit 1; }

mode="${BRAIN_DEV_RUNTIME:-auto}"
os="$(uname -s)"
case "$mode" in
  auto)
    if [ "$os" = Linux ]; then mode=native; else mode=docker; fi
    ;;
  native|docker) ;;
  *) die "BRAIN_DEV_RUNTIME must be auto, native or docker." ;;
esac

if [ "$mode" = native ]; then
  [ "$os" = Linux ] || die "Native sandboxing requires Linux. Use make dev server BRAIN_DEV_RUNTIME=docker."
  if ! command -v bwrap >/dev/null 2>&1; then
    cat >&2 <<'HELP'
Bubblewrap (bwrap) is required for the Linux worker sandbox.
Install it using your distribution's package manager:
  Debian/Ubuntu: sudo apt install bubblewrap
  Fedora:        sudo dnf install bubblewrap
  Arch:          sudo pacman -S bubblewrap
Then run make dev server again.
Alternatively, use a local Docker engine: make dev server BRAIN_DEV_RUNTIME=docker
HELP
    exit 1
  fi
  command -v go >/dev/null 2>&1 || die "Go is required for native development."
  command -v node >/dev/null 2>&1 || die "Node.js 20+ is required for native development."
  [ -f node_modules/tsx/dist/loader.mjs ] || die "Install dependencies first: make setup"
  exec go run ./cmd/brain
fi

# Bubblewrap uses Linux namespaces. Docker Desktop/Colima provides Linux on
# macOS; do not silently substitute an unsandboxed worker.
command -v docker >/dev/null 2>&1 || die "A local Linux Docker engine is required on this platform. Install/start Docker Desktop or Colima, then run make dev server again."
endpoint="${DOCKER_HOST:-}"
if [ -z "$endpoint" ] || [ -n "${DOCKER_CONTEXT:-}" ]; then
  endpoint="$(docker context inspect --format '{{.Endpoints.docker.Host}}')" || die "Cannot determine the local Docker context."
fi
case "$endpoint" in
  unix://*) ;;
  *) die "Local development requires a local Unix-socket Docker engine, not a remote Docker context." ;;
esac
docker info --format '{{.OSType}}' | grep -qx linux || die "Start a local Linux Docker engine and retry make dev server."
case "$HTTP_ADDRESS" in
  127.0.0.1:*) port="${HTTP_ADDRESS#127.0.0.1:}" ;;
  *) die "Docker development requires HTTP_ADDRESS=127.0.0.1:<port>." ;;
esac
[[ "$port" =~ ^[0-9]{1,5}$ ]] && (( 10#$port >= 1 && 10#$port <= 65535 )) || die "Invalid development HTTP port."
port="$((10#$port))"
[ -z "${RUNTIME_DSN:-}" ] && [ -z "${MIGRATION_DSN:-}" ] || die "PostgreSQL application storage is not wired; refusing implicit SQLite fallback."
# Only the dedicated local data directory is writable in the container.
# Preserve the same SQLite file when switching between development runtimes.
[ ! -L .local ] || die ".local must not be a symlink."
(umask 077; mkdir -p .local)
data_dir="$(cd .local && pwd -P)"
db_dir="$(cd "$(dirname "$PROJECTS_FILE")" && pwd -P)" || die "Docker PROJECTS_FILE must be inside .local."
db_name="$(basename "$PROJECTS_FILE")"
[ "$db_dir" = "$data_dir" ] && [[ "$db_name" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]] || die "Docker PROJECTS_FILE must name a database directly inside .local."
[ ! -L "$data_dir/$db_name" ] || die "The development SQLite file must not be a symlink."
[[ "$data_dir" != *,* ]] || die "Docker development paths cannot contain commas."
uid="$(id -u)"
gid="$(id -g)"
[ "$uid" != 0 ] || die "Run the development server as a non-root user."
image=brain:local-dev
printf '%s\n' "Building the Linux development server (sandbox stays enabled)." >&2
docker build --tag "$image" .
printf '%s\n' "Starting the development server at http://127.0.0.1:$port" >&2
# Allow nested namespaces, but do not use --privileged, expose the Docker
# socket, mount the source tree, or publish beyond loopback. The server still
# runs the production bwrap startup probe and fails closed if unsupported.
exec docker run --rm --init --user "$uid:$gid" \
  --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
  --publish "127.0.0.1:$port:8080" \
  --mount "type=bind,source=$data_dir,target=/workspace/.local" \
  --env HTTP_ADDRESS=:8080 --env OPERATOR_USERNAME --env OPERATOR_SECRET \
  --env "PROJECTS_FILE=/workspace/.local/$db_name" \
  "$image"
