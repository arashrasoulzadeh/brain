import { projectRoomID } from "./project-room-id.js";
import type { RoomRuntime } from "./room-runtime.js";
export class RoomReader { constructor(private readonly runtime: RoomRuntime) {} get(projectId: string, roomId: string) { const room = this.runtime.rooms.get(projectRoomID(projectId, roomId)); if (!room) throw new Error("room not found"); return { ...room, slots: [...room.slots] }; } }
