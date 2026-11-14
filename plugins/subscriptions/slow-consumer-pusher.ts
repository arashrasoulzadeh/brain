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
