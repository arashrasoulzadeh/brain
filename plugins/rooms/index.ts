import { randomUUID } from "node:crypto";
import { projectRoomID } from "./project-room-id.js";
import { validateTicTacToeMove } from "./validate-tic-tac-toe-move.js";
import { validateTicTacToeRuleResult } from "./validate-tic-tac-toe-rule-result.js";

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

export class RoomDirectory {
  private readonly rooms = new Map<string, Room>();
  private readonly roomEvents = new Map<string, RoomEvent[]>();
  private sequence = 0;
  constructor(private readonly store: MembershipStore = { save: () => {} }) {}
  create(projectId: string, roomId: string, maxPlayers: number): Room {
    const id = projectRoomID(projectId, roomId);
    if (maxPlayers < 1 || maxPlayers > 64) throw new Error("invalid max players");
    if (this.rooms.has(id)) throw new Error("room already exists");
    const room: Room = { id, projectId, state: "waiting", maxPlayers, slots: [] };
    this.rooms.set(id, room);
    this.roomEvents.set(id, []);
    this.store.save(id, []);
    return { ...room, slots: [] };
  }
  get(projectId: string, roomId: string): Room {
    const room = this.rooms.get(projectRoomID(projectId, roomId));
    if (!room) throw new Error("room not found");
    return { ...room, slots: [...room.slots] };
  }
  isMember(projectId: string, roomId: string, playerId: string): boolean {
    const room = this.rooms.get(projectRoomID(projectId, roomId));
    return room?.slots.some((player) => player.playerId === playerId) ?? false;
  }
  reconnect(projectId: string, roomId: string, playerId: string): Room {
    const room = this.get(projectId, roomId);
    if (!room.slots.some((player) => player.playerId === playerId)) throw new Error("player not in room");
    return room;
  }
  join(projectId: string, roomId: string, playerId: string, slot: number): Room {
    const key = projectRoomID(projectId, roomId);
    const room = this.rooms.get(key);
    if (!room) throw new Error("room not found");
    if (room.state !== "waiting" && room.state !== "matching") throw new Error("room not joinable");
    if (!playerId || slot < 0 || slot >= room.maxPlayers || room.slots.some((player) => player.slot === slot || player.playerId === playerId)) throw new Error("invalid membership");
    room.slots.push({ playerId, slot });
    this.store.save(key, room.slots);
    this.roomEvents.get(key)?.push({ type: "player.joined", projectId, roomId, playerId, sequence: ++this.sequence });
    return { ...room, slots: [...room.slots] };
  }
  leave(projectId: string, roomId: string, playerId: string): void {
    const key = projectRoomID(projectId, roomId);
    const room = this.rooms.get(key);
    if (!room || !room.slots.some((player) => player.playerId === playerId)) throw new Error("player not in room");
    room.slots = room.slots.filter((player) => player.playerId !== playerId);
    this.store.save(key, room.slots);
    this.roomEvents.get(key)?.push({ type: "player.left", projectId, roomId, playerId, sequence: ++this.sequence });
  }
  events(projectId: string, roomId: string): readonly RoomEvent[] { return [...(this.roomEvents.get(projectRoomID(projectId, roomId)) ?? [])]; }
  eventsSince(projectId: string, roomId: string, sequence: number): readonly RoomEvent[] {
    this.get(projectId, roomId);
    if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid event sequence");
    return this.events(projectId, roomId).filter((event) => event.sequence > sequence);
  }
}

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
export class TicTacToeCoordinator {
  private current: TicTacToeState;
  private version = 0;
  constructor(state: TicTacToeState) { this.current = { ...state, players: [...state.players] as [string, string], board: [...state.board] }; }
  revision(): number { return this.version; }
  move(playerId: string, cell: number, expectedRevision: number): TicTacToeState {
    if (expectedRevision !== this.version) throw new Error("stale game revision");
    this.current = validateTicTacToeMove(this.current, playerId, cell);
    this.version++;
    return { ...this.current, players: [...this.current.players] as [string, string], board: [...this.current.board] };
  }
}
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
