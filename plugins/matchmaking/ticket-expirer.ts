import type { QueueRuntime } from "./queue-runtime.js";
import type { MatchmakingTicket } from "./types.js";
export class TicketExpirer { constructor(private readonly runtime: QueueRuntime) {} expire(now = Date.now()): MatchmakingTicket[] { const expired: MatchmakingTicket[] = []; for (const [key, queue] of this.runtime.queues) { const active = queue.filter((ticket) => { if (ticket.expiresAt <= now) { expired.push({ ...ticket }); return false; } return true; }); this.runtime.queues.set(key, active); } return expired; } }
