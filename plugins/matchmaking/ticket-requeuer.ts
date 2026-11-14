import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { MatchmakingTicket } from "./types.js";
import type { QueueRuntime } from "./queue-runtime.js";
export class TicketRequeuer {
  constructor(private readonly runtime: QueueRuntime) {}
  requeue(ticket: MatchmakingTicket, now = Date.now()): void {
    validateProjectId(ticket.projectId);
    if (ticket.expiresAt <= now) throw new Error("ticket expired");
    const key = `${ticket.projectId}:${ticket.scope}`;
    const queue = this.runtime.queues.get(key) ?? [];
    if (queue.some((active) => active.id === ticket.id || active.playerId === ticket.playerId))
      throw new Error("active ticket already exists");
    this.runtime.queues.set(
      key,
      [...queue, { ...ticket }].sort((a, b) => a.createdAt - b.createdAt),
    );
  }
}
