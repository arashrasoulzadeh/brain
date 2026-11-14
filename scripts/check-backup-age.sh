#!/usr/bin/env sh
set -eu

last_verified="${LAST_VERIFIED_BACKUP_EPOCH:?LAST_VERIFIED_BACKUP_EPOCH is required}"
alert_after="${BACKUP_ALERT_AFTER_SECONDS:-64800}"
now="${CURRENT_TIME_EPOCH:-$(date +%s)}"

age=$((now - last_verified))
if [ "$age" -gt "$alert_after" ]; then
  echo "verified backup is stale: age=${age}s threshold=${alert_after}s" >&2
  exit 1
fi
