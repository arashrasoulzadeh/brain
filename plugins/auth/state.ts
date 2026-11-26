export type Account = { username: string; userId: string; passwordHash: string };
export type AuthSession = { projectId: string; username: string; expiresAt: number };
export type AuthState = {
  accounts: Map<string, Map<string, Account>>;
  sessions: Map<string, AuthSession>;
  rotated: Set<string>;
  sessionTtlMs: number;
};

export function createAuthState(sessionTtlMs = 24 * 60 * 60 * 1000): AuthState {
  if (!Number.isFinite(sessionTtlMs) || sessionTtlMs <= 0) throw new Error("invalid session ttl");
  return { accounts: new Map(), sessions: new Map(), rotated: new Set(), sessionTtlMs };
}
