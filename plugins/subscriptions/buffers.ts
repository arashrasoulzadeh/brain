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
