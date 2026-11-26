import { type Room, type ConnectionTicket, type TicTacToeState } from "./index.js";
import { consumeOnce, expiresAt, validateProjectId } from "../sdk-typescript/src/index.js";
import { projectRoomID } from "./directory.js";
import { randomUUID } from "node:crypto";

export function authorizedRoomSnapshot(room: Room, projectId: string, playerId: string): Room {
  validateProjectId(projectId, "room membership required");
  if (room.projectId !== projectId || !room.slots.some((slot) => slot.playerId === playerId))
    throw new Error("room membership required");
  return { ...room, slots: [...room.slots] };
}

export function authorizeRoomCommand(
  room: Room,
  projectId: string,
  playerId: string,
  operation: string,
): boolean {
  if (!operation.trim()) throw new Error("operation required");
  authorizedRoomSnapshot(room, projectId, playerId);
  return true;
}

export class ConnectionTicketConsumer {
  constructor(private readonly tickets: Map<string, ConnectionTicket>) {}
  consume(
    token: string,
    projectId: string,
    roomId: string,
    playerId: string,
    now: number,
  ): boolean {
    projectRoomID(projectId, roomId);
    if (!token.trim() || !playerId.trim() || !Number.isFinite(now)) return false;
    return !!consumeOnce(this.tickets.get(token), now, (ticket) =>
      ticket.projectId === projectId && ticket.roomId === roomId && ticket.playerId === playerId,
    );
  }
}

export class ConnectionTicketIssuer {
  constructor(private readonly tickets: Map<string, ConnectionTicket>) {}
  issue(
    projectId: string,
    roomId: string,
    playerId: string,
    now: number,
    ttlSeconds: number,
  ): ConnectionTicket {
    projectRoomID(projectId, roomId);
    if (
      !playerId.trim() ||
      !Number.isFinite(now) ||
      !Number.isFinite(ttlSeconds) ||
      ttlSeconds <= 0
    )
      throw new Error("invalid connection ticket");
    const ticket = {
      token: randomUUID(),
      projectId,
      roomId,
      playerId,
      expiresAt: expiresAt(now, ttlSeconds * 1000, "invalid connection ticket"),
      used: false,
    };
    this.tickets.set(ticket.token, ticket);
    return { ...ticket };
  }
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

export function authorizedTicTacToeState(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
): TicTacToeState {
  validateProjectId(projectId, "game authorization required");
  if (state.projectId !== projectId || !state.players.includes(playerId))
    throw new Error("game authorization required");
  return { ...state, players: [...state.players] as [string, string], board: [...state.board] };
}

export function authorizeTicTacToeCommand(
  state: TicTacToeState,
  projectId: string,
  playerId: string,
  operation: string,
): boolean {
  authorizedTicTacToeState(state, projectId, playerId);
  if (operation !== "move") throw new Error("game command not authorized");
  return true;
}

export function validateTicTacToeCommandPayload(payload: unknown): { cell: number } {
  if (
    !payload ||
    typeof payload !== "object" ||
    Object.keys(payload).length !== 1 ||
    typeof (payload as { cell?: unknown }).cell !== "number" ||
    !Number.isInteger((payload as { cell: number }).cell) ||
    (payload as { cell: number }).cell < 0 ||
    (payload as { cell: number }).cell > 8
  )
    throw new Error("command payload mismatch");
  return { cell: (payload as { cell: number }).cell };
}
