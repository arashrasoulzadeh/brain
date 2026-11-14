import type { RoomRuntime } from "./room-runtime.js";

export function startRoomDrain(runtime: RoomRuntime, deadline: number): void {
  if (!Number.isFinite(deadline) || deadline <= Date.now())
    throw new Error("invalid drain deadline");
  runtime.drainUntil = deadline;
}
