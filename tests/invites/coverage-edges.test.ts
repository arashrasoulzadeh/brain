import assert from "node:assert/strict";
import { test } from "node:test";
import { InviteCreator, InviteAccepter, InviteCodeAccepter, create, accept, cancel } from "../../plugins/invites/index.js";
import type { Invite } from "../../plugins/invites/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

const context: PluginContext = {
  project: { id: "coverage-invites", active: true },
  plugin: { id: "invites", version: "1" },
  execution: { operation: "create" },
  plugins: { observe: { fire() {} } },
};

test("invite core enforces target, expiry, project, time and input boundaries", () => {
  let seq = 0;
  const store = new Map<string, Invite>();
  const creator = new InviteCreator(store, () => ++seq);
  const accepter = new InviteAccepter(store);
  const codes = new InviteCodeAccepter(store);
  for (const [room, player, ttl, now] of [["", "p", 1, 0], ["r", "", 1, 0], ["r", "p", 0, 0], ["r", "p", 1, NaN]] as const) {
    assert.throws(() => creator.create("alpha", room, player, ttl, now), /invalid invite/);
  }
  const invite = creator.create("alpha", "r", "p", 60);
  for (const [id, player, now] of [["", "p", 1], [invite.id, "", 1], [invite.id, "p", NaN], [invite.id, "wrong", 1], ["missing", "p", 1]] as const) {
    assert.throws(() => accepter.accept(id, "alpha", player, now), /invite unavailable/);
  }
  assert.throws(() => codes.acceptCode("missing", "alpha", "p"), /invite unavailable/);
  assert.equal(codes.acceptCode(invite.code, "alpha", "p").used, true);
});

test("invite routes enforce capacity, defaults, cancellation and credential shape", () => {
  const payload = { roomId: "room", targetPlayerId: "player" };
  for (const ttlSeconds of ["60", 0, 1.5, 86401]) {
    assert.throws(() => create(context, { ...payload, ttlSeconds }), /invalid invite ttl/);
  }
  for (const roomId of ["", "a".repeat(129), "bad/id"]) {
    assert.throws(() => create(context, { ...payload, roomId }), /room id/);
  }
  const first = create(context, payload);
  for (const credentials of [{}, { inviteId: first.id, code: first.code }, { inviteId: "a".repeat(257) }, { code: "a".repeat(257) }]) {
    assert.throws(() => accept(context, { playerId: "player", ...credentials }), /invite unavailable/);
  }
  assert.equal(accept(context, { playerId: "player", inviteId: first.id }).code, "");
  assert.throws(() => cancel(context, { inviteId: first.id }), /invite unavailable/);
  assert.throws(() => cancel(context, { inviteId: "missing" }), /invite unavailable/);
  const realNow = Date.now;
  try {
    let now = realNow() + 100000;
    Date.now = () => now;
    const expired = create(context, { ...payload, ttlSeconds: 1 });
    now += 1000;
    assert.throws(() => cancel(context, { inviteId: expired.id }), /invite unavailable/);
    const live = [];
    for (let i = 0; i < 4096; i++) live.push(create(context, { ...payload, ttlSeconds: 60 }));
    assert.throws(() => create(context, payload), /invite capacity reached/);
    for (const invite of live) cancel(context, { inviteId: invite.id });
  } finally {
    Date.now = realNow;
  }
});
