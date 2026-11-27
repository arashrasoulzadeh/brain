import { deepEqual, equal, throws } from "node:assert/strict";
import { test } from "node:test";
import { open, receive, resync, close, observeRoomConnected } from "../../plugins/subscriptions/index.js";
import type { ObservationIntent, PluginContext } from "../../plugins/sdk-typescript/src/index.js";

function context(id: string, events: ObservationIntent[]): PluginContext {
  return {
    project: { id, active: true },
    plugin: { id: "subscriptions", version: "0.1.0" },
    execution: { operation: "subscriptions.open" },
    plugins: { observe: { fire: (event, payload) => events.push({ event, payload }) } },
  };
}

test("subscription routes read actual project-scoped observations and revoke closed tickets", () => {
  const events: ObservationIntent[] = [];
  const first = context("sub-one", events);
  const second = context("sub-two", []);
  const channel = "rooms.connected";
  throws(() => open(first, { channel, projectId: "sub-two" }), /project context mismatch/);
  throws(() => open(first, { channel: "unwired" }), /channel unavailable/);
  const opened = open(first, { channel });
  const other = open(second, { channel });
  const request = { channel, ticket: opened.ticket, cursor: opened.cursor };
  const event = { projectId: "sub-one", pluginId: "rooms", version: "1", event: "connected", payload: { roomId: "room", reconnected: false } };
  throws(() => observeRoomConnected(second, event), /project context mismatch/);
  observeRoomConnected(first, event);
  deepEqual(receive(first, request), { cursor: 1, events: [{ sequence: 1, value: event.payload }] });
  equal(receive(second, { channel, ticket: other.ticket, cursor: other.cursor }).events.length, 0);
  throws(() => receive(second, request), /invalid subscription ticket/);
  deepEqual(resync(first, request).events, [{ sequence: 1, value: event.payload }]);
  close(first, request);
  throws(() => receive(first, request), /invalid subscription ticket/);
  deepEqual(events, ["opened", "closed"].map(event => ({ event, payload: {} })));
});
