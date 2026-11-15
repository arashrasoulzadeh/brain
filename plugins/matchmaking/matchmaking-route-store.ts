type Ticket = { id: string; status: "queued" | "cancelled" | "assigned" };
export class MatchmakingRouteStore {
  private readonly tickets = new Map<string, Ticket>();
  execute(operation: "queue" | "cancel" | "status" | "assign", id: string): Ticket {
    const existing = this.tickets.get(id);
    if (operation === "queue") {
      const ticket = { id, status: "queued" as const };
      this.tickets.set(id, ticket);
      return ticket;
    }
    if (!existing) throw new Error("ticket not found");
    if (operation === "cancel") existing.status = "cancelled";
    if (operation === "assign") existing.status = "assigned";
    return existing;
  }
}
export const matchmakingRouteStore = new MatchmakingRouteStore();
