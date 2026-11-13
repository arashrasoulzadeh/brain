import { projectRoomID } from "./project-room-id.js";
import type { RoomRuntime } from "./room-runtime.js";
export class RoomMembershipReader { constructor(private readonly runtime: RoomRuntime) {} isMember(projectId: string, roomId: string, playerId: string): boolean { return this.runtime.rooms.get(projectRoomID(projectId, roomId))?.slots.some((player) => player.playerId === playerId) ?? false; } }
