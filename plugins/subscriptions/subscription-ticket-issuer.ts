import { randomUUID } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { SubscriptionTicket } from "./types.js";

export class SubscriptionTicketIssuer {
  constructor(private readonly tickets: Map<string, SubscriptionTicket>) {}
  issue(
    projectId: string,
    channel: string,
    operation = "subscriptions.open",
    now = Date.now(),
    ttlMs = 30_000,
  ): SubscriptionTicket {
    validateProjectId(projectId, "invalid subscription ticket");
    if (!channel || !operation || !Number.isFinite(now) || ttlMs <= 0)
      throw new Error("invalid subscription ticket");
    const ticket = { ticket: randomUUID(), projectId, channel, operation, expiresAt: now + ttlMs };
    this.tickets.set(ticket.ticket, ticket);
    return { ...ticket };
  }
}
