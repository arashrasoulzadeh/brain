#!/usr/bin/env bash
set -euo pipefail

root=$(mktemp -d)
server_pid=""
cleanup() {
  if [ -n "$server_pid" ]; then kill "$server_pid" 2>/dev/null || true; fi
  rm -rf "$root"
}
trap cleanup EXIT
mkdir "$root/plugin"
printf 'outside secret' > "$root/host-secret"
printf 'undeclared secret' > "$root/plugin/undeclared"
node_bin="${NODE_BIN:-node}"
"$node_bin" -e 'const fs=require("node:fs");require("node:http").createServer((req,res)=>res.end("reachable")).listen(0,"127.0.0.1",function(){fs.writeFileSync(process.argv[1],String(this.address().port))})' "$root/port" &
server_pid=$!
for attempt in $(seq 1 50); do
  [ -s "$root/port" ] && break
  sleep 0.1
done
port=$(cat "$root/port")
cat > "$root/plugin/index.mjs" <<EOF
import { readFile, writeFile } from "node:fs/promises";
export async function handle() {
  let networkBlocked = false;
  let filesystemBlocked = false;
  let hostSecretBlocked = false;
  let undeclaredBlocked = false;
  try { const response = await fetch("http://127.0.0.1:$port", { signal: AbortSignal.timeout(1000) }); await response.text(); } catch { networkBlocked = true; }
  try { await writeFile(new URL("./should-not-write", import.meta.url), "blocked"); } catch { filesystemBlocked = true; }
  try { await readFile(new URL("../host-secret", import.meta.url)); } catch { hostSecretBlocked = true; }
  try { await readFile(new URL("./undeclared", import.meta.url)); } catch { undeclaredBlocked = true; }
  return { networkBlocked, filesystemBlocked, hostSecretBlocked, undeclaredBlocked, envBlocked: !process.env.BRAIN_TEST_SECRET };
}
EOF
cat > "$root/execution.json" <<EOF
{"entrypoint":"$root/plugin/index.mjs","pluginId":"sample","version":"1.0.0","projectId":"project","routes":[{"operation":"health","function":"handle"}]}
EOF
request='{"jsonrpc":"2.0","id":"one","params":{"protocol":1,"pluginId":"sample","version":"1.0.0","projectId":"project","operation":"health"}}'
# Positive controls: all probes are accessible without the sandbox.
control=$(printf '%s\n' "$request" | BRAIN_TEST_SECRET=fixture "$node_bin" scripts/typescript-worker.mjs "$root/execution.json")
for field in networkBlocked filesystemBlocked hostSecretBlocked undeclaredBlocked envBlocked; do
  printf '%s\n' "$control" | grep -F "\"$field\":false"
done
test -f "$root/plugin/should-not-write"
rm "$root/plugin/should-not-write"
response=$(printf '%s\n' "$request" | BRAIN_TEST_SECRET=fixture WORKER_SCRIPT="$PWD/scripts/typescript-worker.mjs" EXECUTION_DESCRIPTOR="$root/execution.json" PLUGIN_ROOT="$root/plugin" WORKER_SOURCE_FILES=index.mjs REQUIRE_DECLARED_SOURCES=1 CPU_MILLIS=1000 MEMORY_BYTES=2147483648 PROCESS_LIMIT=16 bash scripts/run-sandboxed-typescript-worker.sh)
for field in networkBlocked filesystemBlocked hostSecretBlocked undeclaredBlocked envBlocked; do
  printf '%s\n' "$response" | grep -F "\"$field\":true"
done
test ! -e "$root/plugin/should-not-write"
