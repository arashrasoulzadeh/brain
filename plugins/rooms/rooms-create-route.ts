import { roomRouteStore } from "./room-route-store.js";
export function create(input: unknown): unknown {
  const id =
    typeof (input as { id?: unknown } | null)?.id === "string"
      ? (input as { id: string }).id.trim()
      : "";
  if (!id) throw new Error("room id is required");
  if (
    (input as { room?: unknown } | null)?.room &&
    typeof (input as { room: unknown }).room === "object"
  )
    return (input as { room: unknown }).room;
  return roomRouteStore.execute("create", id);
}
