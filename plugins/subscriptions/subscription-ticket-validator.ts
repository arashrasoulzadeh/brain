import type { SubscriptionTicket } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class SubscriptionTicketValidator {
  constructor(private readonly tickets: ReadonlyMap<string, SubscriptionTicket>) {}
  validate(
    ticket: string,
    projectId: string,
    channel: string,
    operation = "subscriptions.open",
    now = Date.now(),
  ): SubscriptionTicket {
    validateProjectId(projectId, "invalid subscription ticket");
    if (!ticket || !channel || !operation || !Number.isFinite(now))
      throw new Error("invalid subscription ticket");
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
