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
# Do not race an immediately exiting host with docker exec.
for attempt in $(seq 1 5); do
  if [ "$(docker inspect -f '{{.State.Running}}' "$container")" != true ]; then
    echo "packaged host exited before readiness" >&2
    exit 1
  fi
  sleep 1
done
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
  assert.ok(response.ok, operation + " (" + (extra.plugin_id ?? "tables") + "): " + await response.clone().text());
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
assert.equal(result.Success, true);
const read = await action("plugin.invoke", {target_operation:"tables.read",payload:{table:"messages"}});
assert.equal(read.Success, true);
assert.deepEqual(JSON.parse(Buffer.from(read.Payload,"base64").toString("utf8")).rows,[{id:"one"}]);
// Start the registered plugin inventory, not only the tables fixture.
const catalogResponse = await fetch(base + "/admin/plugins?project_id=ci-project", {headers});
assert.ok(catalogResponse.ok);
const { manifests } = await catalogResponse.json();
for (const id of ["analytics","arena","auth","dashboard","files","functions","invites","matchmaking","rooms","subscriptions","tables"]) {
  assert.ok(manifests.some(m => m.ID === id), "missing plugin " + id);
}
for (const manifest of manifests.filter(m => m.ID !== "tables" && m.ID !== "access-management")) {
  const identity = {plugin_id:manifest.ID,version:manifest.Version};
  await action("plugin.install", identity);
  await action("plugin.readiness", {...identity,ready:true});
  for (const capability of manifest.Capabilities ?? []) await action("plugin.grant", {...identity,capability});
  await action("plugin.enable", identity);
}
async function invoke(plugin_id, target_operation, payload) {
  const response = await action("plugin.invoke", {plugin_id,target_operation,payload});
  assert.equal(response.Success,true, target_operation + ": " + response.ErrorCode);
  return JSON.parse(Buffer.from(response.Payload,"base64").toString("utf8"));
}
// Arena state and sequence protection must survive replacement of the worker.
const arenaRoom = "ci-arena";
await invoke("arena","arena.start",{roomId:arenaRoom,players:["one","two"]});
await invoke("arena","arena.input",{roomId:arenaRoom,playerId:"one",sequence:1});
await action("plugin.stop",{plugin_id:"arena",confirmed:true});
await action("plugin.recover",{plugin_id:"arena"});
await action("plugin.worker.start",{plugin_id:"arena"});
const arenaSnapshot = await invoke("arena","arena.snapshot",{roomId:arenaRoom});
assert.deepEqual(arenaSnapshot.players,["one","two"]);
assert.equal(arenaSnapshot.state,"active");
const replay = await fetch(base+"/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"arena",version:"0.1.0",target_operation:"arena.input",payload:{roomId:arenaRoom,playerId:"one",sequence:1}})});
assert.ok(!replay.ok || !(await replay.json()).Success,"replayed arena sequence accepted after worker recovery");
const match = await invoke("arena","arena.result",{roomId:arenaRoom,winnerIds:["one"],completedAt:"2026-01-01T00:00:00Z"});
await action("plugin.stop",{plugin_id:"arena",confirmed:true});
await action("plugin.recover",{plugin_id:"arena"});
await action("plugin.worker.start",{plugin_id:"arena"});
const finished = await invoke("arena","arena.snapshot",{roomId:arenaRoom});
assert.deepEqual(finished.result,match);
const matches = await action("plugin.admin.inventory",{plugin_id:"arena",page_id:"matches",limit:20});
assert.deepEqual(matches.rows,[{key:arenaRoom}]);
// Account hashes and token digests survive worker recovery, but never appear in admin results.
await invoke("auth","auth.signup",{username:"ci-user",password:"ci-fixture-password"});
const signedIn = await invoke("auth","auth.signin",{username:"ci-user",password:"ci-fixture-password"});
await action("plugin.stop",{plugin_id:"auth",confirmed:true});
await action("plugin.recover",{plugin_id:"auth"});
await action("plugin.worker.start",{plugin_id:"auth"});
assert.deepEqual(await invoke("auth","auth.identity",{token:signedIn.token}),{userId:"ci-project:ci-user"});
const accounts = await invoke("auth","auth.admin.accounts",{});
assert.deepEqual(accounts.rows,[{username:"ci-user",userId:"ci-project:ci-user",guest:false,activeSessions:1}]);
assert.ok(!JSON.stringify(accounts).includes(signedIn.token));
assert.deepEqual(await invoke("auth","auth.admin.sessions.revoke",{username:"ci-user"}),{username:"ci-user",revoked:1});
const revoked = await fetch(base+"/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"auth",version:"0.1.0",target_operation:"auth.session",payload:{token:signedIn.token}})});
assert.ok(!revoked.ok || !(await revoked.json()).Success,"revoked session still admitted");
const channel = "rooms.connected";
const subscription = await invoke("subscriptions","subscriptions.open",{channel});
await invoke("rooms","rooms.create",{id:"ci-room"});
await invoke("rooms","rooms.join",{id:"ci-room",member:"ci-member"});
let observed = false;
let analyticsDelivered = false;
for (let attempt = 0; attempt < 30; attempt++) {
  const stream = await invoke("subscriptions","subscriptions.receive",{channel,ticket:subscription.ticket,cursor:subscription.cursor});
  observed = stream.events.some(event => event.value?.roomId === "ci-room" && event.value?.reconnected === false);
  const audit = await fetch(base + "/admin/audit",{headers});
  assert.ok(audit.ok);
  analyticsDelivered = (await audit.json()).some(event => event.action === "observation_delivered" && event.project_id === "ci-project" && event.plugin_id === "analytics" && event.detail === "rooms.connected");
  if (observed && analyticsDelivered) break;
  await delay(500);
}
assert.ok(observed,"subscriptions did not receive rooms.connected");
assert.ok(analyticsDelivered,"analytics did not receive rooms.connected");
await invoke("subscriptions","subscriptions.close",{channel,ticket:subscription.ticket});
assert.ok((await fetch(base + "/readyz")).ok);
NODE
