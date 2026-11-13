import type { ArenaMeasurement, ArenaState, DurableMatchResult } from "./types.js";
import { validateMatchResult } from "./validate-result.js";

export class ArenaStateMachine {
  private currentState: ArenaState = "waiting";
  private readonly sequences = new Map<string, number>();
  players: readonly string[] = [];
  constructor(readonly projectId: string, readonly roomId: string) {}
  getState(): ArenaState { return this.currentState; }
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
  measure(iterations: number): ArenaMeasurement {
    if (!Number.isInteger(iterations) || iterations < 1) throw new Error("invalid measurement iterations");
    const start = globalThis.performance.now();
    let memoryBytes = 0;
    for (let index = 0; index < iterations; index += 1) memoryBytes += JSON.stringify({ state: this.currentState, players: this.players, sequences: [...this.sequences] }).length;
    return { iterations, elapsedMs: globalThis.performance.now() - start, memoryBytes };
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
