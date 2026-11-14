#!/usr/bin/env sh
set -eu

postgres_dsn="${POSTGRES_DSN:?POSTGRES_DSN is required for deployment}"

psql "$postgres_dsn" \
  --no-password \
  --no-psqlrc \
  --set=ON_ERROR_STOP=1 \
  --tuples-only \
  --command='SELECT 1' >/dev/null
