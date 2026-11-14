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
