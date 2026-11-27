import assert from "node:assert/strict";
import { test } from "node:test";
import {
  authorizedRoomSnapshot, authorizeRoomCommand, ConnectionTicketConsumer,
  ConnectionTicketIssuer, createRoomDirectory, validateMembership,
  RoomRouteStore, dispatchRoute,
} from "../../plugins/rooms/index.js";
import {
  createRoomRuntime, isRoomDraining, RoomCreator, RoomJoiner,
  RoomLeaver, RoomEventReader, startRoomDrain,
} from "../../plugins/rooms/directory.js";
import type { ConnectionTicket } from "../../plugins/rooms/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

test("tickets reject malformed, expired or wrongly scoped claims without consuming a valid ticket", () => {
  const tickets = new Map<string, ConnectionTicket>();
  const issuer = new ConnectionTicketIssuer(tickets);
  const consumer = new ConnectionTicketConsumer(tickets);
  for (const [player, time, ttl] of [
    [" ", 0, 1], ["alice", NaN, 1], ["alice", 0, Infinity], ["alice", 0, 0],
  ] as const) {
    assert.throws(() => issuer.issue("alpha", "room", player, time, ttl), /invalid connection ticket/);
  }
  const ticket = issuer.issue("alpha", "room", "alice", 0, 1);
  assert.equal(consumer.consume("", "alpha", "room", "alice", 1), false);
  assert.equal(consumer.consume(ticket.token, "alpha", "room", "", 1), false);
  assert.equal(consumer.consume("missing", "alpha", "room", "alice", 1), false);
  assert.equal(consumer.consume(ticket.token, "alpha", "room", "alice", 1000), false);
  assert.equal(consumer.consume(ticket.token, "alpha", "other", "alice", 1), false);
  assert.equal(consumer.consume(ticket.token, "alpha", "room", "bob", 1), false);
  assert.equal(consumer.consume(ticket.token, "alpha", "room", "alice", 1), true);
  assert.equal(ticket.used, false);
});

test("room membership, event offsets and admission reject invalid boundaries", () => {
  const directory = createRoomDirectory();
  for (const size of [0, 65]) {
    assert.throws(() => directory.create("alpha", "room", size), /invalid max players/);
    assert.throws(() => validateMembership("alpha", "room", [], size), /invalid player slots/);
  }
  directory.create("alpha", "room", 2);
  directory.join("alpha", "room", "alice", 0);
  assert.equal(directory.isMember("alpha", "missing", "alice"), false);
  for (const [player, slot] of [["", 1], ["bob", -1], ["bob", 2], ["bob", 0]] as const) {
    assert.throws(() => directory.join("alpha", "room", player, slot), /invalid membership/);
  }
  for (const slots of [
    [{ playerId: "", slot: 0 }],
    [{ playerId: "bob", slot: -1 }],
    [{ playerId: "bob", slot: 2 }],
    [{ playerId: "a", slot: 0 }, { playerId: "b", slot: 1 }, { playerId: "c", slot: 2 }],
  ]) {
    assert.throws(() => validateMembership("alpha", "room", slots, 2), /invalid player slots/);
  }
  assert.throws(() => directory.leave("alpha", "missing", "alice"), /player not in room/);
  assert.throws(() => directory.leave("alpha", "room", "bob"), /player not in room/);
  assert.deepEqual(directory.events("alpha", "missing"), []);
  for (const offset of [-1, 0.5]) {
    assert.throws(() => directory.eventsSince("alpha", "room", offset), /invalid event sequence/);
  }
  assert.throws(() => directory.startDrain(NaN), /invalid drain deadline/);
  assert.throws(() => directory.startDrain(0), /invalid drain deadline/);
  const room = directory.get("alpha", "room");
  assert.throws(() => authorizedRoomSnapshot(room, "beta", "alice"), /room membership required/);
  assert.throws(() => authorizeRoomCommand(room, "alpha", "alice", " "), /operation required/);
});

test("room runtime supports matching admission and expired drain, rejects active admission", () => {
  const runtime = createRoomRuntime();
  const creator = new RoomCreator(runtime);
  const joiner = new RoomJoiner(runtime);
  const leaver = new RoomLeaver(runtime);
  const reader = new RoomEventReader(runtime);
  startRoomDrain(runtime, Date.now() + 10000);
  runtime.drainUntil = 0;
  assert.equal(isRoomDraining(runtime), false);
  creator.create("alpha", "room", 2);
  const room = runtime.rooms.get("alpha:room")!;
  room.state = "active";
  assert.throws(() => joiner.join("alpha", "room", "alice", 0), /room not joinable/);
  room.state = "matching";
  joiner.join("alpha", "room", "alice", 0);
  assert.equal(reader.events("alpha", "room").length, 1);
  runtime.events.delete("alpha:room");
  joiner.join("alpha", "room", "bob", 1);
  leaver.leave("alpha", "room", "alice");
  assert.deepEqual(reader.events("alpha", "room"), []);
});

test("room route store handles results, membership and missing rooms", () => {
  const store = new RoomRouteStore();
  assert.throws(() => store.execute("join", "missing", "alice"), /room not found/);
  assert.deepEqual(store.execute("create", "room"), { id: "room", members: [] });
  assert.deepEqual(store.execute("join", "room"), {
    id: "room", member: undefined, reconnected: false, members: [],
  });
  assert.deepEqual(store.execute("join", "room", "alice"), {
    id: "room", member: "alice", reconnected: false, members: ["alice"],
  });
  assert.deepEqual(store.execute("reconnect", "room", "alice"), {
    id: "room", member: "alice", reconnected: true, members: ["alice"],
  });
  assert.deepEqual(store.execute("result", "room", undefined, { winner: "alice" }), {
    id: "room", result: { winner: "alice" }, members: ["alice"],
  });
  assert.deepEqual(store.execute("result", "room"), {
    id: "room", result: { winner: "alice" }, members: ["alice"],
  });
});

test("room routes validate identifiers and emit lifecycle observation events", () => {
  const events: string[] = [];
  const context: PluginContext = {
    project: { id: "alpha", active: true },
    plugin: { id: "rooms", version: "1" },
    execution: { operation: "create" },
    plugins: { observe: { fire(event) { events.push(event); } } },
  };
  assert.throws(() => dispatchRoute("create", null), /room id is required/);
  assert.throws(() => dispatchRoute("join", { member: "alice" }), /room id and member are required/);
  assert.throws(() => dispatchRoute("join", { id: "route-room" }), /room id and member are required/);
  dispatchRoute("create", { id: "route-room" }, context);
  dispatchRoute("join", { id: "route-room", member: "alice" }, context);
  dispatchRoute("reconnect", { id: "route-room", member: "alice" }, context);
  dispatchRoute("result", { id: "route-room", result: "done" }, context);
  dispatchRoute("result", { id: "route-room" }, context);
  const persisted = { id: "persisted", members: ["alice"] };
  assert.deepEqual(dispatchRoute("create", { id: "persisted", room: persisted }, context), persisted);
  assert.deepEqual(dispatchRoute("join", { id: "persisted", member: "alice", room: persisted }, context), {
    ...persisted, member: "alice", reconnected: false,
  });
  assert.deepEqual(dispatchRoute("result", { id: "persisted", room: persisted }, context), persisted);
  assert.deepEqual(events, ["created", "connected", "connected", "reconnected", "result.committed", "created", "connected"]);
  dispatchRoute("create", { id: "no-context" });
  dispatchRoute("reconnect", { id: "no-context", member: "alice" });
  dispatchRoute("result", { id: "no-context", result: "done" });
});
