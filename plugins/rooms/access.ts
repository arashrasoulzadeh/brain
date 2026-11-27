import { type Room, type ConnectionTicket } from "./index.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
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
    const ticket = this.tickets.get(token);
    if (
      !ticket ||
      ticket.used ||
      ticket.expiresAt <= now ||
      ticket.projectId !== projectId ||
      ticket.roomId !== roomId ||
      ticket.playerId !== playerId
    )
      return false;
    ticket.used = true;
    return true;
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
      expiresAt: now + ttlSeconds * 1000,
      used: false,
    };
    this.tickets.set(ticket.token, ticket);
    return { ...ticket };
  }
}

