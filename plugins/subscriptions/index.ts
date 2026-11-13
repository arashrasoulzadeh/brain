import { request } from "../sdk-typescript/src/index.js";
import { randomUUID } from "node:crypto";

export function subscriptionRequest(projectId: string, pluginId: string, version: string, operation: "subscriptions.open" | "subscriptions.receive" | "subscriptions.resync", channel: string, payload: unknown = {}) {
  return request(`subscriptions-${operation}-${projectId}-${channel}`, { protocol: 1, projectId, pluginId, version, operation, payload: { channel, ...(payload as Record<string, unknown>) } });
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
}

