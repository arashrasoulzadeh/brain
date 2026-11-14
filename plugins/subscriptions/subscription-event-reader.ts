import type { SubscriptionEvent } from "./types.js";

export class SubscriptionEventReader<T> {
  constructor(private readonly events: readonly SubscriptionEvent<T>[]) {}
  since(sequence: number): SubscriptionEvent<T>[] {
    if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid subscription cursor");
    return this.events.filter((event) => event.sequence > sequence).map((event) => ({ ...event }));
  }
}
