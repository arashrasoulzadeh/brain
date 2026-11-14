import type { SubscriptionTicket } from "./types.js";

export class SubscriptionTicketValidator {
  constructor(private readonly tickets: ReadonlyMap<string, SubscriptionTicket>) {}
  validate(
    ticket: string,
    projectId: string,
    channel: string,
    operation = "subscriptions.open",
    now = Date.now(),
  ): SubscriptionTicket {
    const found = this.tickets.get(ticket);
    if (
      !found ||
      found.projectId !== projectId ||
      found.channel !== channel ||
      found.operation !== operation ||
      now >= found.expiresAt
    )
      throw new Error("invalid subscription ticket");
    return { ...found };
  }
}
