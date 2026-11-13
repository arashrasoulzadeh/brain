export class DeliveryWriter {
  constructor(private readonly messages: Map<string, string[]>) {}
  deliver(destination: string, body: string): void { if (!destination.trim() || !body.trim()) throw new Error("invalid test delivery"); this.messages.set(destination, [...(this.messages.get(destination) ?? []), body]); }
}
