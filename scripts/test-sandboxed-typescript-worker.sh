#!/usr/bin/env bash
set -eu

root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT
cat > "$root/index.mjs" <<'EOF'
import { writeFile } from "node:fs/promises";
export async function handle() {
  let networkBlocked = false;
  let filesystemBlocked = false;
  try { await fetch("http://127.0.0.1:9"); } catch { networkBlocked = true; }
  try { await writeFile("/plugin/should-not-write", "blocked"); } catch { filesystemBlocked = true; }
  return { accepted: networkBlocked && filesystemBlocked };
}
EOF
cat > "$root/execution.json" <<EOF
{"entrypoint":"$root/index.mjs","pluginId":"sample","version":"1.0.0","projectId":"project","operations":["health"]}
EOF
response=$(WORKER_SCRIPT="$PWD/scripts/typescript-worker.mjs" EXECUTION_DESCRIPTOR="$root/execution.json" PLUGIN_ROOT="$root" CPU_MILLIS=1000 MEMORY_BYTES=134217728 PROCESS_LIMIT=16 bash scripts/run-sandboxed-typescript-worker.sh <<'EOF'
{"jsonrpc":"2.0","id":"one","params":{"pluginId":"sample","version":"1.0.0","projectId":"project","operation":"health"}}
EOF
)
printf '%s\n' "$response" | rg '"accepted":true'
