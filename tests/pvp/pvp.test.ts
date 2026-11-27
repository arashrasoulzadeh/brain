import assert from "node:assert/strict";
import { test } from "node:test";
import { CommandDeduplicator, assertMatchMember, nextTurn, createRevisionRuntime, commitRevision, replayHistory, expireTurn, pauseTurn, resumeTurn, privatePlayerView, parseMatchState, describe } from "../../plugins/pvp/index.js";
import { modes, initializeTicTacToe, reconnectTicTacToe, serializeTicTacToeState } from "../../plugins/games/index.js";
import * as rooms from "../../plugins/rooms/index.js";
import type { PluginContext } from "../../plugins/sdk-typescript/src/index.js";

const state = { projectId: "dev", roomId: "r1", players: ["a", "b", "c"], turn: "a", status: "active" as "active" | "paused", turnDeadlineAt: 100, score: 0 };
const copy = (value: typeof state) => ({ ...value, players: [...value.players] });
const errors = { invalid: "invalid", unexpired: "unexpired" };

test("room exports are game agnostic and PvP utilities support a three-player mode", () => {
  assert.equal("initializeTicTacToe" in rooms, false);
  assert.equal("CommandDeduplicator" in rooms, false);
  assert.equal(nextTurn(state.players, "a"), "b");
  assert.equal(nextTurn(state.players, "c"), "a");
  assert.throws(() => nextTurn(state.players, "x"), /invalid game turn/);
  assert.throws(() => nextTurn(["a"], "a"), /invalid game turn/);
  assert.doesNotThrow(() => assertMatchMember(state, "dev", "a"));
  assert.throws(() => assertMatchMember(state, "prod", "a"), /game authorization required/);
  assert.throws(() => assertMatchMember(state, "dev", "x"), /game authorization required/);
  assert.throws(() => assertMatchMember({ ...state, players: null as unknown as string[] }, "dev", "a"), /game authorization required/);
  assert.throws(() => assertMatchMember(state, "bad/project", "a"), /game authorization required/);
  const commands = new CommandDeduplicator();
  assert.throws(() => commands.accept(" "), /command id required/);
  assert.equal(commands.accept("cmd-1"), true);
  assert.equal(commands.accept("cmd-1"), false);
});

test("revision and replay reject invalid commands without mutating committed state", () => {
  const runtime = createRevisionRuntime(state, copy);
  assert.throws(() => commitRevision(runtime, 1, copy, copy), /stale game revision/);
  assert.throws(() => commitRevision(runtime, 0, (value) => { value.score = 9; throw new Error("denied"); }, copy), /denied/);
  assert.equal(runtime.current.score, 0);
  const result = commitRevision(runtime, 0, (value) => ({ ...value, score: 1 }), copy);
  result.players.push("outsider");
  assert.equal(runtime.current.players.length, 3);
  assert.equal(runtime.revision, 1);
  const apply = (value: typeof state, command: { revision: number; playerId: string }) => {
    assertMatchMember(value, value.projectId, command.playerId);
    return { ...value, score: value.score + 1 };
  };
  assert.equal(replayHistory(state, [{ revision: 0, playerId: "a" }], apply).score, 1);
  for (const command of [{ revision: 1, playerId: "a" }, { revision: 0, playerId: " " }, { revision: 0, playerId: "x" }])
    assert.throws(() => replayHistory(state, [command], apply), /invalid committed game history/);
});

test("turn recovery, private state and persisted input remain mode-independent", () => {
  assert.equal(expireTurn(state, 100, 200, errors).turn, "b");
  for (const invalid of [{ ...state, status: "paused" as const }, state])
    assert.throws(() => expireTurn(invalid, 100, 100, errors), /invalid/);
  assert.throws(() => expireTurn(state, NaN, 200, errors), /invalid/);
  assert.throws(() => expireTurn(state, 100, NaN, errors), /invalid/);
  assert.throws(() => expireTurn(state, 99, 200, errors), /unexpired/);
  assert.throws(() => expireTurn({ ...state, turnDeadlineAt: undefined }, 100, 200, errors), /unexpired/);
  assert.throws(() => expireTurn({ ...state, turnDeadlineAt: NaN }, 100, 200, errors), /unexpired/);
  const paused = pauseTurn(state, copy, "not active");
  assert.equal(paused.turnDeadlineAt, undefined);
  assert.throws(() => pauseTurn(paused, copy, "not active"), /not active/);
  assert.throws(() => resumeTurn(state, "v1", "v1", copy, "not paused"), /not paused/);
  for (const version of ["", "v2"]) assert.throws(() => resumeTurn(paused, version, "v1", copy, "not paused"), /pinned/);
  assert.equal(resumeTurn(paused, "v1", "v1", copy, "not paused").status, "active");
  assert.equal(privatePlayerView(state, "a", { a: 42, b: 99 }, copy).privateState, 42);
  assert.equal(privatePlayerView(state, "x", { x: 42 }, copy).privateState, undefined);
  assert.equal(privatePlayerView(state, "a", Object.create({ a: 42 }) as Record<string, number>, copy).privateState, undefined);
  assert.deepEqual(parseMatchState('{"score":1}'), { score: 1 });
  for (const invalid of ["{", "null", "[]", "3"])
    assert.throws(() => parseMatchState(invalid), /invalid persisted game/);
  const game = { ...initializeTicTacToe("dev", "r1", "a", "b"), turnDeadlineAt: 100 };
  assert.equal(reconnectTicTacToe(serializeTicTacToeState(game), "dev", "a").turnDeadlineAt, 100);
});

test("plugin discovery routes use trusted active project context", () => {
  const context: PluginContext = {
    project: { id: "dev", active: true },
    plugin: { id: "games", version: "0.1.0" },
    execution: { operation: "games.modes" },
    plugins: { observe: { fire() {} } },
  };
  assert.deepEqual(modes(context, {}), { projectId: "dev", modes: ["tic-tac-toe"] });
  assert.deepEqual(describe({ ...context, plugin: { id: "pvp", version: "0.1.0" }, execution: { operation: "pvp.describe" } }, {}), { projectId: "dev", turnBased: true, revisionedCommands: true });
  for (const handler of [modes, describe]) {
    assert.throws(() => handler(context, { projectId: "prod" }), /host project context mismatch/);
    assert.throws(() => handler({ ...context, project: { id: "dev", active: false } }, {}), /host project context required/);
  }
});
