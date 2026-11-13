export class DeliveryReader {
  constructor(private readonly messages: ReadonlyMap<string, readonly string[]>) {}
  read(destination: string): string | undefined { return this.messages.get(destination)?.at(-1); }
}
