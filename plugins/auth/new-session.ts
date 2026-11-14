import { randomBytes } from "node:crypto";
import { sessionTokenDigest } from "./session-token-digest.js";
import type { AuthState } from "./auth-state.js";
export function newSession(
  state: AuthState,
  projectId: string,
  username: string,
  now: number,
): { token: string; username: string } {
  const token = randomBytes(32).toString("hex");
  state.sessions.set(sessionTokenDigest(token), {
    projectId,
    username,
    expiresAt: now + state.sessionTtlMs,
  });
  return { token, username };
}
