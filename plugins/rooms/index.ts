import { randomUUID } from "node:crypto";
import { projectRoomID } from "./project-room-id.js";
import { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
import { validateTicTacToeRuleResult } from "./validate-tic-tac-toe-rule-result.js";
import { createRoomRuntime } from "./room-runtime.js";
import { RoomCreator } from "./room-creator.js";
import { RoomReader } from "./room-reader.js";
import { RoomMembershipReader } from "./room-membership-reader.js";
import { RoomJoiner } from "./room-joiner.js";
import { RoomLeaver } from "./room-leaver.js";
import { RoomEventReader } from "./room-event-reader.js";
import { createTicTacToeRuntime } from "./tic-tac-toe-runtime.js";
import { TicTacToeRevisionReader } from "./tic-tac-toe-revision-reader.js";
import { TicTacToeMover } from "./tic-tac-toe-mover.js";

export { roomRequest } from "./room-request.js";

export { projectRoomID } from "./project-room-id.js";

export type PlayerSlot = { playerId: string; slot: number };

export type RoomState = "waiting" | "matching" | "active" | "finished" | "closed";

export type Room = { id: string; projectId: string; state: RoomState; maxPlayers: number; slots: PlayerSlot[] };
export type ConnectionTicket = { token: string; projectId: string; roomId: string; playerId: string; expiresAt: number; used: boolean };

export function authorizedRoomSnapshot(room: Room, projectId: string, playerId: string): Room {
  if (room.projectId !== projectId || !room.slots.some((slot) => slot.playerId === playerId)) throw new Error("room membership required");
  return { ...room, slots: [...room.slots] };
}

export function authorizeRoomCommand(room: Room, projectId: string, playerId: string, operation: string): boolean {
  if (!operation.trim()) throw new Error("operation required");
  authorizedRoomSnapshot(room, projectId, playerId);
  return true;
}

export { ConnectionTicketIssuer } from "./connection-ticket-issuer.js";
export { ConnectionTicketConsumer } from "./connection-ticket-consumer.js";
export type RoomEvent = { type: "player.joined" | "player.left"; projectId: string; roomId: string; playerId: string; sequence: number };
export type MembershipStore = { save(roomId: string, slots: readonly PlayerSlot[]): void };
export function createRoomDirectory(store?: MembershipStore, admission?: () => void) { const runtime = createRoomRuntime(store, admission); const creator = new RoomCreator(runtime); const reader = new RoomReader(runtime); const membership = new RoomMembershipReader(runtime); const joiner = new RoomJoiner(runtime); const leaver = new RoomLeaver(runtime); const events = new RoomEventReader(runtime); return { create: creator.create.bind(creator), get: reader.get.bind(reader), isMember: membership.isMember.bind(membership), reconnect: (projectId: string, roomId: string, playerId: string) => { const room = reader.get(projectId, roomId); if (!room.slots.some((slot) => slot.playerId === playerId)) throw new Error("player not in room"); return room; }, join: joiner.join.bind(joiner), leave: leaver.leave.bind(leaver), events: events.events.bind(events), eventsSince: (projectId: string, roomId: string, sequence: number) => { reader.get(projectId, roomId); if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid event sequence"); return events.events(projectId, roomId).filter((event) => event.sequence > sequence); } }; }

export { canTransitionRoom } from "./room-transition.js";
export { validateMembership } from "./validate-membership.js";
export { initializeTicTacToe } from "./initialize-tic-tac-toe.js";
export { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
export { expireTicTacToeTurn } from "./expire-tic-tac-toe-turn.js";
export { ticTacToeWinner } from "./tic-tac-toe-winner.js";
export { ticTacToeDraw } from "./tic-tac-toe-draw.js";
export { validateTicTacToeRuleResult } from "./validate-tic-tac-toe-rule-result.js";
export type TicTacToeState = { projectId: string; roomId: string; players: readonly [string, string]; board: readonly ("X" | "O" | null)[]; turn: string; status: "active"; turnDeadlineAt?: number };
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
