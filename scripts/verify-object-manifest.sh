#!/usr/bin/env sh
set -eu

object_manifest="${OBJECT_MANIFEST:?OBJECT_MANIFEST is required for verification}"
brain_binary="${BRAIN_BIN:-./brain}"

"$brain_binary" verify-object-manifest "$object_manifest"
