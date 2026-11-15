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
