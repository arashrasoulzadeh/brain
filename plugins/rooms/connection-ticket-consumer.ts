import type { ConnectionTicket } from "./index.js";

export class ConnectionTicketConsumer {
  constructor(private readonly tickets: Map<string, ConnectionTicket>) {}
  consume(token: string, projectId: string, roomId: string, playerId: string, now: number): boolean { const ticket = this.tickets.get(token); if (!ticket || ticket.used || ticket.expiresAt <= now || ticket.projectId !== projectId || ticket.roomId !== roomId || ticket.playerId !== playerId) return false; ticket.used = true; return true; }
}
