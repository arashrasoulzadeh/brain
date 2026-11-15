import { roomRouteStore } from "./room-route-store.js";
export function result(input: unknown): unknown {
  const value = (input ?? {}) as { id?: unknown; result?: unknown };
  const id = typeof value.id === "string" ? value.id.trim() : "";
  if (!id) throw new Error("room id is required");
  return roomRouteStore.execute("result", id, undefined, value.result);
}
