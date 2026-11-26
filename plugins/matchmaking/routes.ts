import { createOperationRoutes, requiredString } from "../sdk-typescript/src/index.js";
import { matchmakingRouteStore } from "./matchmaking-route-store.js";

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
