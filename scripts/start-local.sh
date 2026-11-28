#!/usr/bin/env bash
set -euo pipefail

# Local development only. Production credentials must come from a secret manager.
export HTTP_ADDRESS="${HTTP_ADDRESS:-127.0.0.1:18082}"
export OPERATOR_USERNAME="${OPERATOR_USERNAME:-admin}"
export OPERATOR_SECRET="${OPERATOR_SECRET:-secret}"
export PROJECTS_FILE="${PROJECTS_FILE:-.local/brain.sqlite}"

cd "$(dirname "${BASH_SOURCE[0]}")/.."
case "$(uname -s)" in
  Darwin)
    case "$HTTP_ADDRESS" in
      127.0.0.1:*) ;;
      *) echo "macOS development requires a loopback HTTP_ADDRESS (127.0.0.1:<port>)." >&2; exit 1 ;;
    esac
    [ -x /usr/bin/sandbox-exec ] || { echo "macOS Seatbelt sandbox is unavailable; refusing to run unsandboxed." >&2; exit 1; }
    export BRAIN_LOCAL_MACOS=1
    ;;
  Linux)
    command -v bwrap >/dev/null 2>&1 || { echo "Install bubblewrap (for Debian/Ubuntu: sudo apt install bubblewrap), then retry." >&2; exit 1; }
    ;;
  *) echo "Native development supports macOS and Linux." >&2; exit 1 ;;
esac
command -v go >/dev/null 2>&1 || { echo "Install Go first." >&2; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Install Node.js 20+ first." >&2; exit 1; }
[ -f node_modules/tsx/dist/loader.mjs ] || { echo "Install dependencies first: make setup" >&2; exit 1; }

if [[ "${BRAIN_LOCAL_DEV:-}" == 1 ]]; then
  mkdir -p .local
  go build -o .local/brain-dev ./cmd/brain
  exec .local/brain-dev
fi

exec go run ./cmd/brain
