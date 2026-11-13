import { request } from "../sdk-typescript/src/index.js";
import { randomUUID } from "node:crypto";

export function projectRoomID(projectId: string, roomId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(roomId)) throw new Error("invalid project room id");
  return `${projectId}:${roomId}`;
}

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

export class ConnectionTickets {
  private readonly tickets = new Map<string, ConnectionTicket>();
  issue(projectId: string, roomId: string, playerId: string, now: number, ttlSeconds: number): ConnectionTicket {
    projectRoomID(projectId, roomId);
    if (!playerId.trim() || !Number.isFinite(now) || !Number.isFinite(ttlSeconds) || ttlSeconds <= 0) throw new Error("invalid connection ticket");
    const ticket = { token: randomUUID(), projectId, roomId, playerId, expiresAt: now + ttlSeconds * 1000, used: false };
    this.tickets.set(ticket.token, ticket);
    return { ...ticket };
  }
  consume(token: string, projectId: string, roomId: string, playerId: string, now: number): boolean {
    const ticket = this.tickets.get(token);
    if (!ticket || ticket.used || ticket.expiresAt <= now || ticket.projectId !== projectId || ticket.roomId !== roomId || ticket.playerId !== playerId) return false;
    ticket.used = true;
    return true;
  }
}
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

export function canTransitionRoom(from: RoomState, to: RoomState): boolean {
  return ({ waiting: ["matching", "closed"], matching: ["active", "waiting", "closed"], active: ["finished", "closed"], finished: ["closed"], closed: [] } as Record<RoomState, RoomState[]>)[from].includes(to);
}

export function validateMembership(projectId: string, roomId: string, slots: readonly PlayerSlot[], maxPlayers: number): void {
  projectRoomID(projectId, roomId);
  if (maxPlayers < 1 || maxPlayers > 64 || slots.length > maxPlayers) throw new Error("invalid player slots");
  const seen = new Set<number>();
  for (const player of slots) {
    if (!player.playerId || player.slot < 0 || player.slot >= maxPlayers || seen.has(player.slot)) throw new Error("invalid player slots");
    seen.add(player.slot);
  }
}
export type TicTacToeState = { projectId: string; roomId: string; players: readonly [string, string]; board: readonly ("X" | "O" | null)[]; turn: string; status: "active" };
export function initializeTicTacToe(projectId: string, roomId: string, firstPlayer: string, secondPlayer: string): TicTacToeState {
  projectRoomID(projectId, roomId);
  if (!firstPlayer.trim() || !secondPlayer.trim() || firstPlayer === secondPlayer) throw new Error("invalid tic-tac-toe players");
  return { projectId, roomId, players: [firstPlayer, secondPlayer], board: Array(9).fill(null), turn: firstPlayer, status: "active" };
}
export function validateTicTacToeMove(state: TicTacToeState, playerId: string, cell: number): TicTacToeState {
  if (state.status !== "active" || !state.players.includes(playerId) || !Number.isInteger(cell) || cell < 0 || cell >= 9) throw new Error("invalid tic-tac-toe move");
  if (state.turn !== playerId) throw new Error("out-of-turn move");
  if (state.board[cell] !== null) throw new Error("cell occupied");
  const mark = state.players[0] === playerId ? "X" : "O";
  const board = [...state.board] as ("X" | "O" | null)[];
  board[cell] = mark;
  return { ...state, board, turn: state.players.find((player) => player !== playerId)! };
}
export function ticTacToeWinner(state: TicTacToeState): string | undefined {
  const lines = [[0, 1, 2], [3, 4, 5], [6, 7, 8], [0, 3, 6], [1, 4, 7], [2, 5, 8], [0, 4, 8], [2, 4, 6]];
  for (const [a, b, c] of lines) if (state.board[a] && state.board[a] === state.board[b] && state.board[a] === state.board[c]) return state.players[state.board[a] === "X" ? 0 : 1];
  return undefined;
}
export function ticTacToeDraw(state: TicTacToeState): boolean {
  return !ticTacToeWinner(state) && state.board.every((cell) => cell !== null);
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

export function roomRequest(projectId: string, pluginId: string, version: string, operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result", roomId: string, payload: unknown = {}) {
  return request(`rooms-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

