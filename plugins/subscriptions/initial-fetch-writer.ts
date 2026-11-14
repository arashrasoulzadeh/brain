export class InitialFetchWriter<T> {
  constructor(
    private readonly pending: T[],
    private readonly isFetching: () => boolean,
  ) {}
  push(event: T): void {
    if (this.isFetching()) this.pending.push(event);
  }
}
