import { matchmakingRouteStore } from "./matchmaking-route-store.js";
export function cancel(input: unknown): unknown {
  const value = (input ?? {}) as { id?: unknown; ticket?: unknown };
  const id = typeof value.id === "string" ? value.id.trim() : "";
  if (!id) throw new Error("ticket id is required");
  if (value.ticket && typeof value.ticket === "object") return value.ticket;
  return matchmakingRouteStore.execute("cancel", id);
}
