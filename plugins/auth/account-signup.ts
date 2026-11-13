import { projectUserID } from "./auth-state.js";
import { accountProject, passwordHash, type AuthState, validateAccount } from "./auth-state.js";
export class AccountSignup {
  constructor(private readonly state: AuthState) {}
  async signup(projectId: string, username: string, password: string) { validateAccount(projectId, username, password); const project = accountProject(this.state, projectId); if (project.has(username)) throw new Error("account already exists"); const userId = projectUserID(projectId, username); project.set(username, { username, userId, passwordHash: await passwordHash(password) }); return { username, userId }; }
}
