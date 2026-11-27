#!/bin/bash
set -euo pipefail

[ "$(uname -s)" = Darwin ] || exit 78
node="${NODE_BIN:?NODE_BIN is required}"
profile="${BRAIN_MACOS_PROFILE:?BRAIN_MACOS_PROFILE is required}"
loader="${WORKER_RUNTIME_ROOT:?WORKER_RUNTIME_ROOT is required}/node_modules/tsx/dist/loader.mjs"
worker="${WORKER_SCRIPT:?WORKER_SCRIPT is required}"
descriptor="${EXECUTION_DESCRIPTOR:?EXECUTION_DESCRIPTOR is required}"
cd "${PLUGIN_ROOT:?PLUGIN_ROOT is required}"
ulimit -t "$(( (CPU_MILLIS + 999) / 1000 ))"
# macOS does not implement Linux's reliable per-process address-space limit.
# Keep the V8 heap cap and Go's existing invocation deadline.
exec /usr/bin/env -i PATH=/usr/bin:/bin HOME=/tmp TMPDIR=/tmp \
  TSX_DISABLE_CACHE=1 BRAIN_WORKER_READY=1 GOMAXPROCS=2 UV_THREADPOOL_SIZE=2 \
  /usr/bin/sandbox-exec -p "$profile" "$node" --max-old-space-size=128 \
  --disable-wasm-trap-handler --import "$loader" "$worker" "$descriptor"
