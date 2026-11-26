import { requireProjectInput, requiredString, type Observation, type PluginContext } from "../sdk-typescript/src/index.js";
import { SubscriptionTicketIssuer, SubscriptionTicketValidator } from "./authorization.js";
import { SubscriptionEventAppender, SubscriptionEventReader } from "./stream.js";
import type { SubscriptionEvent, SubscriptionTicket } from "./types.js";

type RoomConnection = Readonly<{ roomId: string; reconnected: boolean }>;
type Stream = { sequence: number; events: SubscriptionEvent<RoomConnection>[] };

const channelName = "rooms.connected";
const retention = 256;
const tickets = new Map<string, SubscriptionTicket>();
const streams = new Map<string, Stream>();
const issuer = new SubscriptionTicketIssuer(tickets);
const validator = new SubscriptionTicketValidator(tickets);

/** Operator-only polling adapter for the currently wired observation channel. */
export function open(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const channel = requireChannel(payload);
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
  const ticket = validator.validate(requiredString(payload.ticket, "ticket required"), context.project.id, channel);
  tickets.delete(ticket.ticket);
  context.plugins.observe.fire("closed", {});
  return { closed: true };
}

export function observeRoomConnected(context: PluginContext, input: unknown) {
  requireProjectInput(context, input);
  const event = input as Observation<RoomConnection>;
  if (event.pluginId !== "rooms" || event.event !== "connected" || !event.payload ||
      typeof event.payload.roomId !== "string" || typeof event.payload.reconnected !== "boolean")
    throw new Error("invalid room observation");
  const stream = streamFor(context.project.id);
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
  validator.validate(requiredString(payload.ticket, "ticket required"), context.project.id, requireChannel(payload));
  return streamFor(context.project.id);
}

function streamFor(projectId: string): Stream {
  let stream = streams.get(projectId);
  if (!stream) {
    stream = { sequence: 0, events: [] };
    streams.set(projectId, stream);
  }
  return stream;
}
