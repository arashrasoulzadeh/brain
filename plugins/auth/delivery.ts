export class DeliveryWriter {
  constructor(private readonly messages: Map<string, string[]>) {}
  deliver(destination: string, body: string): void {
    if (!destination.trim() || !body.trim()) throw new Error("invalid test delivery");
    this.messages.set(destination, [...(this.messages.get(destination) ?? []), body]);
  }
}

export class DeliveryReader {
  constructor(private readonly messages: ReadonlyMap<string, readonly string[]>) {}
  read(destination: string): string | undefined {
    return this.messages.get(destination)?.at(-1);
  }
}

export function validateDeliveryConfig(environment: string, provider: string): void {
  if (environment === "production" && ["local", "fake", "test"].includes(provider))
    throw new Error("fake delivery forbidden in production");
  if (!environment || !provider) throw new Error("invalid delivery configuration");
}
