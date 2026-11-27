import { createProjectMap, requireProjectInput, requiredString, validateProjectId, type PluginContext } from "../sdk-typescript/src/index.js";
import type { ArenaState, ArenaMeasurement, DurableMatchResult } from "./types.js";
import { validateMatchResult } from "./match-results.js";

const arenas = createProjectMap<ReturnType<typeof createArena>>();

export type ArenaRuntime = {
  projectId: string;
  roomId: string;
  state: ArenaState;
  players: string[];
  sequences: Map<string, number>;
};
export function createArenaRuntime(projectId: string, roomId: string): ArenaRuntime {
  validateProjectId(projectId, "invalid arena project");
  validateProjectId(roomId, "invalid arena room");
  return { projectId, roomId, state: "waiting", players: [], sequences: new Map() };
}

export class ArenaStateReader {
  constructor(private readonly runtime: ArenaRuntime) {}
  read(): ArenaState {
    return this.runtime.state;
  }
}

export class ArenaStarter {
  constructor(private readonly runtime: ArenaRuntime) {}
  start(players: readonly string[]): void {
    if (this.runtime.state !== "waiting") throw new Error("arena already started");
    if (
      players.length < 2 ||
      new Set(players).size !== players.length ||
      players.some((player) => !player.trim())
    )
      throw new Error("invalid arena players");
    this.runtime.players = [...players];
    this.runtime.state = "active";
  }
}

export class ArenaActionRecorder {
  constructor(private readonly runtime: ArenaRuntime) {}
  record(playerId: string, sequence: number, payload: unknown): void {
    if (this.runtime.state !== "active") throw new Error("arena is not active");
    if (!this.runtime.players.includes(playerId)) throw new Error("player not in arena");
    if (!Number.isSafeInteger(sequence) || sequence <= (this.runtime.sequences.get(playerId) ?? 0))
      throw new Error("action sequence conflict");
    void payload;
    this.runtime.sequences.set(playerId, sequence);
  }
}

export class ArenaMeasurer {
  constructor(private readonly runtime: ArenaRuntime) {}
  measure(iterations: number): ArenaMeasurement {
    if (!Number.isInteger(iterations) || iterations < 1)
      throw new Error("invalid measurement iterations");
    const start = globalThis.performance.now();
    let memoryBytes = 0;
    for (let index = 0; index < iterations; index += 1)
      memoryBytes += JSON.stringify({
        state: this.runtime.state,
        players: this.runtime.players,
        sequences: [...this.runtime.sequences],
      }).length;
    return { iterations, elapsedMs: globalThis.performance.now() - start, memoryBytes };
  }
}

export class ArenaFinisher {
  constructor(private readonly runtime: ArenaRuntime) {}
  finish(winnerIds: readonly string[], completedAt: string): DurableMatchResult {
    if (this.runtime.state !== "active") throw new Error("arena is not active");
    if (
      !winnerIds.length ||
      new Set(winnerIds).size !== winnerIds.length ||
      winnerIds.some((winner) => !this.runtime.players.includes(winner)) ||
      !Number.isFinite(Date.parse(completedAt))
    )
      throw new Error("invalid arena result");
    const result = {
      resultId: `${this.runtime.projectId}:${this.runtime.roomId}-${Date.parse(completedAt)}`,
      projectId: this.runtime.projectId,
      roomId: this.runtime.roomId,
      winnerIds: [...winnerIds],
      completedAt,
    };
    validateMatchResult(result);
    this.runtime.state = "finished";
    return result;
  }
}

export function createArena(projectId: string, roomId: string) {
  const runtime = createArenaRuntime(projectId, roomId);
  const reader = new ArenaStateReader(runtime);
  const starter = new ArenaStarter(runtime);
  const actions = new ArenaActionRecorder(runtime);
  const measurer = new ArenaMeasurer(runtime);
  const finisher = new ArenaFinisher(runtime);
  return {
    projectId,
    roomId,
    getState: reader.read.bind(reader),
    start: starter.start.bind(starter),
    action: actions.record.bind(actions),
    measure: measurer.measure.bind(measurer),
    result: finisher.finish.bind(finisher),
    get players() {
      return runtime.players;
    },
  };
}

export function arenaStartCommand(
  arena: ReturnType<typeof createArena>,
  projectId: string,
  roomId: string,
  players: readonly string[],
): ReturnType<typeof createArena> {
  validateProjectId(projectId, "arena project mismatch");
  if (arena.projectId !== projectId || arena.roomId !== roomId)
    throw new Error("arena project mismatch");
  arena.start(players);
  return arena;
}

export function interruptExpiredActionRoom(
  state: ArenaState,
  now: number,
  expiresAt: number,
): ArenaState {
  if (state !== "active" || !Number.isFinite(now) || !Number.isFinite(expiresAt) || now < expiresAt)
    throw new Error("action room has not expired");
  return "interrupted";
}

/** Operator-invoked worker routes; realtime actor/membership binding remains a host responsibility. */
export function start(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const roomId = routeRoomId(payload);
  const projectId = context.project.id;
  if (arenas.get(projectId, roomId)) throw new Error("arena already started");
  const players = playerList(payload.players);
  const arena = createArena(projectId, roomId);
  arena.start(players);
  context.plugins.observe.fire("started", { roomId, playerCount: players.length });
  arenas.set(projectId, roomId, arena);
  return arenaSnapshot(arena);
}

export function input(context: PluginContext, request: unknown) {
  const payload = requireProjectInput(context, request);
  const arena = routeArena(context, payload);
  const playerId = requiredString(payload.playerId, "player id is required");
  if (typeof payload.sequence !== "number") throw new Error("invalid action sequence");
  arena.action(playerId, payload.sequence, payload.payload);
  return { projectId: arena.projectId, roomId: arena.roomId, accepted: true };
}

export function snapshot(context: PluginContext, input: unknown) {
  return arenaSnapshot(routeArena(context, requireProjectInput(context, input)));
}

export function result(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const arena = routeArena(context, payload);
  const completedAt = requiredString(payload.completedAt, "completion time is required");
  const finished = arena.result(playerList(payload.winnerIds), completedAt);
  context.plugins.observe.fire("finished", { roomId: arena.roomId, winnerCount: finished.winnerIds.length });
  return finished;
}

function routeRoomId(payload: Record<string, unknown>): string {
  const roomId = requiredString(payload.roomId, "room id is required");
  if (roomId.length > 128) throw new Error("invalid arena room");
  return validateProjectId(roomId, "invalid arena room");
}

function routeArena(context: PluginContext, payload: Record<string, unknown>) {
  const arena = arenas.get(context.project.id, routeRoomId(payload));
  if (!arena) throw new Error("arena not found");
  return arena;
}

function playerList(value: unknown): string[] {
  if (!Array.isArray(value) || value.length > 64 ||
      value.some((player: unknown) => typeof player !== "string" || !player.trim() || player.length > 128))
    throw new Error("invalid arena players");
  return value as string[];
}

function arenaSnapshot(arena: ReturnType<typeof createArena>) {
  return { projectId: arena.projectId, roomId: arena.roomId, state: arena.getState(), players: [...arena.players] };
}
