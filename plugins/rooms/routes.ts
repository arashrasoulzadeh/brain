import { pluginRequest, createOperationRoutes, requiredString, trimmedString } from "../sdk-typescript/src/index.js";

export function roomRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result",
  roomId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "rooms",
    projectId,
    pluginId,
    version,
    operation,
    { roomId, ...(payload as Record<string, unknown>) },
    roomId,
  );
}

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
