import {
  projectUserID,
  accountProject,
  passwordHash,
  type AuthState,
  validateAccount,
} from "./auth-state.js";
export class AccountPasswordResetter {
  constructor(private readonly state: AuthState) {}
  async resetPassword(projectId: string, username: string, password: string): Promise<void> {
    validateAccount(projectId, username, password);
    const account = this.state.accounts.get(projectId)?.get(username);
    if (!account) throw new Error("account not found");
    account.passwordHash = await passwordHash(password);
    for (const [digest, session] of this.state.sessions)
      if (session.projectId === projectId && session.username === username)
        this.state.sessions.delete(digest);
  }
}
