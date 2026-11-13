import { validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import type { AuthState } from "./auth-state.js";
export class AccountSessionRevoker { constructor(private readonly state: AuthState) {} logout(projectId: string, token: string): void { validateProjectId(projectId); const digest = sessionTokenDigest(token); const session = this.state.sessions.get(digest); if (!session || session.projectId !== projectId) throw new Error("invalid session"); this.state.sessions.delete(digest); } }
