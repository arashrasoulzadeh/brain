import { createResourceRequest, createOperationRoutes, requiredString, trimmedString, type PluginContext } from "../sdk-typescript/src/index.js";

export const roomRequest = createResourceRequest<"rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result">("rooms", "roomId");

type Room = { id: string; members: string[]; result?: unknown };
export class RoomRouteStore {
  private readonly rooms = new Map<string, Room>();
  execute(
    operation: "create" | "join" | "reconnect" | "result",
    roomID: string,
    member?: string,
    value?: unknown,
  ): unknown {
    const room = this.rooms.get(roomID);
    if (operation === "create") {
      const created = { id: roomID, members: [] };
      this.rooms.set(roomID, created);
      return created;
    }
    if (!room) throw new Error("room not found");
    if (operation === "join" || operation === "reconnect") {
      if (member && !room.members.includes(member)) room.members.push(member);
      return { id: roomID, member, reconnected: operation === "reconnect", members: room.members };
    }
    if (value !== undefined) room.result = value;
    return { id: roomID, result: room.result, members: room.members };
  }
}
export const roomRouteStore = new RoomRouteStore();

type RoomOperation = "create" | "join" | "reconnect" | "result";
type RoomInput = { id?: unknown; member?: unknown; room?: unknown; result?: unknown };

const routes = createOperationRoutes(["create", "join", "reconnect", "result"], executeRoomRoute);
export const { create, join, reconnect, result } = routes.handlers;
export const dispatchRoute = routes.dispatch;

function executeRoomRoute(operation: RoomOperation, input: unknown, context?: PluginContext): unknown {
  const value = (input ?? {}) as RoomInput;
  if (operation === "join" || operation === "reconnect") {
    const id = trimmedString(value.id);
    const member = trimmedString(value.member);
    if (!id || !member) throw new Error("room id and member are required");
    const result = value.room && typeof value.room === "object"
      ? { ...(value.room as object), member, reconnected: operation === "reconnect" }
      : roomRouteStore.execute(operation, id, member);
    context?.plugins.observe.fire("connected", { roomId: id, reconnected: operation === "reconnect" });
    if (operation === "reconnect") context?.plugins.observe.fire("reconnected", { roomId: id });
    return result;
  }
  const id = requiredString(value.id, "room id is required");
  const result = value.room && typeof value.room === "object" ? value.room : operation === "create"
    ? roomRouteStore.execute(operation, id)
    : roomRouteStore.execute(operation, id, undefined, value.result);
  if (operation === "create") context?.plugins.observe.fire("created", { roomId: id });
  if (operation === "result" && value.result !== undefined) context?.plugins.observe.fire("result.committed", { roomId: id });
  return result;
}
