import { requireProjectInput, requireOperator, requiredString, type Observation, type PluginContext } from "../sdk-typescript/src/index.js";
import { SubscriptionTicketIssuer, SubscriptionTicketValidator } from "./authorization.js";
import { SubscriptionEventAppender, SubscriptionEventReader } from "./stream.js";
import type { SubscriptionEvent, SubscriptionTicket } from "./types.js";

type RoomConnection = Readonly<{ roomId: string; reconnected: boolean }>;
type Stream = { sequence: number; events: SubscriptionEvent<RoomConnection>[]; seen: Set<string> };

const channelName = "rooms.connected";
const retention = 256;
const tickets = new Map<string, SubscriptionTicket>();
const streams = new Map<string, Stream>();
const issuer = new SubscriptionTicketIssuer(tickets);
const validator = new SubscriptionTicketValidator(tickets);

type StoredSubscription = { stream: { sequence: number; events: SubscriptionEvent<RoomConnection>[]; seen: string[] }; allocation?: SubscriptionTicket; ticket?: SubscriptionTicket; count?: number };
function storedSubscription(payload: Record<string, unknown>): StoredSubscription | undefined {
  if (!Object.hasOwn(payload, "_brainSubscriptions")) return undefined;
  const stored = payload._brainSubscriptions as StoredSubscription;
  if (!stored?.stream || !Number.isSafeInteger(stored.stream.sequence) || !Array.isArray(stored.stream.events) || !Array.isArray(stored.stream.seen))
    throw new Error("invalid stored subscription");
  return stored;
}

function requestStream(context: PluginContext, payload: Record<string, unknown>): Stream {
  const stored = storedSubscription(payload);
  return stored ? { sequence: stored.stream.sequence, events: stored.stream.events, seen: new Set(stored.stream.seen) } : streamFor(context.project.id);
}

function requestValidator(payload: Record<string, unknown>): SubscriptionTicketValidator {
  const stored = storedSubscription(payload);
  if (!stored) return validator;
  const ticket = stored.ticket;
  return new SubscriptionTicketValidator(new Map(ticket ? [[ticket.ticket, ticket]] : []));
}

/** Operator-only polling adapter for the currently wired observation channel. */
export function open(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const channel = requireChannel(payload);
  const stored = storedSubscription(payload);
  if (stored) {
    const ticket = stored.allocation;
    if (!ticket || ticket.projectId !== context.project.id || ticket.channel !== channel || ticket.operation !== "subscriptions.open" || !ticket.ticket || ticket.expiresAt <= Date.now()) throw new Error("invalid stored ticket allocation");
    context.plugins.observe.fire("opened", {});
    return { ...ticket, cursor: stored.stream.sequence };
  }
  const now = Date.now();
  for (const [id, ticket] of tickets) if (ticket.expiresAt <= now) tickets.delete(id);
  if (tickets.size >= 4096) throw new Error("subscription capacity reached");
  const ticket = issuer.issue(context.project.id, channel, "subscriptions.open", now);
  context.plugins.observe.fire("opened", {});
  return { ...ticket, cursor: streamFor(context.project.id).sequence };
}

export function receive(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const stream = authorizedStream(context, payload);
  const cursor = payload.cursor ?? 0;
  if (typeof cursor !== "number" || !Number.isSafeInteger(cursor) || cursor < 0 || cursor > stream.sequence)
    throw new Error("invalid subscription cursor");
  const earliest = stream.events[0]?.sequence ?? stream.sequence + 1;
  if (cursor < earliest - 1) throw new Error("subscription resync required");
  return { events: new SubscriptionEventReader(stream.events).since(cursor), cursor: stream.sequence };
}

export function resync(context: PluginContext, input: unknown) {
  const stream = authorizedStream(context, requireProjectInput(context, input));
  return { events: new SubscriptionEventReader(stream.events).since(0), cursor: stream.sequence, retainedFrom: stream.events[0]?.sequence ?? stream.sequence + 1 };
}

export function close(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const channel = requireChannel(payload);
  const ticket = requestValidator(payload).validate(requiredString(payload.ticket, "ticket required"), context.project.id, channel);
  if (!storedSubscription(payload)) tickets.delete(ticket.ticket);
  context.plugins.observe.fire("closed", {});
  return { closed: true };
}

/** Redacted operator inspection and revocation without exporting ticket credentials. */
export function inspect(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const channel = requireChannel(payload);
  requireOperator(context, "subscriptions.inspect", channel);
  const stored = storedSubscription(payload);
  if (stored) return { channel, activeTickets: stored.count ?? 0, cursor: stored.stream.sequence, retainedEvents: stored.stream.events.length };
  const now = Date.now();
  const stream = streams.get(context.project.id);
  const activeTickets = [...tickets.values()].filter(ticket => ticket.projectId === context.project.id && ticket.channel === channel && ticket.expiresAt > now).length;
  return { channel, activeTickets, cursor: stream?.sequence ?? 0, retainedEvents: stream?.events.length ?? 0 };
}

export function revokeChannel(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const channel = requireChannel(payload);
  requireOperator(context, "subscriptions.revoke", channel);
  const stored = storedSubscription(payload);
  if (stored) {
    const revoked = stored.count ?? 0;
    if (revoked) context.plugins.observe.fire("closed", {});
    return { channel, revoked };
  }
  let revoked = 0;
  for (const [key, ticket] of tickets) {
    if (ticket.projectId === context.project.id && ticket.channel === channel) {
      tickets.delete(key);
      revoked++;
    }
  }
  if (revoked) context.plugins.observe.fire("closed", {});
  return { channel, revoked };
}

export function observeRoomConnected(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const event = input as Observation<RoomConnection>;
  if (event.pluginId !== "rooms" || event.event !== "connected" || !event.payload ||
      typeof event.payload.roomId !== "string" || typeof event.payload.reconnected !== "boolean")
    throw new Error("invalid room observation");
  const stream = requestStream(context, payload);
  if (event.id && stream.seen.has(event.id)) return { recorded: true };
  if (event.id) {
    stream.seen.add(event.id);
    if (stream.seen.size > retention * 2) {
      const oldest = stream.seen.values().next().value;
      if (oldest) stream.seen.delete(oldest);
    }
  }
  new SubscriptionEventAppender(stream.events, () => ++stream.sequence).append({
    roomId: event.payload.roomId, reconnected: event.payload.reconnected,
  });
  if (stream.events.length > retention) stream.events.splice(0, stream.events.length - retention);
  return { recorded: true };
}

function requireChannel(payload: Record<string, unknown>): string {
  if (payload.channel !== channelName) throw new Error("subscription channel unavailable");
  return channelName;
}

function authorizedStream(context: PluginContext, payload: Record<string, unknown>): Stream {
  requestValidator(payload).validate(requiredString(payload.ticket, "ticket required"), context.project.id, requireChannel(payload));
  return requestStream(context, payload);
}

function streamFor(projectId: string): Stream {
  let stream = streams.get(projectId);
  if (!stream) {
    stream = { sequence: 0, events: [], seen: new Set<string>() };
    streams.set(projectId, stream);
  }
  return stream;
}
