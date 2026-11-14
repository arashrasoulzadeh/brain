#!/usr/bin/env sh
set -eu

last_verified="${LAST_VERIFIED_BACKUP_EPOCH:?LAST_VERIFIED_BACKUP_EPOCH is required}"
target_seconds="${DATA_LOSS_TARGET_SECONDS:-86400}"
now="${CURRENT_TIME_EPOCH:-$(date +%s)}"

age=$((now - last_verified))
if [ "$age" -gt "$target_seconds" ]; then
  echo "backup exceeds the data-loss target: age=${age}s target=${target_seconds}s" >&2
  exit 1
fi
