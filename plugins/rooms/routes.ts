import { createOperationRoutes, requiredString, trimmedString } from "../sdk-typescript/src/index.js";
import { roomRouteStore } from "./room-route-store.js";

type RoomOperation = "create" | "join" | "reconnect" | "result";
type RoomInput = { id?: unknown; member?: unknown; room?: unknown; result?: unknown };

const routes = createOperationRoutes(["create", "join", "reconnect", "result"], executeRoomRoute);
export const { create, join, reconnect, result } = routes.handlers;
export const dispatchRoute = routes.dispatch;

function executeRoomRoute(operation: RoomOperation, input: unknown): unknown {
  const value = (input ?? {}) as RoomInput;
  if (operation === "join" || operation === "reconnect") {
    const id = trimmedString(value.id);
    const member = trimmedString(value.member);
    if (!id || !member) throw new Error("room id and member are required");
    if (value.room && typeof value.room === "object")
      return { ...(value.room as object), member, reconnected: operation === "reconnect" };
    return roomRouteStore.execute(operation, id, member);
  }
  const id = requiredString(value.id, "room id is required");
  if (value.room && typeof value.room === "object") return value.room;
  return operation === "create"
    ? roomRouteStore.execute(operation, id)
    : roomRouteStore.execute(operation, id, undefined, value.result);
}
