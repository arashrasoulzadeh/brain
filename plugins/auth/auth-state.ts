import { randomBytes } from "node:crypto";
import { hash as argon2Hash } from "@node-rs/argon2";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import { projectUserID, sessionTokenDigest } from "./index.js";

export type Account = { username: string; userId: string; passwordHash: string };
export type AuthSession = { projectId: string; username: string; expiresAt: number };
export type AuthState = { accounts: Map<string, Map<string, Account>>; sessions: Map<string, AuthSession>; rotated: Set<string>; sessionTtlMs: number };

export function createAuthState(sessionTtlMs = 24 * 60 * 60 * 1000): AuthState {
  if (sessionTtlMs <= 0) throw new Error("invalid session ttl");
  return { accounts: new Map(), sessions: new Map(), rotated: new Set(), sessionTtlMs };
}

export function validateAccount(projectId: string, username: string, password: string): void {
  validateProjectId(projectId);
  if (!username || password.length < 8) throw new Error("invalid signup");
}

export function passwordHash(password: string): Promise<string> { return argon2Hash(password, { memoryCost: 19456, timeCost: 2, parallelism: 1, outputLen: 32 }); }
export function newSession(state: AuthState, projectId: string, username: string, now: number): { token: string; username: string } {
  const token = randomBytes(32).toString("hex");
  state.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: now + state.sessionTtlMs });
  return { token, username };
}
export function accountProject(state: AuthState, projectId: string): Map<string, Account> {
  const project = state.accounts.get(projectId) ?? new Map<string, Account>();
  state.accounts.set(projectId, project);
  return project;
}
export { projectUserID };
