import type { ArenaState } from "./types.js";

export function interruptExpiredActionRoom(
  state: ArenaState,
  now: number,
  expiresAt: number,
): ArenaState {
  if (state !== "active" || !Number.isFinite(now) || !Number.isFinite(expiresAt) || now < expiresAt)
    throw new Error("action room has not expired");
  return "interrupted";
}
