export class SlowConsumerShifter<T> {
  constructor(private readonly items: T[]) {}
  shift(): T | undefined {
    return this.items.shift();
  }
}
