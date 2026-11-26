import { createRouteDispatcher, requiredString } from "../sdk-typescript/src/index.js";
import { matchmakingRouteStore } from "./matchmaking-route-store.js";

type TicketOperation = "queue" | "cancel" | "status" | "assign";
type TicketInput = { id?: unknown; ticket?: unknown };

const dispatch = createRouteDispatcher({ queue, cancel, status, assign });

export function queue(input: unknown): unknown {
  return executeTicketRoute("queue", input);
}

export function cancel(input: unknown): unknown {
  return executeTicketRoute("cancel", input);
}

export function status(input: unknown): unknown {
  return executeTicketRoute("status", input);
}

export function assign(input: unknown): unknown {
  return executeTicketRoute("assign", input);
}

export function dispatchRoute(name: string, input: unknown): unknown {
  return dispatch(name, input);
}

function executeTicketRoute(operation: TicketOperation, input: unknown): unknown {
  const value = (input ?? {}) as TicketInput;
  const id = requiredString(value.id, "ticket id is required");
  if (operation !== "queue" && value.ticket && typeof value.ticket === "object")
    return value.ticket;
  return matchmakingRouteStore.execute(operation, id);
}
