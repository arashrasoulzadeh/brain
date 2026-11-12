import { request } from "../sdk-typescript/src/index.js";

export type DurableMatchResult = {
  resultId: string;
  projectId: string;
  roomId: string;
  winnerIds: readonly string[];
  completedAt: string;
};

export type EyeEvent = { projectId: string; name: "arena.result"; fields: { resultId: string; roomId: string; winnerCount: number } };

export function exportEyeEvent(result: DurableMatchResult): EyeEvent {
  validateMatchResult(result);
  return { projectId: result.projectId, name: "arena.result", fields: { resultId: result.resultId, roomId: result.roomId, winnerCount: result.winnerIds.length } };
}

export type ArenaState = "waiting" | "active" | "finished";

export class MatchResultStore {
  private readonly results = new Map<string, DurableMatchResult>();
  save(result: DurableMatchResult): void {
    validateMatchResult(result);
    if (this.results.has(result.resultId)) throw new Error("result already exists");
    this.results.set(result.resultId, { ...result, winnerIds: [...result.winnerIds] });
  }
  get(resultId: string): DurableMatchResult | undefined {
    const result = this.results.get(resultId);
    return result ? { ...result, winnerIds: [...result.winnerIds] } : undefined;
  }
}

export class ArenaStateMachine {
  private currentState: ArenaState = "waiting";
  getState(): ArenaState { return this.currentState; }
  players: readonly string[] = [];
  private readonly sequences = new Map<string, number>();
  constructor(readonly projectId: string, readonly roomId: string) {}
  start(players: readonly string[]): void {
    if (this.getState() !== "waiting") throw new Error("arena already started");
    if (players.length < 2 || new Set(players).size !== players.length || players.some((player) => !player.trim())) throw new Error("invalid arena players");
    this.players = [...players];
    this.currentState = "active";
  }
  action(playerId: string, sequence: number, payload: unknown): void {
    if (this.getState() !== "active") throw new Error("arena is not active");
    if (!this.players.includes(playerId)) throw new Error("player not in arena");
    if (!Number.isInteger(sequence) || sequence <= (this.sequences.get(playerId) ?? 0)) throw new Error("action sequence conflict");
    void payload;
    this.sequences.set(playerId, sequence);
  }
  result(winnerIds: readonly string[], completedAt: string): DurableMatchResult {
    if (this.getState() !== "active") throw new Error("arena is not active");
    if (!winnerIds.length || winnerIds.some((winner) => !this.players.includes(winner)) || !completedAt) throw new Error("invalid arena result");
    this.currentState = "finished";
    const result = { resultId: `${this.projectId}:${this.roomId}-${Date.parse(completedAt)}`, projectId: this.projectId, roomId: this.roomId, winnerIds: [...winnerIds], completedAt };
    validateMatchResult(result);
    return result;
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

