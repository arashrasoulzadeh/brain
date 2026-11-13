import { projectRoomID } from "./project-room-id.js";
import type { RoomRuntime } from "./room-runtime.js";
import type { Room } from "./index.js";
export class RoomCreator { constructor(private readonly runtime: RoomRuntime) {} create(projectId: string, roomId: string, maxPlayers: number): Room { const id = projectRoomID(projectId, roomId); if (maxPlayers < 1 || maxPlayers > 64) throw new Error("invalid max players"); if (this.runtime.rooms.has(id)) throw new Error("room already exists"); const room: Room = { id, projectId, state: "waiting", maxPlayers, slots: [] }; this.runtime.rooms.set(id, room); this.runtime.events.set(id, []); this.runtime.store.save(id, []); return { ...room, slots: [] }; } }
