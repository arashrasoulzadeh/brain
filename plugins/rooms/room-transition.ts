import type { RoomState } from "./index.js";

export function canTransitionRoom(from: RoomState, to: RoomState): boolean {
  return ({ waiting: ["matching", "closed"], matching: ["active", "waiting", "closed"], active: ["finished", "closed"], finished: ["closed"], closed: [] } as Record<RoomState, RoomState[]>)[from].includes(to);
}
