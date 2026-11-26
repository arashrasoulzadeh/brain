import { createResourceRequest, createOperationRoutes, requiredString } from "../sdk-typescript/src/index.js";

export const matchmakingRequest = createResourceRequest<"matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign">("matchmaking", "config");

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

type TicketOperation = "queue" | "cancel" | "status" | "assign";
type TicketInput = { id?: unknown; ticket?: unknown };

const routes = createOperationRoutes(["queue", "cancel", "status", "assign"], executeTicketRoute);
export const { queue, cancel, status, assign } = routes.handlers;
export const dispatchRoute = routes.dispatch;

function executeTicketRoute(operation: TicketOperation, input: unknown): unknown {
  const value = (input ?? {}) as TicketInput;
  const id = requiredString(value.id, "ticket id is required");
  if (operation !== "queue" && value.ticket && typeof value.ticket === "object")
    return value.ticket;
  return matchmakingRouteStore.execute(operation, id);
}
