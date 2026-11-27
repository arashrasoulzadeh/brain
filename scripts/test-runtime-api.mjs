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
for (const capability of ["tables.read","tables.write","tables.schema","tables.update","tables.delete","tables.admin.schema"]) await action("plugin.grant", {capability});
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
const createdTable = await tableAdmin("tables.admin.create",{table:"schema-fixture"});
assert.equal(createdTable.exists,true);
await tableAdmin("tables.write",{table:"schema-fixture",row:{name:"one",flag:true}});
const profile = await tableAdmin("tables.schema.profile",{table:"schema-fixture"},true);
assert.deepEqual(profile.columns,[{name:"flag",types:["boolean"],present:1,missing:0},{name:"name",types:["string"],present:1,missing:0}]);
const renamed = await tableAdmin("tables.admin.rename-column",{table:"schema-fixture",revision:profile.revision,column:"name",new_column:"label"});
const added = await tableAdmin("tables.admin.add-column",{table:"schema-fixture",revision:renamed.revision,column:"count",default:0});
await tableAdmin("tables.admin.drop-column",{table:"schema-fixture",revision:added.revision,column:"flag"});
assert.deepEqual((await tableAdmin("tables.read",{table:"schema-fixture",limit:10},true)).rows,[{row_index:0,row:{label:"one",count:0}}]);
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
// Exercise the authenticated streaming endpoint in the instrumented host too.
const blobPath = base + "/ci-project/admin/files/blob?";
const blobURL = (key, more = {}) => blobPath + new URLSearchParams({version:"0.1.0",key,...more});
const blobHeaders = {...headers,"content-type":"application/octet-stream"};
const binaryKey = "ci-binary";
assert.equal((await fetch(blobURL(binaryKey))).status,401);
assert.equal((await fetch(blobPath+"key=ci-binary",{headers})).status,400);
assert.equal((await fetch(blobURL(binaryKey)+"&key=duplicate",{headers})).status,400);
assert.equal((await fetch(blobURL(binaryKey),{headers})).status,404);
assert.equal((await fetch(blobURL(binaryKey),{method:"POST",headers:blobHeaders,body:"abc"})).status,400);
assert.equal((await fetch(blobURL(binaryKey,{confirmed:"true"}),{method:"POST",headers,body:"abc"})).status,415);
await action("plugin.configure",{plugin_id:"files",values:{maxUploadBytes:16}});
assert.equal((await fetch(blobURL(binaryKey,{confirmed:"true"}),{method:"POST",headers:blobHeaders,body:Buffer.alloc(17)})).status,413);
await action("plugin.configure",{plugin_id:"files",values:{maxUploadBytes:2*1024*1024}});
const binaryData = Buffer.alloc(1024*1024+5,0xa5);
binaryData[0]=0;
binaryData[binaryData.length-1]=42;
const binaryUpload = await fetch(blobURL(binaryKey,{confirmed:"true",filename:"fixture.bin",contentType:"application/octet-stream"}),{method:"POST",headers:blobHeaders,body:binaryData});
assert.equal(binaryUpload.status,200,await binaryUpload.clone().text());
const binaryMetadata = (await binaryUpload.json()).metadata;
assert.equal(binaryMetadata.size_bytes,binaryData.length);
assert.equal(binaryMetadata.encoding,"binary");
assert.equal(binaryMetadata.filename,"fixture.bin");
assert.match(binaryMetadata.sha256,/^[a-f0-9]{64}$/);
const binaryDownload = await fetch(blobURL(binaryKey),{headers});
assert.equal(binaryDownload.status,200);
assert.equal(binaryDownload.headers.get("content-length"),String(binaryData.length));
assert.equal(binaryDownload.headers.get("cache-control"),"no-store");
assert.deepEqual(Buffer.from(await binaryDownload.arrayBuffer()),binaryData);
assert.deepEqual(await invoke("games","games.modes",{}),{projectId:"ci-project",modes:["tic-tac-toe"]});
assert.deepEqual(await invoke("pvp","pvp.describe",{}),{projectId:"ci-project",turnBased:true,revisionedCommands:true});
const rules = await invoke("games","games.mode.describe",{mode:"tic-tac-toe"});
assert.equal(rules.players,2);
assert.deepEqual(rules.board,{rows:3,columns:3});
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
assert.deepEqual(matches.rows,[{key:arenaRoom,status:"finished",member_count:2}]);
assert.deepEqual((await action("plugin.admin.inventory",{plugin_id:"arena",page_id:"matches",limit:20,search:"CI-ARENA"})).rows,[{key:arenaRoom,status:"finished",member_count:2}]);
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
const accountSearch = await invoke("auth","auth.admin.accounts",{search:"CI-USER",kind:"registered",limit:1});
assert.equal(accountSearch.rows[0].username,"ci-user");
const renewed = await invoke("auth","auth.signin",{username:"ci-user",password:"ci-fixture-password"});
const sessionSummary = await invoke("auth","auth.admin.sessions.inspect",{username:"ci-user"});
assert.equal(sessionSummary.activeSessions,1);
assert.ok(sessionSummary.latestExpiry>Date.now());
assert.ok(!JSON.stringify(sessionSummary).includes(renewed.token));
async function confirmedAdmin(plugin_id,page_id,target_operation,payload) {
  const response = await action("plugin.admin.invoke",{plugin_id,page_id,target_operation,payload,confirmed:true});
  assert.equal(response.Success,true,target_operation + ": " + response.ErrorCode);
  return JSON.parse(Buffer.from(response.Payload,"base64").toString("utf8"));
}
assert.deepEqual(await confirmedAdmin("auth","accounts","auth.admin.accounts.suspend",{username:"ci-user"}),{username:"ci-user",suspended:true,revoked:1});
const suspendedSignin = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"auth",version:"0.1.0",target_operation:"auth.signin",payload:{username:"ci-user",password:"ci-fixture-password"}})});
assert.ok(!suspendedSignin.ok || !(await suspendedSignin.json()).Success,"suspended account signed in");
assert.equal((await invoke("auth","auth.admin.accounts",{kind:"suspended"})).rows[0].suspended,true);
assert.deepEqual(await confirmedAdmin("auth","accounts","auth.admin.accounts.reactivate",{username:"ci-user"}),{username:"ci-user",suspended:false,revoked:0});
assert.ok((await invoke("auth","auth.signin",{username:"ci-user",password:"ci-fixture-password"})).token);
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
const interruptedRoom = "ci-interrupted";
await invoke("arena","arena.start",{roomId:interruptedRoom,players:["one","two"]});
const interrupted = await confirmedAdmin("arena","matches","arena.admin.interrupt",{roomId:interruptedRoom,reason:"operator test"});
assert.equal(interrupted.state,"interrupted");
assert.equal(interrupted.interruption.reason,"operator test");
await action("plugin.stop",{plugin_id:"arena",confirmed:true});
await action("plugin.recover",{plugin_id:"arena"});
await action("plugin.worker.start",{plugin_id:"arena"});
assert.deepEqual(await invoke("arena","arena.snapshot",{roomId:interruptedRoom}),interrupted);
const afterInterrupt = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"arena",version:"0.1.0",target_operation:"arena.input",payload:{roomId:interruptedRoom,playerId:"one",sequence:1}})});
assert.ok(!afterInterrupt.ok || !(await afterInterrupt.json()).Success,"interrupted match accepted input");
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
const timeline = await invoke("analytics","analytics.timeline",{name:"rooms.connected",from:new Date(Date.now()-3600000).toISOString(),until:new Date(Date.now()+3600000).toISOString(),bucket:"hour"});
assert.ok(timeline.total_events>=1);
assert.equal(timeline.rows.reduce((total,row)=>total+row.count,0),timeline.total_events);
const managedRoom = "admin-room";
await invoke("rooms","rooms.create",{id:managedRoom});
await invoke("rooms","rooms.join",{id:managedRoom,member:"member-one"});
await invoke("rooms","rooms.join",{id:managedRoom,member:"member-two"});
assert.deepEqual((await confirmedAdmin("rooms","manage","rooms.admin.remove-member",{id:managedRoom,member:"member-one"})).members,["member-two"]);
assert.equal((await confirmedAdmin("rooms","manage","rooms.admin.close",{id:managedRoom})).closed,true);
const closedJoin = await fetch(base+"/ci-project/admin/actions",{method:"POST",headers,body:JSON.stringify({operation:"plugin.invoke",project_id:"ci-project",plugin_id:"rooms",version:"0.1.0",target_operation:"rooms.join",payload:{id:managedRoom,member:"member-three"}})});
assert.ok(!closedJoin.ok || !(await closedJoin.json()).Success,"closed room accepted join");
const roomRows = await action("plugin.admin.inventory",{plugin_id:"rooms",page_id:"operations",search:managedRoom,limit:10});
assert.deepEqual(roomRows.rows,[{key:managedRoom,status:"closed",member_count:1}]);
await invoke("matchmaking","matchmaking.queue",{id:"ci-ticket"});
const tickets = await action("plugin.admin.inventory",{plugin_id:"matchmaking",page_id:"operations",search:"ci-ticket",limit:10});
assert.deepEqual(tickets.rows,[{key:"ci-ticket",status:"queued"}]);
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
