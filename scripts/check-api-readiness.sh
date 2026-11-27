#!/usr/bin/env sh
set -eu

readiness_url="${READINESS_URL:-http://127.0.0.1:8080/_system/readyz}"
readiness_attempts="${READINESS_ATTEMPTS:-30}"
readiness_delay_seconds="${READINESS_DELAY_SECONDS:-2}"

attempt=1
while [ "$attempt" -le "$readiness_attempts" ]; do
  if curl --fail --silent --show-error --max-time 5 "$readiness_url" >/dev/null; then
    exit 0
  fi
  if [ "$attempt" -lt "$readiness_attempts" ]; then
    sleep "$readiness_delay_seconds"
  fi
  attempt=$((attempt + 1))
done

echo "API readiness check failed after $readiness_attempts attempts" >&2
exit 1
