import { projectRoomID } from "./project-room-id.js";
import type { PlayerSlot } from "./index.js";

export function validateMembership(projectId: string, roomId: string, slots: readonly PlayerSlot[], maxPlayers: number): void {
  projectRoomID(projectId, roomId);
  if (maxPlayers < 1 || maxPlayers > 64 || slots.length > maxPlayers) throw new Error("invalid player slots");
  const seen = new Set<number>();
  for (const player of slots) {
    if (!player.playerId || player.slot < 0 || player.slot >= maxPlayers || seen.has(player.slot)) throw new Error("invalid player slots");
    seen.add(player.slot);
  }
}
