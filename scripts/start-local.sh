#!/usr/bin/env bash
set -euo pipefail

# Local development only. Production credentials must come from a secret manager.
export HTTP_ADDRESS="${HTTP_ADDRESS:-127.0.0.1:18082}"
export OPERATOR_USERNAME="${OPERATOR_USERNAME:-admin}"
export OPERATOR_SECRET="${OPERATOR_SECRET:-secret}"
export PROJECTS_FILE="${PROJECTS_FILE:-.local/brain.sqlite}"

exec go run ./cmd/brain
