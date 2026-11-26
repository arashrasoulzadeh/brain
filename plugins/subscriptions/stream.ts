import { createManifestRequest } from "../sdk-typescript/src/index.js";
import type { SubscriptionEvent } from "./types.js";

export class InitialFetchWriter<T> {
  constructor(
    private readonly pending: T[],
    private readonly isFetching: () => boolean,
  ) {}
  push(event: T): void {
    if (this.isFetching()) this.pending.push(event);
  }
}

export class InitialFetchCompleter<T> {
  constructor(
    private readonly pending: T[],
    private readonly stopFetching: () => void,
  ) {}
  complete(snapshot: readonly T[]): T[] {
    this.stopFetching();
    return [...snapshot, ...this.pending.splice(0)];
  }
}

export class SlowConsumerPusher<T> {
  constructor(
    private readonly items: T[],
    private readonly capacity: number,
  ) {
    if (!Number.isInteger(capacity) || capacity < 1) throw new Error("invalid queue capacity");
  }
  push(item: T): boolean {
    if (this.items.length >= this.capacity) return false;
    this.items.push(item);
    return true;
  }
}

export class SlowConsumerShifter<T> {
  constructor(private readonly items: T[]) {}
  shift(): T | undefined {
    return this.items.shift();
  }
}

export class SlowConsumerSize {
  constructor(private readonly items: readonly unknown[]) {}
  size(): number {
    return this.items.length;
  }
}

export class SubscriptionEventAppender<T> {
  constructor(
    private readonly events: SubscriptionEvent<T>[],
    private readonly nextSequence: () => number,
  ) {}
  append(value: T): SubscriptionEvent<T> {
    const event = { sequence: this.nextSequence(), value };
    this.events.push(event);
    return { ...event };
  }
}

export class SubscriptionEventReader<T> {
  constructor(private readonly events: readonly SubscriptionEvent<T>[]) {}
  since(sequence: number): SubscriptionEvent<T>[] {
    if (!Number.isInteger(sequence) || sequence < 0) throw new Error("invalid subscription cursor");
    return this.events.filter((event) => event.sequence > sequence).map((event) => ({ ...event }));
  }
}

export const subscriptionRequest = createManifestRequest("subscriptions", "channel");
