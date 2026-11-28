#!/usr/bin/env bash
set -euo pipefail
image="${BRAIN_TEST_IMAGE:-brain:production-smoke}"
coverage_dir="${BRAIN_PRODUCTION_COVERDIR:-}"
coverage_env=()
if [ -n "$coverage_dir" ]; then
  mkdir -p "$coverage_dir"
  docker build --build-arg BRAIN_COVERAGE=1 -t "$image" .
  coverage_env=(-e GOCOVERDIR=/workspace/.local/coverage)
else
  docker build -t "$image" .
fi
if output=$(docker run --rm --network none -e HTTP_ADDRESS=:8080 "$image" 2>&1); then
  echo "host accepted missing operator credentials" >&2
  exit 1
fi
printf '%s\n' "$output" | grep -F 'OPERATOR_USERNAME and OPERATOR_SECRET are required'
if output=$(docker run --rm --network none -e HTTP_ADDRESS=:8080 -e OPERATOR_USERNAME=ci-operator -e OPERATOR_SECRET=ci-fixture-only -e PROJECTS_FILE=:memory: "$image" 2>&1); then
  echo "host accepted non-durable production storage" >&2
  exit 1
fi
printf '%s\n' "$output" | grep -F 'PROJECTS_FILE must name an explicit durable SQLite file'

# Privileged namespace access is used only in this isolated CI test, never in
# production deployment. No host paths, credentials, or network are attached.
container=$(docker run -d --privileged --network none "${coverage_env[@]}" -e HTTP_ADDRESS=:8080 -e OPERATOR_USERNAME=ci-operator -e OPERATOR_SECRET=ci-fixture-only -e PROJECTS_FILE=/workspace/.local/brain.sqlite "$image")
cleanup() {
  docker logs "$container" 2>&1 || true
  docker rm -f "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT
# Do not race an immediately exiting host with docker exec.
for attempt in $(seq 1 5); do
  if [ "$(docker inspect -f '{{.State.Running}}' "$container")" != true ]; then
    echo "packaged host exited before readiness" >&2
    exit 1
  fi
  sleep 1
done
docker exec -i "$container" node --input-type=module - < scripts/test-runtime-api.mjs

# Graceful shutdown flushes the real instrumented host's execution counters.
if [ -n "$coverage_dir" ]; then
  docker stop --time 15 "$container" >/dev/null
  test "$(docker inspect -f '{{.State.ExitCode}}' "$container")" = 0
  docker cp "$container:/workspace/.local/coverage/." "$coverage_dir"
  find "$coverage_dir" -name 'covcounters.*' -print -quit | grep -q .
fi

# Exercise the real managed pool in a separate loopback-only CI container.
# Preserve coverage from the original production-mode test above. The child
# Go Neurons inherit only the trusted coverage directory, not host credentials.
docker rm -f "$container" >/dev/null
container=$(docker run -d --privileged --network none "${coverage_env[@]}" \
  -e HTTP_ADDRESS=127.0.0.1:8080 -e OPERATOR_USERNAME=ci-operator \
  -e OPERATOR_SECRET=ci-fixture-only -e PROJECTS_FILE=/workspace/.local/brain.sqlite \
  -e BRAIN_LOCAL_DEV=1 -e BRAIN_DEV_WORKERS=4 -e BRAIN_DEV_SLOTS=1 \
  -e BRAIN_DEV_PLUGIN_EXECUTION=cortex "$image")
docker exec -i "$container" node --input-type=module - < tests/development/verify-neurons.mjs
# Run the same full HTTP/storage/lifecycle parity suite on the scheduled path.
docker exec -i "$container" node --input-type=module - < scripts/test-runtime-api.mjs
docker stop --time 15 "$container" >/dev/null
test "$(docker inspect -f '{{.State.ExitCode}}' "$container")" = 0
if [ -n "$coverage_dir" ]; then
  docker cp "$container:/workspace/.local/coverage/." "$coverage_dir"
  find "$coverage_dir" -name 'covcounters.*' -print -quit | grep -q .
fi
