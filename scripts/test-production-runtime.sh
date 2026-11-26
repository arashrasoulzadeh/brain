#!/usr/bin/env bash
set -euo pipefail
image="${BRAIN_TEST_IMAGE:-brain:production-smoke}"
docker build -t "$image" .
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
container=$(docker run -d --privileged --network none -e HTTP_ADDRESS=:8080 -e OPERATOR_USERNAME=ci-operator -e OPERATOR_SECRET=ci-fixture-only -e PROJECTS_FILE=/workspace/.local/brain.sqlite "$image")
cleanup() {
  docker logs "$container" 2>&1 || true
  docker rm -f "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT
docker exec -i "$container" node --input-type=module - <<'NODE'
import assert from "node:assert/strict";
import { setTimeout as delay } from "node:timers/promises";
const base = "http://127.0.0.1:8080";
let ready = false;
for (let attempt = 0; attempt < 30; attempt++) {
  try {
    ready = (await fetch(base + "/readyz", { signal: AbortSignal.timeout(1000) })).ok;
    if (ready) break;
  } catch {}
  await delay(1000);
}
assert.ok(ready, "packaged host never became ready");
const login = await fetch(base + "/admin/session", {method:"POST", headers:{"content-type":"application/json"}, body:JSON.stringify({username:"ci-operator",secret:"ci-fixture-only"})});
assert.ok(login.ok);
const { session } = await login.json();
const headers = {authorization:"Bearer " + session, "content-type":"application/json"};
async function action(operation, extra = {}) {
  const response = await fetch(base + "/admin/actions", {method:"POST", headers, body:JSON.stringify({operation,project_id:"ci-project",plugin_id:"tables",version:"0.1.0",...extra}), signal:AbortSignal.timeout(20000)});
  assert.ok(response.ok, operation + ": " + await response.clone().text());
  return response.json();
}
const forbidden = await fetch(base + "/admin/actions", {method:"POST", headers, body:JSON.stringify({operation:"plugin.enable",project_id:"ci-project",plugin_id:"tables",version:"0.1.0"})});
assert.equal(forbidden.status, 403, "uninstalled plugin was enabled");
await action("project.create");
await action("plugin.install");
await action("plugin.readiness", {ready:true});
for (const capability of ["tables.read","tables.write","tables.schema"]) await action("plugin.grant", {capability});
await action("plugin.enable");
const result = await action("plugin.invoke", {target_operation:"tables.write",payload:{table:"messages",row:{id:"one"}}});
assert.equal(result.success, true);
const read = await action("plugin.invoke", {target_operation:"tables.read",payload:{table:"messages"}});
assert.equal(read.success, true);
assert.deepEqual(JSON.parse(Buffer.from(read.payload,"base64").toString("utf8")).rows,[{id:"one"}]);
assert.ok((await fetch(base + "/readyz")).ok);
NODE
