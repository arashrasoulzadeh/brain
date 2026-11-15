import type { Room } from "./index.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export function authorizedRoomSnapshot(room: Room, projectId: string, playerId: string): Room {
  validateProjectId(projectId, "room membership required");
  if (room.projectId !== projectId || !room.slots.some((slot) => slot.playerId === playerId))
    throw new Error("room membership required");
  return { ...room, slots: [...room.slots] };
}
