import type { RoomRuntime } from "./room-runtime.js";

export function isRoomDraining(runtime: RoomRuntime): boolean {
  return runtime.drainUntil !== undefined && runtime.drainUntil > Date.now();
}
