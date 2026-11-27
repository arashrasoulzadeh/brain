import { hash as argon2Hash, verify as argon2Verify } from "@node-rs/argon2";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { Account, AuthState } from "./state.js";
import { newSession, sessionTokenDigest } from "./sessions.js";

export function projectUserID(projectId: string, userId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(userId))
    throw new Error("invalid project user id");
  return `${projectId}:${userId}`;
}

export function validateAccount(projectId: string, username: string, password: string): void {
  validateProjectId(projectId);
  if (!username || username.startsWith("guest-") || password.length < 8) throw new Error("invalid signup");
}

export function passwordHash(password: string): Promise<string> {
  return argon2Hash(password, { memoryCost: 19456, timeCost: 2, parallelism: 1, outputLen: 32 });
}

export function accountProject(state: AuthState, projectId: string): Map<string, Account> {
  validateProjectId(projectId);
  const project = state.accounts.get(projectId) ?? new Map<string, Account>();
  state.accounts.set(projectId, project);
  return project;
}

export class AccountSignup {
  constructor(private readonly state: AuthState) {}
  async signup(projectId: string, username: string, password: string) {
    validateAccount(projectId, username, password);
    const project = accountProject(this.state, projectId);
    if (project.has(username)) throw new Error("account already exists");
    const userId = projectUserID(projectId, username);
    project.set(username, { username, userId, passwordHash: await passwordHash(password) });
    return { username, userId };
  }
}

export class AccountSignin {
  constructor(private readonly state: AuthState) {}
  async signin(projectId: string, username: string, password: string, now = Date.now()) {
    validateProjectId(projectId);
    const account = this.state.accounts.get(projectId)?.get(username);
    if (!account || !(await argon2Verify(account.passwordHash, password)))
      throw new Error("invalid credentials");
    return newSession(this.state, projectId, username, now);
  }
}

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

export class GuestAccountUpgrader {
  constructor(private readonly state: AuthState) {}
  async upgradeGuest(projectId: string, token: string, username: string, password: string) {
    validateProjectId(projectId);
    const session = this.state.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || session.expiresAt <= Date.now() || !session.username.startsWith("guest-"))
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

export class IdentityLinker {
  private readonly links = new Map<string, string>();
  link(userId: string, provider: string, subject: string): void {
    if (!userId.trim() || !provider.trim() || !subject.trim())
      throw new Error("invalid identity link");
    const key = `${provider}:${subject}`;
    const existing = this.links.get(key);
    if (existing && existing !== userId) throw new Error("identity already linked");
    this.links.set(key, userId);
  }
}

export function publicAuthError(error: unknown): string {
  if (
    error instanceof Error &&
    ["account already exists", "invalid credentials", "account not found"].includes(error.message)
  )
    return "invalid credentials";
  return "authentication failed";
}
