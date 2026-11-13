export class InitialFetchBuffer<T> {
  private fetching = true;
  private readonly pending: T[] = [];
  push(event: T): void { if (this.fetching) this.pending.push(event); }
  complete(snapshot: readonly T[]): T[] { this.fetching = false; return [...snapshot, ...this.pending.splice(0)]; }
}
