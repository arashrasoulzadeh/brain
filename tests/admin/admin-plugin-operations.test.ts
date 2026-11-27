import assert from "node:assert/strict";
import { test } from "node:test";
import { create as createInvite, inspect as inspectInvites } from "../../plugins/invites/index.js";
import { open, inspect as inspectSubscriptions, revokeChannel, receive } from "../../plugins/subscriptions/index.js";
import { dispatchRequest, signUpRequest, signInRequest, guestRequest } from "../../plugins/auth/index.js";
import { pluginRequest, requireOperator, type InvocationActor, type PluginContext } from "../../plugins/sdk-typescript/src/index.js";

const operator: InvocationActor = { kind: "operator", id: "root", authenticated: true, level: "authenticated" };
const user: InvocationActor = { kind: "user", id: "user", authenticated: true, level: "authenticated" };
const delegated: InvocationActor = { kind: "plugin", id: "dashboard@1", authenticated: false, level: "system" };
function context(project: string, plugin: string, operation: string, capability = operation, resource = "", caller = operator): PluginContext {
  return {
    project: { id: project, active: true },
    plugin: { id: plugin, version: "0.1.0" },
    execution: { operation },
    auth: { actor: operator, caller, permissions: [{ operation, capability, resource }],
      can: (requested, key = "") => requested === capability && key === resource },
    plugins: { observe: { fire() {} } },
  };
}

test("administrative worker operations require effective operator and exact permission", () => {
  const scope = context("one", "auth", "auth.admin.accounts", "auth.admin.read");
  assert.equal(requireOperator(scope, "auth.admin.read"), scope);
  assert.throws(() => requireOperator(scope, "auth.admin.manage"), /operator authorization/);
  assert.throws(() => requireOperator(scope, "auth.admin.read", "different"), /operator authorization/);
  assert.throws(() => requireOperator(undefined, "auth.admin.read"), /operator authorization/);
  for (const caller of [user, delegated, { ...operator, authenticated: false }]) {
    assert.throws(() => requireOperator(context("one", "auth", "auth.admin.accounts", "auth.admin.read", "", caller), "auth.admin.read"), /operator authorization/);
  }
});

test("invite inspection filters project and room and excludes invitation codes", () => {
  const project = "admin-invites";
  const created = createInvite(context(project, "invites", "invites.create"), { roomId: "room-a", targetPlayerId: "player", ttlSeconds: 60 });
  createInvite(context(project, "invites", "invites.create"), { roomId: "room-b", targetPlayerId: "player" });
  createInvite(context("other-invites", "invites", "invites.create"), { roomId: "room-a", targetPlayerId: "player" });
  const reader = context(project, "invites", "invites.inspect", "invites.inspect", "room-a");
  const rows = inspectInvites(reader, { roomId: "room-a" });
  assert.equal(rows.length, 1);
  assert.equal(rows[0].id, created.id);
  assert.equal(rows[0].status, "pending");
  assert.ok(!JSON.stringify(rows).includes(created.code));
  assert.throws(() => inspectInvites(reader, { roomId: "room-b" }), /operator authorization/);
  assert.throws(() => inspectInvites(context(project, "invites", "invites.inspect", "invites.inspect", "room-a", user), { roomId: "room-a" }), /operator authorization/);
});

test("subscription inspection never returns tokens and revocation is project scoped", () => {
  const project = "admin-subscriptions";
  const channel = "rooms.connected";
  const opened = open(context(project, "subscriptions", "subscriptions.open"), { channel });
  const other = open(context("other-subscriptions", "subscriptions", "subscriptions.open"), { channel });
  const reader = context(project, "subscriptions", "subscriptions.inspect", "subscriptions.inspect", channel);
  const summary = inspectSubscriptions(reader, { channel });
  assert.deepEqual(summary, { channel, activeTickets: 1, cursor: 0, retainedEvents: 0 });
  assert.ok(!JSON.stringify(summary).includes(opened.ticket));
  assert.throws(() => revokeChannel(reader, { channel }), /operator authorization/);
  const revoker = context(project, "subscriptions", "subscriptions.revoke", "subscriptions.revoke", channel);
  assert.deepEqual(revokeChannel(revoker, { channel }), { channel, revoked: 1 });
  assert.equal(revokeChannel(revoker, { channel }).revoked, 0);
  assert.throws(() => receive(context(project, "subscriptions", "subscriptions.receive"), { channel, ticket: opened.ticket }), /invalid subscription ticket/);
  assert.equal(receive(context("other-subscriptions", "subscriptions", "subscriptions.receive"), { channel, ticket: other.ticket }).events.length, 0);
});

test("auth account metadata and session revocation never expose credential material", async () => {
  const project = "admin-auth";
  const foreign = "other-admin-auth";
  const username = "operator-test-user";
  for (const id of [project, foreign]) await dispatchRequest("signup", signUpRequest(id, "auth", "0.1.0", username, "test-password"));
  const signedIn = await dispatchRequest("signin", signInRequest(project, "auth", "0.1.0", username, "test-password")) as { token: string };
  const foreignSession = await dispatchRequest("signin", signInRequest(foreign, "auth", "0.1.0", username, "test-password")) as { token: string };
  const guest = dispatchRequest("guest", guestRequest(project, "auth", "0.1.0")) as { token: string; username: string };
  const list = pluginRequest("auth", project, "auth", "0.1.0", "auth.admin.accounts", {});
  const reader = context(project, "auth", "auth.admin.accounts", "auth.admin.read");
  const data = dispatchRequest("listAccounts", list, reader) as { rows: { username: string; activeSessions: number; guest: boolean }[]; next: string };
  assert.equal(data.rows.length, 2);
  assert.equal(data.rows.find(row => row.username === username)?.activeSessions, 1);
  assert.equal(data.rows.find(row => row.username === guest.username)?.guest, true);
  const encoded = JSON.stringify(data);
  for (const secret of [signedIn.token, guest.token, foreignSession.token, "test-password", "passwordHash", "digest"]) assert.ok(!encoded.includes(secret));
  assert.throws(() => dispatchRequest("listAccounts", list), /operator authorization/);
  assert.throws(() => dispatchRequest("listAccounts", list, context(project, "auth", "auth.admin.accounts", "auth.admin.read", "", delegated)), /operator authorization/);
  assert.throws(() => dispatchRequest("listAccounts", { ...list, params: { ...list.params, payload: { after: 12 } } }, reader), /invalid account cursor/);
  const revoke = pluginRequest("auth", project, "auth", "0.1.0", "auth.admin.sessions.revoke", { username });
  assert.throws(() => dispatchRequest("revokeSessions", revoke, reader), /operator authorization/);
  const revoker = context(project, "auth", "auth.admin.sessions.revoke", "auth.admin.manage", username);
  assert.deepEqual(dispatchRequest("revokeSessions", revoke, revoker), { username, revoked: 1 });
  assert.throws(() => dispatchRequest("session", pluginRequest("auth", project, "auth", "0.1.0", "auth.session", { token: signedIn.token })), /invalid session/);
  assert.deepEqual(dispatchRequest("session", pluginRequest("auth", foreign, "auth", "0.1.0", "auth.session", { token: foreignSession.token })), { username });
});
