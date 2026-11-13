export class SlowConsumerSize {
  constructor(private readonly items: readonly unknown[]) {}
  size(): number { return this.items.length; }
}
