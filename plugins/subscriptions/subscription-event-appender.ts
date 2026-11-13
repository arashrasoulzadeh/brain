import type { SubscriptionEvent } from "./types.js";

export class SubscriptionEventAppender<T> {
  constructor(private readonly events: SubscriptionEvent<T>[], private readonly nextSequence: () => number) {}
  append(value: T): SubscriptionEvent<T> { const event = { sequence: this.nextSequence(), value }; this.events.push(event); return { ...event }; }
}
