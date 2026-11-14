import { validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import type { AuthState } from "./auth-state.js";
export class AccountSessionReader {
  constructor(private readonly state: AuthState) {}
  session(projectId: string, token: string, now = Date.now()) {
    validateProjectId(projectId);
    const session = this.state.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || session.expiresAt <= now)
      throw new Error("invalid session");
    return { username: session.username };
  }
}
