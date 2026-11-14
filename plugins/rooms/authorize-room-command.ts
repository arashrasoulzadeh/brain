import { authorizedRoomSnapshot } from "./authorized-room-snapshot.js";
import type { Room } from "./index.js";
export function authorizeRoomCommand(
  room: Room,
  projectId: string,
  playerId: string,
  operation: string,
): boolean {
  if (!operation.trim()) throw new Error("operation required");
  authorizedRoomSnapshot(room, projectId, playerId);
  return true;
}
