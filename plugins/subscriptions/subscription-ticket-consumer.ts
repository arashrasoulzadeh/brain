import type { SubscriptionTicket } from "./types.js";
import { SubscriptionTicketValidator } from "./subscription-ticket-validator.js";

export class SubscriptionTicketConsumer {
  private readonly validator: SubscriptionTicketValidator;
  constructor(private readonly tickets: Map<string, SubscriptionTicket>) { this.validator = new SubscriptionTicketValidator(tickets); }
  consume(ticket: string, projectId: string, channel: string, operation = "subscriptions.open", now = Date.now()): SubscriptionTicket {
    const found = this.validator.validate(ticket, projectId, channel, operation, now);
    this.tickets.delete(ticket);
    return found;
  }
}
