#!/usr/bin/env sh
set -eu

last_verified="${LAST_VERIFIED_BACKUP_EPOCH:?LAST_VERIFIED_BACKUP_EPOCH is required}"
alert_after="${BACKUP_ALERT_AFTER_SECONDS:-64800}"
now="${CURRENT_TIME_EPOCH:-$(date +%s)}"

case "$last_verified" in
  ''|*[!0-9]*) echo "LAST_VERIFIED_BACKUP_EPOCH must be an integer" >&2; exit 1 ;;
esac
case "$alert_after" in
  ''|*[!0-9]*) echo "BACKUP_ALERT_AFTER_SECONDS must be an integer" >&2; exit 1 ;;
esac
case "$now" in
  ''|*[!0-9]*) echo "CURRENT_TIME_EPOCH must be an integer" >&2; exit 1 ;;
esac

age=$((now - last_verified))
if [ "$age" -lt 0 ]; then
  echo "verified backup timestamp is in the future" >&2
  exit 1
fi
if [ "$age" -gt "$alert_after" ]; then
  echo "verified backup is stale: age=${age}s threshold=${alert_after}s" >&2
  exit 1
fi
