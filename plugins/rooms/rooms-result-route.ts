import { roomRouteStore } from "./room-route-store.js";
export function result(input: unknown): unknown {
  const value = (input ?? {}) as { id?: unknown; result?: unknown; room?: unknown };
  const id = typeof value.id === "string" ? value.id.trim() : "";
  if (!id) throw new Error("room id is required");
  if (value.room && typeof value.room === "object") return value.room;
  return roomRouteStore.execute("result", id, undefined, value.result);
}
