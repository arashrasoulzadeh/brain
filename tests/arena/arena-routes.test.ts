import { deepEqual, equal, throws } from "node:assert/strict";
import { test } from "node:test";
import { input, result, snapshot, start } from "../../plugins/arena/index.js";
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
