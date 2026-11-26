#!/usr/bin/env bash
set -euo pipefail

bwrap_bin="${BWRAP_BIN:-bwrap}"
node_bin="${NODE_BIN:-node}"
worker_script="${WORKER_SCRIPT:?WORKER_SCRIPT is required}"
descriptor="${EXECUTION_DESCRIPTOR:?EXECUTION_DESCRIPTOR is required}"
plugin_root="${PLUGIN_ROOT:?PLUGIN_ROOT is required}"
cpu_millis="${CPU_MILLIS:?CPU_MILLIS is required}"
memory_bytes="${MEMORY_BYTES:?MEMORY_BYTES is required}"
process_limit="${PROCESS_LIMIT:?PROCESS_LIMIT is required}"
runtime_root="${WORKER_RUNTIME_ROOT:-}"
ready="${BRAIN_WORKER_READY:-0}"
source_files="${WORKER_SOURCE_FILES:-}"
require_sources="${REQUIRE_DECLARED_SOURCES:-0}"

if ! command -v "$bwrap_bin" >/dev/null 2>&1; then
  echo "sandbox launcher requires bubblewrap (set BWRAP_BIN)" >&2
  exit 78
fi
if [ ! -f "$worker_script" ] || [ ! -f "$descriptor" ] || [ ! -d "$plugin_root" ]; then
  echo "sandbox launcher received an invalid worker path" >&2
  exit 64
fi
node_path=$(readlink -f "$(command -v "$node_bin")")
worker_script=$(realpath "$worker_script")
descriptor=$(realpath "$descriptor")
plugin_root=$(realpath "$plugin_root")
if [ ! -x "$node_path" ]; then
  echo "sandbox launcher received an invalid Node executable" >&2
  exit 64
fi
for limit in "$cpu_millis" "$memory_bytes" "$process_limit"; do
  if ! [[ "$limit" =~ ^[1-9][0-9]*$ ]]; then
    echo "sandbox limits must be positive integers" >&2
    exit 64
  fi
done

mounts=()
for dir in /usr/lib /usr/lib64 /lib /lib64 /usr/share/nodejs; do
  if [ -d "$dir" ]; then mounts+=(--ro-bind "$dir" "$dir"); fi
done
# Bind only the descriptor, worker script, plugin directory, and trusted runtime
# dependencies. Never bind the descriptor directory (it may contain a database).
mounts+=(--ro-bind "$node_path" "$node_path" --ro-bind "$worker_script" "$worker_script")
if [ -n "$source_files" ]; then
  mounts+=(--tmpfs "$plugin_root")
  while IFS= read -r source; do
    if [ -z "$source" ] || [ ! -f "$plugin_root/$source" ]; then
      echo "sandbox source is unavailable" >&2
      exit 64
    fi
    source_path=$(realpath "$plugin_root/$source")
    if [ "$source_path" != "$plugin_root/$source" ]; then
      echo "sandbox source must be a canonical path inside the plugin" >&2
      exit 64
    fi
    mounts+=(--ro-bind "$source_path" "$source_path")
  done <<< "$source_files"
elif [ "$require_sources" = 1 ]; then
  echo "sandbox requires manifest-declared source files" >&2
  exit 78
else
  # Only the standalone fixture launcher uses a whole plugin directory.
  mounts+=(--ro-bind "$plugin_root" "$plugin_root")
fi
mounts+=(--ro-bind "$descriptor" "$descriptor")
if [ -n "$source_files" ]; then mounts+=(--remount-ro "$plugin_root"); fi
node_args=(--max-old-space-size=128)
if [ -n "$runtime_root" ]; then
  runtime_root=$(realpath "$runtime_root")
  loader="$runtime_root/node_modules/tsx/dist/loader.mjs"
  if [ ! -f "$loader" ] || [ ! -d "$runtime_root/plugins/sdk-typescript" ] || [ ! -f "$runtime_root/package.json" ]; then
    echo "sandbox TypeScript runtime dependencies unavailable" >&2
    exit 78
  fi
  mounts+=(--ro-bind "$runtime_root/node_modules" "$runtime_root/node_modules"
    --ro-bind "$runtime_root/plugins/sdk-typescript" "$runtime_root/plugins/sdk-typescript"
    --ro-bind "$runtime_root/package.json" "$runtime_root/package.json")
  if [ -f "$runtime_root/tsconfig.json" ]; then
    mounts+=(--ro-bind "$runtime_root/tsconfig.json" "$runtime_root/tsconfig.json")
  fi
  node_args+=(--import "$loader")
fi

ulimit -t "$(( (cpu_millis + 999) / 1000 ))"
ulimit -v "$(( (memory_bytes + 1023) / 1024 ))"
ulimit -u "$process_limit"

exec "$bwrap_bin" --die-with-parent --new-session --unshare-all --cap-drop ALL \
  --tmpfs /tmp \
  "${mounts[@]}" \
  --proc /proc --dev /dev --chdir "$plugin_root" \
  --clearenv --setenv PATH "$(dirname "$node_path")" \
  --setenv HOME /tmp --setenv TMPDIR /tmp --setenv BRAIN_WORKER_READY "$ready" \
  --setenv GOMAXPROCS 2 --setenv UV_THREADPOOL_SIZE 2 \
  "$node_path" "${node_args[@]}" "$worker_script" "$descriptor"
