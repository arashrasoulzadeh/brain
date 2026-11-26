import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { ArenaState, ArenaMeasurement, DurableMatchResult } from "./types.js";
import { validateMatchResult } from "./match-results.js";

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
    if (!Number.isInteger(sequence) || sequence <= (this.runtime.sequences.get(playerId) ?? 0))
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
      winnerIds.some((winner) => !this.runtime.players.includes(winner)) ||
      !completedAt
    )
      throw new Error("invalid arena result");
    this.runtime.state = "finished";
    const result = {
      resultId: `${this.runtime.projectId}:${this.runtime.roomId}-${Date.parse(completedAt)}`,
      projectId: this.runtime.projectId,
      roomId: this.runtime.roomId,
      winnerIds: [...winnerIds],
      completedAt,
    };
    validateMatchResult(result);
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
