import type { RoomRuntime } from "./room-runtime.js";

export function stopRoomDrain(runtime: RoomRuntime): void {
  runtime.drainUntil = undefined;
}
