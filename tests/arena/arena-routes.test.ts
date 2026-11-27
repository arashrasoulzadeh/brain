import { deepEqual, equal, throws } from "node:assert/strict";
import { test } from "node:test";
import { input, result, snapshot, start, connect, playerInput } from "../../plugins/arena/index.js";
import type { ObservationIntent, PluginContext } from "../../plugins/sdk-typescript/src/index.js";

function context(id: string, observations: ObservationIntent[]): PluginContext {
  return {
    project: { id, active: true },
    plugin: { id: "arena", version: "0.1.0" },
    execution: { operation: "arena.start" },
    plugins: { observe: { fire: (event, payload) => observations.push({ event, payload }) } },
  };
}

test("arena routes execute real project-scoped state transitions and publish sanitized observations", () => {
  const events: ObservationIntent[] = [];
  const first = context("arena-one", events);
  const second = context("arena-two", []);
  const roomId = "shared-room";
  const players = ["player-one", "player-two"];

  throws(() => start(first, { projectId: "arena-two", roomId, players }), /project context mismatch/);
  deepEqual(start(first, { roomId, players }), { projectId: "arena-one", roomId, state: "active", players });
  throws(() => start(first, { roomId, players }), /already started/);
  throws(() => snapshot(second, { roomId }), /arena not found/);
  start(second, { roomId, players: ["other-one", "other-two"] });
  throws(() => input(first, { roomId, playerId: "other-one", sequence: 1 }), /player not in arena/);
  input(first, { roomId, playerId: players[0], sequence: 1, payload: { move: "left" } });
  throws(() => input(first, { roomId, playerId: players[0], sequence: 1 }), /sequence conflict/);
  throws(() => result(first, { roomId, winnerIds: [players[0]], completedAt: "invalid" }), /invalid arena result/);
  equal(snapshot(first, { roomId }).state, "active");
  const finished = result(first, { roomId, winnerIds: [players[0]], completedAt: "2026-11-13T00:00:00.000Z" });
  equal(finished.projectId, "arena-one");
  equal(snapshot(first, { roomId }).state, "finished");
  equal(snapshot(second, { roomId }).state, "active");
  throws(() => result(first, { roomId, winnerIds: [players[0]], completedAt: "2026-11-13T00:00:00.000Z" }), /not active/);
  deepEqual(events, [
    { event: "started", payload: { roomId, playerCount: 2 } },
    { event: "finished", payload: { roomId, winnerCount: 1 } },
  ]);
});

test("arena routes reject absent or inactive host project contexts", () => {
  const inactive = { ...context("inactive", []), project: { id: "inactive", active: false } };
  throws(() => snapshot(inactive, { roomId: "room" }), /host project context required/);
  throws(() => start(undefined as unknown as PluginContext, { roomId: "room", players: ["one", "two"] }), /host project context required/);
});

test("host-owned arena snapshots rehydrate a fresh runtime without replaying prior input", () => {
  const ctx = context("durable-arena", []);
  const roomId = "durable-room";
  const players = ["one", "two"];
  deepEqual(start(ctx, { roomId, players, _brainArena: null }), { projectId: "durable-arena", roomId, players, state: "active" });
  const stored = { state: "active", players, sequences: { one: 3 } };
  equal(snapshot(ctx, { roomId, _brainArena: stored }).state, "active");
  throws(() => input(ctx, { roomId, playerId: "one", sequence: 3, _brainArena: stored }), /sequence conflict/);
  deepEqual(input(ctx, { roomId, playerId: "one", sequence: 4, _brainArena: stored }), { projectId: "durable-arena", roomId, accepted: true });
  const completedAt = "2026-01-01T00:00:00Z";
  const match = result(ctx, { roomId, winnerIds: ["one"], completedAt, _brainArena: stored });
  deepEqual(snapshot(ctx, { roomId, _brainArena: { ...stored, state: "finished", result: { winnerIds: ["one"], completedAt } } }), { projectId: "durable-arena", roomId, players, state: "finished", result: match });
  throws(() => snapshot(ctx, { roomId }), /arena not found/);
  throws(() => snapshot(ctx, { roomId, _brainArena: { ...stored, sequences: { outsider: 3 } } }), /invalid stored sequence/);
});

test("realtime admission and player input derive identity from the trusted user session", () => {
  const ctx = context("socket-arena", []);
  const roomId = "room";
  const players = ["socket-arena:alice", "socket-arena:bob"];
  start(ctx, { roomId, players });
  function user(id: string): PluginContext {
    const actor = { kind: "user" as const, id, authenticated: true, level: "authenticated" as const };
    return { ...ctx, auth: { actor, caller: actor, permissions: [], can: () => true } };
  }
  const alice = user(players[0]);
  const bob = user(players[1]);
  throws(() => connect(ctx, { roomId }), /membership required/);
  throws(() => connect(user("foreign:alice"), { roomId }), /membership required/);
  throws(() => connect(user("socket-arena:outsider"), { roomId }), /membership required/);
  deepEqual(connect(alice, { roomId }), { admitted: true });
  deepEqual(connect(bob, { roomId }), { admitted: true });
  throws(() => playerInput(alice, { roomId, playerId: players[1], sequence: 1 }), /identity mismatch/);
  throws(() => playerInput(alice, { roomId, sequence: "1" }), /invalid action sequence/);
  deepEqual(playerInput(alice, { roomId, sequence: 1 }), { projectId: "socket-arena", roomId, accepted: true });
  throws(() => playerInput(alice, { roomId, sequence: 1 }), /sequence conflict/);
  playerInput(bob, { roomId, sequence: 1 });
  result(ctx, { roomId, winnerIds: [players[0]], completedAt: "2026-01-01T00:00:00Z" });
  throws(() => connect(alice, { roomId }), /membership required/);
  throws(() => playerInput(alice, { roomId, sequence: 2 }), /membership required/);
});
