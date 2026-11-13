import type { AssignmentDecision, MatchmakingTicket } from "./types.js";
import type { MatchmakingConfigReader } from "./matchmaking-config-reader.js";
import type { MatchmakingConfigStatus } from "./matchmaking-config-status.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class MatchmakingQueue {
  private sequence = 0;
  private decisionSequence = 0;
  private readonly queues = new Map<string, MatchmakingTicket[]>();
  constructor(private readonly configs?: { reader: MatchmakingConfigReader; status: MatchmakingConfigStatus }) {}
  enqueue(projectId: string, scope: string, playerId: string, now = Date.now()): MatchmakingTicket {
    validateProjectId(projectId);
    if (!scope || !playerId) throw new Error("invalid matchmaking ticket");
    if (this.configs && !this.configs.status.isEnabled(projectId, scope)) throw new Error("matchmaking scope disabled");
    const key = `${projectId}:${scope}`;
    if ((this.queues.get(key) ?? []).some((ticket) => ticket.playerId === playerId)) throw new Error("active ticket already exists");
    const ttl = this.configs?.reader.get(projectId, scope).ticketTtlSeconds ?? 60;
    const ticket = { id: `${key}:ticket-${++this.sequence}`, projectId, scope, playerId, createdAt: now, expiresAt: now + ttl * 1000 };
    this.queues.set(key, [...(this.queues.get(key) ?? []), ticket]);
    return { ...ticket };
  }
  peek(projectId: string, scope: string): readonly MatchmakingTicket[] { validateProjectId(projectId); return [...(this.queues.get(`${projectId}:${scope}`) ?? [])]; }
  match(projectId: string, scope: string, matchSize: number): MatchmakingTicket[] {
    validateProjectId(projectId);
    const key = `${projectId}:${scope}`;
    const queue = this.queues.get(key) ?? [];
    if (matchSize < 1 || queue.length < matchSize) return [];
    const matched = queue.slice(0, matchSize);
    this.queues.set(key, queue.slice(matchSize));
    return matched.map((ticket) => ({ ...ticket }));
  }
  expire(now = Date.now()): MatchmakingTicket[] {
    const expired: MatchmakingTicket[] = [];
    for (const [key, queue] of this.queues) {
      const active = queue.filter((ticket) => { if (ticket.expiresAt <= now) { expired.push({ ...ticket }); return false; } return true; });
      this.queues.set(key, active);
    }
    return expired;
  }
  requeue(ticket: MatchmakingTicket, now = Date.now()): void {
    validateProjectId(ticket.projectId);
    if (ticket.expiresAt <= now) throw new Error("ticket expired");
    const key = `${ticket.projectId}:${ticket.scope}`;
    const queue = this.queues.get(key) ?? [];
    if (queue.some((active) => active.id === ticket.id || active.playerId === ticket.playerId)) throw new Error("active ticket already exists");
    this.queues.set(key, [...queue, { ...ticket }].sort((a, b) => a.createdAt - b.createdAt));
  }
  assign(projectId: string, scope: string, matchSize: number, now = Date.now()): AssignmentDecision | null {
    validateProjectId(projectId);
    const tickets = this.match(projectId, scope, matchSize);
    if (tickets.length !== matchSize) return null;
    return { id: `${projectId}:${scope}:assignment-${++this.decisionSequence}`, projectId, scope, ticketIds: tickets.map((ticket) => ticket.id), createdAt: now };
  }
}
