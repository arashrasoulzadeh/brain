import assert from "node:assert/strict";
import { test } from "node:test";
import { dispatchRequest } from "../../plugins/auth/service.js";
import { pluginRequest, type PluginContext, type InvocationActor } from "../../plugins/sdk-typescript/src/index.js";

const projectId = "durable-auth";
const operator: InvocationActor = { kind: "operator", id: "operator", authenticated: true, level: "authenticated" };
const anonymous: InvocationActor = { kind: "public", id: "", authenticated: false, level: "anonymous" };

function authContext(operation: string, capability = "", resource = "", actor = anonymous, events: string[] = []): PluginContext {
  return {
    project: { id: projectId, active: true },
    plugin: { id: "auth", version: "1" },
    execution: { operation },
    auth: { actor, caller: actor, permissions: capability ? [{ operation, capability, resource }] : [],
      can: (requested, key = "") => capability !== "" && requested === capability && key === resource },
    plugins: { observe: { fire(event) { events.push(event); } } },
  };
}

function durable(operation: string, payload: Record<string, unknown>, stored: unknown) {
  return pluginRequest("auth", projectId, "auth", "1", operation, { ...payload, _brainAuth: stored });
}

test("durable authentication restores only the host-provided account and session snapshot", async () => {
  const signupContext = authContext("auth.signup");
  const created = await dispatchRequest("signup", durable("auth.signup", { username: "alice", password: "fixture-password" }, {}), signupContext) as { result: { username: string; userId: string }; passwordHash: string };
  assert.deepEqual(created.result, { username: "alice", userId: projectId + ":alice" });
  assert.match(created.passwordHash, /^\$argon2/);
  assert.notEqual(created.passwordHash, "fixture-password");
  const account = { ...created.result, passwordHash: created.passwordHash };
  const signinContext = authContext("auth.signin");
  await assert.rejects(async () => dispatchRequest("signin", durable("auth.signin", { username: "alice", password: "wrong" }, { account }), signinContext), /invalid credentials/);
  await assert.rejects(async () => dispatchRequest("signin", durable("auth.signin", { username: "alice", password: "fixture-password" }, { account: { ...account, userId: "foreign:alice" } }), signinContext), /invalid account state/);
  const signed = await dispatchRequest("signin", durable("auth.signin", { username: "alice", password: "fixture-password" }, { account }), signinContext) as { result: { token: string }; expiresAt: number };
  assert.ok(signed.result.token);
  assert.ok(signed.expiresAt > Date.now());
  const session = { projectId, username: "alice", expiresAt: signed.expiresAt };
  const token = signed.result.token;
  assert.deepEqual(await dispatchRequest("identity", durable("auth.identity", { token }, { account, session }), authContext("auth.identity")), { userId: projectId + ":alice" });
  assert.deepEqual(await dispatchRequest("session", durable("auth.session", { token }, { session }), authContext("auth.session")), { username: "alice" });
  await assert.rejects(async () => dispatchRequest("session", durable("auth.session", { token }, { session: { ...session, projectId: "other" } }), authContext("auth.session")), /invalid session state/);
  await assert.rejects(async () => dispatchRequest("session", durable("auth.session", {}, { session }), authContext("auth.session")), /invalid session state/);
  await assert.rejects(async () => dispatchRequest("session", durable("auth.session", { token }, { session: { ...session, expiresAt: 0 } }), authContext("auth.session")), /invalid session/);
  const guest = await dispatchRequest("guest", durable("auth.guest", {}, {}), authContext("auth.guest")) as { result: { userId: string; guest: boolean }; expiresAt: number };
  assert.ok(guest.result.userId.startsWith(projectId + ":guest-"));
  assert.equal(guest.result.guest, true);
  assert.ok(guest.expiresAt > Date.now());
});

test("durable auth rejects missing context, corrupt snapshots and unadmitted admin operations", async () => {
  await assert.rejects(async () => dispatchRequest("guest", durable("auth.guest", {}, {})), /host project context required/);
  for (const snapshot of [null, 1, []]) {
    await assert.rejects(async () => dispatchRequest("guest", durable("auth.guest", {}, snapshot), authContext("auth.guest")), /invalid stored authentication state/);
  }
  await assert.rejects(async () => dispatchRequest("guest", durable("auth.guest", { projectId: "foreign" }, {}), authContext("auth.guest")), /project context mismatch/);
  const page = { rows: [{ username: "alice", userId: projectId + ":alice", guest: false, activeSessions: 1 }], next: "" };
  const readContext = authContext("auth.admin.accounts", "auth.admin.read", "", operator);
  const read = durable("auth.admin.accounts", {}, { page });
  await assert.rejects(async () => dispatchRequest("listAccounts", read, authContext("auth.admin.accounts")), /operator authorization/);
  assert.deepEqual(await dispatchRequest("listAccounts", read, readContext), page);
  for (const invalid of [undefined, { rows: {}, next: "" }, { rows: [], next: 1 }]) {
    await assert.rejects(async () => dispatchRequest("listAccounts", durable("auth.admin.accounts", {}, { page: invalid }), readContext), /invalid account page/);
  }
  const events: string[] = [];
  const manage = authContext("auth.admin.sessions.revoke", "auth.admin.manage", "alice", operator, events);
  await assert.rejects(async () => dispatchRequest("revokeSessions", durable("auth.admin.sessions.revoke", {}, {}), manage), /invalid username/);
  await assert.rejects(async () => dispatchRequest("revokeSessions", durable("auth.admin.sessions.revoke", { username: "bob" }, {}), manage), /operator authorization/);
  assert.deepEqual(await dispatchRequest("revokeSessions", durable("auth.admin.sessions.revoke", { username: "alice" }, {}), manage), { username: "alice", revoked: 0 });
  assert.deepEqual(events, ["sessions.revoked"]);
});

test("durable account management uses the exact operator scope and redacted session summary", async () => {
  const summary = { username: "alice", activeSessions: 2, earliestExpiry: 100, latestExpiry: 200 };
  const inspect = durable("auth.admin.sessions.inspect", { username: "alice" }, { sessions: summary });
  const read = authContext("auth.admin.sessions.inspect", "auth.admin.read", "alice", operator);
  assert.deepEqual(await dispatchRequest("inspectSessions", inspect, read), summary);
  await assert.rejects(async () => dispatchRequest("inspectSessions", inspect, authContext("auth.admin.sessions.inspect", "auth.admin.read", "bob", operator)), /operator authorization/);
  const events: string[] = [];
  const manage = authContext("auth.admin.accounts.suspend", "auth.admin.manage", "alice", operator, events);
  assert.deepEqual(await dispatchRequest("suspendAccount", durable("auth.admin.accounts.suspend", { username: "alice" }, {}), manage), { username: "alice", suspended: true });
  assert.deepEqual(events, ["account.status.changed"]);
  await assert.rejects(async () => dispatchRequest("suspendAccount", durable("auth.admin.accounts.suspend", { username: "bob" }, {}), manage), /operator authorization/);
});
