import { randomUUID } from "node:crypto";
import { projectRoomID } from "./project-room-id.js";
import { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
import { validateTicTacToeRuleResult } from "./validate-tic-tac-toe-rule-result.js";
import { createTicTacToeRuntime } from "./tic-tac-toe-runtime.js";
import { TicTacToeRevisionReader } from "./tic-tac-toe-revision-reader.js";
import { TicTacToeMover } from "./tic-tac-toe-mover.js";

export { roomRequest } from "./room-request.js";

export { projectRoomID } from "./project-room-id.js";

export type PlayerSlot = { playerId: string; slot: number };

export type RoomState = "waiting" | "matching" | "active" | "finished" | "closed";

export type Room = { id: string; projectId: string; state: RoomState; maxPlayers: number; slots: PlayerSlot[] };
export type ConnectionTicket = { token: string; projectId: string; roomId: string; playerId: string; expiresAt: number; used: boolean };

export { authorizedRoomSnapshot } from "./authorized-room-snapshot.js";
export { authorizeRoomCommand } from "./authorize-room-command.js";

export { ConnectionTicketIssuer } from "./connection-ticket-issuer.js";
export { ConnectionTicketConsumer } from "./connection-ticket-consumer.js";
export type RoomEvent = { type: "player.joined" | "player.left"; projectId: string; roomId: string; playerId: string; sequence: number };
export type MembershipStore = { save(roomId: string, slots: readonly PlayerSlot[]): void };
export { createRoomDirectory } from "./create-room-directory.js";

export { canTransitionRoom } from "./room-transition.js";
export { validateMembership } from "./validate-membership.js";
export { initializeTicTacToe } from "./initialize-tic-tac-toe.js";
export { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
export { expireTicTacToeTurn } from "./expire-tic-tac-toe-turn.js";
export { ticTacToeWinner } from "./tic-tac-toe-winner.js";
export { ticTacToeDraw } from "./tic-tac-toe-draw.js";
export { validateTicTacToeRuleResult } from "./validate-tic-tac-toe-rule-result.js";
export type TicTacToeState = { projectId: string; roomId: string; players: readonly [string, string]; board: readonly ("X" | "O" | null)[]; turn: string; status: "active" | "paused"; turnDeadlineAt?: number };
export { pauseTicTacToeTurn } from "./pause-tic-tac-toe-turn.js";
export { resumePausedTicTacToeTurn } from "./resume-tic-tac-toe-turn.js";
export type CommittedTicTacToeMove = { revision: number; playerId: string; cell: number };
export function reconstructTicTacToeHistory(initialState: TicTacToeState, committedMoves: readonly CommittedTicTacToeMove[]): TicTacToeState {
  return committedMoves.reduce((state, outcome, expectedRevision) => {
    if (outcome.revision !== expectedRevision || !outcome.playerId.trim()) throw new Error("invalid committed game history");
    try {
      return validateTicTacToeMove(state, outcome.playerId, outcome.cell);
    } catch {
      throw new Error("invalid committed game history");
    }
  }, initialState);
}
export const TIC_TAC_TOE_RULES_SOURCE = "package main\n// Brain Tic-tac-toe rules\n";
export function publishTicTacToeRules(compile: (source: string) => string): string {
  const artifact = compile(TIC_TAC_TOE_RULES_SOURCE);
  if (!artifact) throw new Error("rules compilation failed");
  return artifact;
}
export function authorizedTicTacToeState(state: TicTacToeState, projectId: string, playerId: string): TicTacToeState {
  if (state.projectId !== projectId || !state.players.includes(playerId)) throw new Error("game authorization required");
  return { ...state, players: [...state.players] as [string, string], board: [...state.board] };
}
export function authorizeTicTacToeCommand(state: TicTacToeState, projectId: string, playerId: string, operation: string): boolean {
  authorizedTicTacToeState(state, projectId, playerId);
  if (operation !== "move") throw new Error("game command not authorized");
  return true;
}
export class CommandDeduplicator {
  private readonly commands = new Set<string>();
  accept(commandId: string): boolean {
    if (!commandId.trim()) throw new Error("command id required");
    if (this.commands.has(commandId)) return false;
    this.commands.add(commandId);
    return true;
  }
}
export function validateTicTacToeCommandPayload(payload: unknown): { cell: number } {
  if (!payload || typeof payload !== "object" || Object.keys(payload).length !== 1 || typeof (payload as { cell?: unknown }).cell !== "number" || !Number.isInteger((payload as { cell: number }).cell) || (payload as { cell: number }).cell < 0 || (payload as { cell: number }).cell > 8) throw new Error("command payload mismatch");
  return { cell: (payload as { cell: number }).cell };
}
export function createTicTacToeCoordinator(state: TicTacToeState) { const runtime = createTicTacToeRuntime(state); const revision = new TicTacToeRevisionReader(runtime); const mover = new TicTacToeMover(runtime); return { revision: revision.read.bind(revision), move: mover.move.bind(mover) }; }
export { TicTacToeRevisionReader } from "./tic-tac-toe-revision-reader.js";
export { TicTacToeMover } from "./tic-tac-toe-mover.js";
export function serializeTicTacToeState(state: TicTacToeState): string { return JSON.stringify(state); }
export function reconnectTicTacToe(serialized: string, projectId: string, playerId: string): TicTacToeState {
  let parsed: unknown;
  try { parsed = JSON.parse(serialized); } catch { throw new Error("invalid persisted game"); }
  const state = parsed as Partial<TicTacToeState>;
  if (state.projectId !== projectId || !Array.isArray(state.players) || !state.players.includes(playerId)) throw new Error("game authorization required");
  validateTicTacToeRuleResult({ board: state.board, turn: state.turn, status: state.status });
  return { projectId: state.projectId, roomId: state.roomId!, players: [...state.players] as [string, string], board: [...state.board!] as ("X" | "O" | null)[], turn: state.turn!, status: state.status! };
}
export function hiddenTicTacToeView<T>(state: TicTacToeState, playerId: string, privateState: Readonly<Record<string, T>>): TicTacToeState & { privateState?: T } {
  const view = { ...state, players: [...state.players] as [string, string], board: [...state.board] };
  return { ...view, privateState: state.players.includes(playerId) ? privateState[playerId] : undefined };
}
