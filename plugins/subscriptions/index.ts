import { pluginRequest } from "../sdk-typescript/src/index.js";
import { randomUUID } from "node:crypto";

export function subscriptionRequest(projectId: string, pluginId: string, version: string, operation: "subscriptions.open" | "subscriptions.receive" | "subscriptions.resync", channel: string, payload: unknown = {}) {
  return pluginRequest("subscriptions", projectId, pluginId, version, operation, { channel, ...(payload as Record<string, unknown>) }, channel);
}

export type SubscriptionTicket = { ticket: string; projectId: string; channel: string; operation: string; expiresAt: number };

export class SubscriptionTicketStore {
  private readonly tickets = new Map<string, SubscriptionTicket>();
  issue(projectId: string, channel: string, operation = "subscriptions.open", now = Date.now(), ttlMs = 30_000): SubscriptionTicket {
    if (!projectId || !channel || !operation || !Number.isFinite(now) || ttlMs <= 0) throw new Error("invalid subscription ticket");
    const ticket = { ticket: randomUUID(), projectId, channel, operation, expiresAt: now + ttlMs };
    this.tickets.set(ticket.ticket, ticket);
    return { ...ticket };
  }
  validate(ticket: string, projectId: string, channel: string, operation = "subscriptions.open", now = Date.now()): SubscriptionTicket {
    const found = this.tickets.get(ticket);
    if (!found || found.projectId !== projectId || found.channel !== channel || found.operation !== operation || now >= found.expiresAt) throw new Error("invalid subscription ticket");
    return { ...found };
  }
  consume(ticket: string, projectId: string, channel: string, operation = "subscriptions.open", now = Date.now()): SubscriptionTicket {
    const found = this.validate(ticket, projectId, channel, operation, now);
    this.tickets.delete(ticket);
    return found;
  }
}

export function validateBrowserOrigin(origin: string, allowedOrigins: readonly string[]): void {
  if (!origin || !allowedOrigins.includes(origin)) throw new Error("origin denied");
  let parsed: URL;
  try { parsed = new URL(origin); } catch { throw new Error("origin denied"); }
  if (!['http:', 'https:'].includes(parsed.protocol) || parsed.username || parsed.password || parsed.pathname !== '/' || parsed.search || parsed.hash) throw new Error("origin denied");
}

export class SubscriptionAuthorizer {
  private readonly grants = new Set<string>();
  grant(projectId: string, channel: string, actorId: string): void {
    if (!projectId || !channel || !actorId) throw new Error("invalid subscription grant");
    this.grants.add(`${projectId}:${channel}:${actorId}`);
  }
  authorize(projectId: string, channel: string, actorId: string): void {
    if (!this.grants.has(`${projectId}:${channel}:${actorId}`)) throw new Error("subscription denied");
  }
  invalidateOwner(projectId: string, oldOwnerId: string): void {
    for (const grant of this.grants) if (grant.startsWith(`${projectId}:`) && grant.endsWith(`:${oldOwnerId}`)) this.grants.delete(grant);
  }
}

export class InitialFetchBuffer<T> {
  private fetching = true;
  private readonly pending: T[] = [];
  push(event: T): void { if (this.fetching) this.pending.push(event); }
  complete(snapshot: readonly T[]): T[] { this.fetching = false; return [...snapshot, ...this.pending.splice(0)]; }
}

export class SlowConsumerQueue<T> {
  private readonly items: T[] = [];
  constructor(private readonly capacity: number) { if (!Number.isInteger(capacity) || capacity < 1) throw new Error("invalid queue capacity"); }
  push(item: T): boolean { if (this.items.length >= this.capacity) return false; this.items.push(item); return true; }
  shift(): T | undefined { return this.items.shift(); }
  get size(): number { return this.items.length; }
}

export type SubscriptionEvent<T> = { sequence: number; value: T };
export class SubscriptionReplay<T> {
  private sequence = 0;
  private readonly events: SubscriptionEvent<T>[] = [];
  append(value: T): SubscriptionEvent<T> { const event = { sequence: ++this.sequence, value }; this.events.push(event); return { ...event }; }
  since(sequence: number): SubscriptionEvent<T>[] { if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid subscription cursor"); return this.events.filter((event) => event.sequence > sequence).map((event) => ({ ...event })); }
}

