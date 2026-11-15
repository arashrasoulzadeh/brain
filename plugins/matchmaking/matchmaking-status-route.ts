import { matchmakingRouteStore } from "./matchmaking-route-store.js";
export function status(input: unknown): unknown {
  const id =
    typeof (input as { id?: unknown } | null)?.id === "string"
      ? (input as { id: string }).id.trim()
      : "";
  if (!id) throw new Error("ticket id is required");
  return matchmakingRouteStore.execute("status", id);
}
