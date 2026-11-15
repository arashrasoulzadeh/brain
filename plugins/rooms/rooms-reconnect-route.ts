import { roomRouteStore } from "./room-route-store.js";
export function reconnect(input: unknown): unknown {
  const value = (input ?? {}) as { id?: unknown; member?: unknown };
  const id = typeof value.id === "string" ? value.id.trim() : "";
  const member = typeof value.member === "string" ? value.member.trim() : "";
  if (!id || !member) throw new Error("room id and member are required");
  return roomRouteStore.execute("reconnect", id, member);
}
