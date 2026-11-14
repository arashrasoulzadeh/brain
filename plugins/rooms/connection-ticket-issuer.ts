import { randomUUID } from "node:crypto";
import type { ConnectionTicket } from "./index.js";
import { projectRoomID } from "./project-room-id.js";

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
