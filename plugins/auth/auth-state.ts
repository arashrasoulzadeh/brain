export type Account = { username: string; userId: string; passwordHash: string };
export type AuthSession = { projectId: string; username: string; expiresAt: number };
export type AuthState = {
  accounts: Map<string, Map<string, Account>>;
  sessions: Map<string, AuthSession>;
  rotated: Set<string>;
  sessionTtlMs: number;
};
export { createAuthState } from "./create-auth-state.js";
export { validateAccount } from "./validate-account.js";
export { passwordHash } from "./password-hash.js";
export { newSession } from "./new-session.js";
export { accountProject } from "./account-project.js";
export { projectUserID } from "./project-user-id.js";
