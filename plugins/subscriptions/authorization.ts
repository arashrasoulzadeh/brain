import { randomUUID } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { SubscriptionTicket } from "./types.js";

export function validateBrowserOrigin(origin: string, allowedOrigins: readonly string[]): void {
  if (!origin || !allowedOrigins.includes(origin)) throw new Error("origin denied");
  let parsed: URL;
  try {
    parsed = new URL(origin);
  } catch {
    throw new Error("origin denied");
  }
  if (
    !["http:", "https:"].includes(parsed.protocol) ||
    parsed.username ||
    parsed.password ||
    parsed.pathname !== "/" ||
    parsed.search ||
    parsed.hash
  )
    throw new Error("origin denied");
}

export class SubscriptionAuthorization {
  constructor(private readonly grants: ReadonlySet<string>) {}
  authorize(projectId: string, channel: string, actorId: string): void {
    validateProjectId(projectId, "subscription denied");
    if (!channel || !actorId) throw new Error("subscription denied");
    if (!this.grants.has(`${projectId}:${channel}:${actorId}`))
      throw new Error("subscription denied");
  }
}

export class SubscriptionGrantStore {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, channel: string, actorId: string): void {
    validateProjectId(projectId, "invalid subscription grant");
    if (!channel || !actorId) throw new Error("invalid subscription grant");
    this.grants.add(`${projectId}:${channel}:${actorId}`);
  }
}

export class SubscriptionOwnerInvalidator {
  constructor(private readonly grants: Set<string>) {}
  invalidate(projectId: string, oldOwnerId: string): void {
    validateProjectId(projectId, "invalid subscription owner");
    if (!oldOwnerId) throw new Error("invalid subscription owner");
    for (const grant of this.grants)
      if (grant.startsWith(`${projectId}:`) && grant.endsWith(`:${oldOwnerId}`))
        this.grants.delete(grant);
  }
}

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

export class SubscriptionTicketValidator {
  constructor(private readonly tickets: ReadonlyMap<string, SubscriptionTicket>) {}
  validate(
    ticket: string,
    projectId: string,
    channel: string,
    operation = "subscriptions.open",
    now = Date.now(),
  ): SubscriptionTicket {
    validateProjectId(projectId, "invalid subscription ticket");
    if (!ticket || !channel || !operation || !Number.isFinite(now))
      throw new Error("invalid subscription ticket");
    const found = this.tickets.get(ticket);
    if (
      !found ||
      found.projectId !== projectId ||
      found.channel !== channel ||
      found.operation !== operation ||
      now >= found.expiresAt
    )
      throw new Error("invalid subscription ticket");
    return { ...found };
  }
}

export class SubscriptionTicketConsumer {
  private readonly validator: SubscriptionTicketValidator;
  constructor(private readonly tickets: Map<string, SubscriptionTicket>) {
    this.validator = new SubscriptionTicketValidator(tickets);
  }
  consume(
    ticket: string,
    projectId: string,
    channel: string,
    operation = "subscriptions.open",
    now = Date.now(),
  ): SubscriptionTicket {
    const found = this.validator.validate(ticket, projectId, channel, operation, now);
    this.tickets.delete(ticket);
    return found;
  }
}
