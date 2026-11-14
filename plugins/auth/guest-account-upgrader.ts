import { validateProjectId } from "../sdk-typescript/src/index.js";
import {
  accountProject,
  passwordHash,
  projectUserID,
  type AuthState,
  validateAccount,
} from "./auth-state.js";
import { sessionTokenDigest } from "./session-token-digest.js";
export class GuestAccountUpgrader {
  constructor(private readonly state: AuthState) {}
  async upgradeGuest(projectId: string, token: string, username: string, password: string) {
    validateProjectId(projectId);
    const session = this.state.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || !session.username.startsWith("guest-"))
      throw new Error("guest session required");
    const userId = projectUserID(projectId, session.username);
    validateAccount(projectId, username, password);
    const project = accountProject(this.state, projectId);
    if (project.has(username)) throw new Error("account already exists");
    project.set(username, { username, userId, passwordHash: await passwordHash(password) });
    session.username = username;
    return { username, userId };
  }
}
