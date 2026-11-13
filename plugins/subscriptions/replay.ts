import type { SubscriptionEvent } from "./types.js";

export class SubscriptionReplay<T> {
  private sequence = 0;
  private readonly events: SubscriptionEvent<T>[] = [];
  append(value: T): SubscriptionEvent<T> { const event = { sequence: ++this.sequence, value }; this.events.push(event); return { ...event }; }
  since(sequence: number): SubscriptionEvent<T>[] { if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid subscription cursor"); return this.events.filter((event) => event.sequence > sequence).map((event) => ({ ...event })); }
}
