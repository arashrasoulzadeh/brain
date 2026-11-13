import { randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import type { AuthState } from "./auth-state.js";
export class AccountSessionRotator { constructor(private readonly state: AuthState) {} rotate(projectId: string, token: string, now = Date.now()) { validateProjectId(projectId); const digest = sessionTokenDigest(token); const session = this.state.sessions.get(digest); if (!session && this.state.rotated.has(digest)) throw new Error("refresh token reuse detected"); if (!session || session.projectId !== projectId || session.expiresAt <= now) throw new Error("invalid session"); this.state.sessions.delete(digest); this.state.rotated.add(digest); const rotated = randomBytes(32).toString("hex"); this.state.sessions.set(sessionTokenDigest(rotated), { projectId, username: session.username, expiresAt: now + this.state.sessionTtlMs }); return rotated; } }
