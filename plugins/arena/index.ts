import { request } from "../sdk-typescript/src/index.js";

export type DurableMatchResult = {
  resultId: string;
  projectId: string;
  roomId: string;
  winnerIds: readonly string[];
  completedAt: string;
};

export type ArenaState = "waiting" | "active" | "finished";

export class ArenaStateMachine {
  private currentState: ArenaState = "waiting";
  getState(): ArenaState { return this.currentState; }
  players: readonly string[] = [];
  constructor(readonly projectId: string, readonly roomId: string) {}
  start(players: readonly string[]): void {
    if (this.getState() !== "waiting") throw new Error("arena already started");
    if (players.length < 2 || new Set(players).size !== players.length || players.some((player) => !player.trim())) throw new Error("invalid arena players");
    this.players = [...players];
    this.currentState = "active";
  }
}

export function arenaStartCommand(arena: ArenaStateMachine, projectId: string, roomId: string, players: readonly string[]): ArenaStateMachine {
  if (arena.projectId !== projectId || arena.roomId !== roomId) throw new Error("arena project mismatch");
  arena.start(players);
  return arena;
}

export function validateMatchResult(result: DurableMatchResult): void {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*:[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(result.resultId)) throw new Error("invalid result id");
  if (!result.projectId || !result.roomId || !result.completedAt || result.winnerIds.length === 0) throw new Error("invalid match result");
  if (!result.resultId.startsWith(`${result.projectId}:`)) throw new Error("result project mismatch");
}

export function arenaRequest(projectId: string, pluginId: string, version: string, operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result", roomId: string, payload: unknown = {}) {
  return request(`arena-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

