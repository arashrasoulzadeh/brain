import assert from "node:assert/strict";
import { setTimeout as delay } from "node:timers/promises";
const base = process.env.BRAIN_TEST_ORIGIN ?? "http://127.0.0.1:8080";
let ready = false;
for (let attempt = 0; attempt < 30; attempt++) {
  try {
    ready = (await fetch(base + "/_system/readyz", { signal: AbortSignal.timeout(1000) })).ok;
    if (ready) break;
  } catch {}
  await delay(1000);
}
assert.ok(ready, "packaged host never became ready");
const login = await fetch(base + "/_system/admin/session", {method:"POST", headers:{"content-type":"application/json"}, body:JSON.stringify({username:"ci-operator",secret:"ci-fixture-only"})});
assert.ok(login.ok);
const { session } = await login.json();
const headers = {authorization:"Bearer " + session, "content-type":"application/json"};
async function action(operation, extra = {}) {
  const body = {operation,project_id:"ci-project",plugin_id:"tables",version:"0.1.0",...extra};
  const namespace = operation === "project.create" || operation === "plugin.readiness" ? "_system" : encodeURIComponent(body.project_id);
  const response = await fetch(base + "/" + namespace + "/admin/actions", {method:"POST", headers, body:JSON.stringify(body), signal:AbortSignal.timeout(20000)});
  assert.ok(response.ok, operation + " (" + (extra.plugin_id ?? "tables") + "): " + await response.clone().text());
  return response.json();
}
const forbidden = await fetch(base + "/_system/admin/actions", {method:"POST", headers, body:JSON.stringify({operation:"plugin.enable",project_id:"ci-project",plugin_id:"tables",version:"0.1.0"})});
assert.equal(forbidden.status, 403, "project command accepted through system namespace");
await action("project.create");
const uninstalled = await fetch(base + "/ci-project/admin/actions", {method:"POST", headers, body:JSON.stringify({operation:"plugin.enable",plugin_id:"tables",version:"0.1.0"})});
assert.equal(uninstalled.status, 403, "uninstalled plugin was enabled");
await action("plugin.install");
await action("plugin.readiness", {ready:true});
for (const capability of ["tables.read","tables.write","tables.schema","tables.update","tables.delete"]) await action("plugin.grant", {capability});
await action("plugin.enable");
const result = await action("plugin.invoke", {target_operation:"tables.write",payload:{table:"messages",row:{id:"one"}}});
assert.equal(result.Success, true);
const read = await action("plugin.invoke", {target_operation:"tables.read",payload:{table:"messages"}});
assert.equal(read.Success, true);
assert.deepEqual(JSON.parse(Buffer.from(read.Payload,"base64").toString("utf8")).rows,[{id:"one"}]);
// Exercise bounded row reads and confirmed mutations through the manifest admin boundary.
async function tableAdmin(operation, payload, readOnly = false) {
  const response = await action(readOnly ? "plugin.admin.query" : "plugin.admin.invoke", {page_id:"operations",target_operation:operation,confirmed:!readOnly,payload});
  assert.equal(response.Success,true,operation + ": " + response.ErrorCode);
  return JSON.parse(Buffer.from(response.Payload,"base64").toString("utf8"));
}
const tablePage = await tableAdmin("tables.read",{table:"messages",limit:1},true);
assert.equal(tablePage.total_rows,1);
assert.deepEqual(tablePage.rows,[{row_index:0,row:{id:"one"}}]);
assert.ok(tablePage.revision);
const changed = await tableAdmin("tables.update",{table:"messages",row_index:0,revision:tablePage.revision,row:{id:"updated"}});
assert.equal(changed.updated,true);
const staleTable = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.admin.invoke",project_id:"ci-project",plugin_id:"tables",version:"0.1.0",page_id:"operations",target_operation:"tables.delete",confirmed:true,payload:{table:"messages",row_index:0,revision:tablePage.revision}})});
assert.ok(!staleTable.ok || !(await staleTable.json()).Success,"stale row revision accepted");
const currentTable = await tableAdmin("tables.read",{table:"messages",limit:1},true);
assert.deepEqual(currentTable.rows,[{row_index:0,row:{id:"updated"}}]);
assert.equal((await tableAdmin("tables.delete",{table:"messages",row_index:0,revision:currentTable.revision})).deleted,true);
assert.equal((await tableAdmin("tables.read",{table:"messages",limit:1},true)).total_rows,0);
// Start the registered plugin inventory, not only the tables fixture.
const catalogResponse = await fetch(base + "/ci-project/admin/plugins", {headers});
assert.ok(catalogResponse.ok);
const { manifests } = await catalogResponse.json();
for (const id of ["analytics","arena","auth","dashboard","files","functions","games","invites","matchmaking","pvp","rooms","subscriptions","tables"]) {
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
assert.deepEqual(await invoke("games","games.modes",{}),{projectId:"ci-project",modes:["tic-tac-toe"]});
assert.deepEqual(await invoke("pvp","pvp.describe",{}),{projectId:"ci-project",turnBased:true,revisionedCommands:true});
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
const replay = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"arena",version:"0.1.0",target_operation:"arena.input",payload:{roomId:arenaRoom,playerId:"one",sequence:1}})});
assert.ok(!replay.ok || !(await replay.json()).Success,"replayed arena sequence accepted after worker recovery");
const match = await invoke("arena","arena.result",{roomId:arenaRoom,winnerIds:["one"],completedAt:"2026-01-01T00:00:00Z"});
await action("plugin.stop",{plugin_id:"arena",confirmed:true});
await action("plugin.recover",{plugin_id:"arena"});
await action("plugin.worker.start",{plugin_id:"arena"});
const finished = await invoke("arena","arena.snapshot",{roomId:arenaRoom});
assert.deepEqual(finished.result,match);
const matches = await action("plugin.admin.inventory",{plugin_id:"arena",page_id:"matches",limit:20});
assert.deepEqual(matches.rows,[{key:arenaRoom}]);
assert.deepEqual((await action("plugin.admin.inventory",{plugin_id:"arena",page_id:"matches",limit:20,search:"CI-ARENA"})).rows,[{key:arenaRoom}]);
assert.deepEqual((await action("plugin.admin.inventory",{plugin_id:"arena",page_id:"matches",limit:20,search:"not-present"})).rows,[]);
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
const revoked = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"auth",version:"0.1.0",target_operation:"auth.session",payload:{token:signedIn.token}})});
assert.ok(!revoked.ok || !(await revoked.json()).Success,"revoked session still admitted");
// Invitation codes remain valid across worker replacement, but only once.
const invite = await invoke("invites","invites.create",{roomId:"ci-room",targetPlayerId:"ci-member",ttlSeconds:120});
await action("plugin.stop",{plugin_id:"invites",confirmed:true});
await action("plugin.recover",{plugin_id:"invites"});
await action("plugin.worker.start",{plugin_id:"invites"});
const inviteRows = await invoke("invites","invites.inspect",{roomId:"ci-room"});
assert.equal(inviteRows[0].id,invite.id);
assert.ok(!JSON.stringify(inviteRows).includes(invite.code));
const accepted = await invoke("invites","invites.accept",{roomId:"ci-room",playerId:"ci-member",code:invite.code});
assert.equal(accepted.used,true);
assert.equal(accepted.code,"");
const repeated = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"invites",version:"0.1.0",target_operation:"invites.accept",payload:{roomId:"ci-room",playerId:"ci-member",code:invite.code}})});
assert.ok(!repeated.ok || !(await repeated.json()).Success,"accepted invitation replayed");
const channel = "rooms.connected";
const subscription = await invoke("subscriptions","subscriptions.open",{channel});
await invoke("rooms","rooms.create",{id:"ci-room"});
await invoke("rooms","rooms.join",{id:"ci-room",member:"ci-member"});
let observed = false;
let analyticsDelivered = false;
for (let attempt = 0; attempt < 30; attempt++) {
  const stream = await invoke("subscriptions","subscriptions.receive",{channel,ticket:subscription.ticket,cursor:subscription.cursor});
  observed = stream.events.some(event => event.value?.roomId === "ci-room" && event.value?.reconnected === false);
  const audit = await fetch(base + "/_system/admin/audit",{headers});
  assert.ok(audit.ok);
  analyticsDelivered = (await audit.json()).some(event => event.action === "observation_delivered" && event.project_id === "ci-project" && event.plugin_id === "analytics" && event.detail === "rooms.connected");
  if (observed && analyticsDelivered) break;
  await delay(500);
}
assert.ok(observed,"subscriptions did not receive rooms.connected");
assert.ok(analyticsDelivered,"analytics did not receive rooms.connected");
const history = await invoke("analytics","analytics.history",{name:"rooms.connected",limit:10});
assert.ok(history.rows.some(event=>event.fields.roomId==="ci-room" && event.occurred_at));
const analyticsSummary = await invoke("analytics","analytics.summary",{limit:10});
assert.ok(analyticsSummary.rows.some(event=>event.name==="rooms.connected" && event.count>=1));
// The ticket and event cursor survive replacement of the subscription worker.
await action("plugin.stop",{plugin_id:"subscriptions",confirmed:true});
await action("plugin.recover",{plugin_id:"subscriptions"});
await action("plugin.worker.start",{plugin_id:"subscriptions"});
const resumed = await invoke("subscriptions","subscriptions.receive",{channel,ticket:subscription.ticket,cursor:subscription.cursor});
assert.ok(resumed.events.some(event=>event.value.roomId==="ci-room"));
const inspected = await invoke("subscriptions","subscriptions.inspect",{channel});
assert.equal(inspected.activeTickets,1);
assert.ok(!JSON.stringify(inspected).includes(subscription.ticket));
assert.deepEqual(await invoke("subscriptions","subscriptions.revoke",{channel}),{channel,revoked:1});
const revokedTicket = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"subscriptions",version:"0.1.0",target_operation:"subscriptions.receive",payload:{channel,ticket:subscription.ticket,cursor:0}})});
assert.ok(!revokedTicket.ok || !(await revokedTicket.json()).Success,"revoked subscription ticket admitted");
const fresh = await invoke("subscriptions","subscriptions.open",{channel});
await invoke("subscriptions","subscriptions.close",{channel,ticket:fresh.ticket});
assert.ok((await fetch(base + "/_system/readyz")).ok);
