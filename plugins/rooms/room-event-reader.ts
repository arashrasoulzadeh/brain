import { projectRoomID } from "./project-room-id.js";
import type { RoomRuntime } from "./room-runtime.js";
export class RoomEventReader { constructor(private readonly runtime: RoomRuntime) {} events(projectId: string, roomId: string) { return [...(this.runtime.events.get(projectRoomID(projectId, roomId)) ?? [])]; } }
