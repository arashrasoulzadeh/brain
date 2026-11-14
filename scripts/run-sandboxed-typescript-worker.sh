#!/bin/sh
set -eu

bwrap_bin="${BWRAP_BIN:-bwrap}"
node_bin="${NODE_BIN:-node}"
worker_script="${WORKER_SCRIPT:?WORKER_SCRIPT is required}"
descriptor="${EXECUTION_DESCRIPTOR:?EXECUTION_DESCRIPTOR is required}"
plugin_root="${PLUGIN_ROOT:?PLUGIN_ROOT is required}"

if ! command -v "$bwrap_bin" >/dev/null 2>&1; then
  echo "sandbox launcher requires bubblewrap (set BWRAP_BIN)" >&2
  exit 78
fi
if [ ! -f "$worker_script" ] || [ ! -f "$descriptor" ] || [ ! -d "$plugin_root" ]; then
  echo "sandbox launcher received an invalid worker path" >&2
  exit 64
fi

node_path=$(command -v "$node_bin")
node_dir=$(dirname "$node_path")
worker_dir=$(dirname "$worker_script")
descriptor_dir=$(dirname "$descriptor")

exec "$bwrap_bin" \
  --die-with-parent \
  --unshare-net \
  --ro-bind "$node_path" "$node_path" \
  --ro-bind "$node_dir" "$node_dir" \
  --ro-bind "$worker_dir" "$worker_dir" \
  --ro-bind "$descriptor_dir" "$descriptor_dir" \
  --ro-bind "$plugin_root" "$plugin_root" \
  --proc /proc \
  --dev /dev \
  --tmpfs /tmp \
  --chdir "$plugin_root" \
  --clearenv \
  --setenv PATH "$node_dir" \
  "$node_path" "$worker_script" "$descriptor"
