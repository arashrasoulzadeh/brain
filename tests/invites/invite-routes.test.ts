import { deepEqual, equal, notEqual, throws } from "node:assert/strict";
import { test } from "node:test";
import { create, accept, cancel } from "../../plugins/invites/index.js";
import type { ObservationIntent, PluginContext } from "../../plugins/sdk-typescript/src/index.js";

function context(id: string, events: ObservationIntent[]): PluginContext {
  return {
    project: { id, active: true },
    plugin: { id: "invites", version: "0.1.0" },
    execution: { operation: "invites.create" },
    plugins: { observe: { fire: (event, payload) => events.push({ event, payload }) } },
  };
}

test("invite routes isolate projects and reject replay, wrong player and cancelled codes", () => {
  const events: ObservationIntent[] = [];
  const first = context("invite-one", events);
  const second = context("invite-two", []);
  const payload = { roomId: "room", target: "player", ttlSeconds: 60 };
  throws(() => create(first, { ...payload, projectId: "invite-two" }), /project context mismatch/);
  const invite = create(first, payload);
  const another = create(first, payload);
  notEqual(invite.id, another.id);
  notEqual(invite.code, another.code);
  throws(() => accept(second, { code: invite.code, playerId: "player" }), /invite unavailable/);
  throws(() => accept(first, { inviteId: invite.id, playerId: "wrong" }), /invite unavailable/);
  throws(() => cancel(second, { inviteId: another.id }), /invite unavailable/);
  equal(accept(first, { code: invite.code, playerId: "player" }).used, true);
  throws(() => accept(first, { inviteId: invite.id, playerId: "player" }), /invite unavailable/);
  cancel(first, { inviteId: another.id });
  throws(() => accept(first, { code: another.code, playerId: "player" }), /invite unavailable/);
  deepEqual(events, ["created", "created", "accepted", "cancelled"].map(event => ({ event, payload: {} })));
});
