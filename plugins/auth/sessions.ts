import { createHash, randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import { projectUserID } from "./accounts.js";
import type { AuthState } from "./state.js";

export function sessionTokenDigest(token: string): string {
  return createHash("sha256").update(token).digest("hex");
}

export function newSession(
  state: AuthState,
  projectId: string,
  username: string,
  now: number,
): { token: string; username: string } {
  const token = randomBytes(32).toString("hex");
  state.sessions.set(sessionTokenDigest(token), {
    projectId,
    username,
    expiresAt: now + state.sessionTtlMs,
  });
  return { token, username };
}

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

export class AccountSessionRotator {
  constructor(private readonly state: AuthState) {}
  rotate(projectId: string, token: string, now = Date.now()) {
    validateProjectId(projectId);
    const digest = sessionTokenDigest(token);
    const session = this.state.sessions.get(digest);
    if (!session && this.state.rotated.has(digest)) throw new Error("refresh token reuse detected");
    if (!session || session.projectId !== projectId || session.expiresAt <= now)
      throw new Error("invalid session");
    this.state.sessions.delete(digest);
    this.state.rotated.add(digest);
    const rotated = randomBytes(32).toString("hex");
    this.state.sessions.set(sessionTokenDigest(rotated), {
      projectId,
      username: session.username,
      expiresAt: now + this.state.sessionTtlMs,
    });
    return rotated;
  }
}

export class AccountSessionRevoker {
  constructor(private readonly state: AuthState) {}
  logout(projectId: string, token: string): void {
    validateProjectId(projectId);
    const digest = sessionTokenDigest(token);
    const session = this.state.sessions.get(digest);
    if (!session || session.projectId !== projectId) throw new Error("invalid session");
    this.state.sessions.delete(digest);
  }
}

export class GuestSessionIssuer {
  constructor(private readonly state: AuthState) {}
  guest(projectId: string) {
    validateProjectId(projectId);
    const username = `guest-${randomBytes(8).toString("hex")}`;
    const session = newSession(this.state, projectId, username, Date.now());
    return { ...session, userId: projectUserID(projectId, username), guest: true };
  }
}

export class OperatorSessionIssuer {
  constructor(private readonly sessions: Map<string, string>) {}
  issue(operatorId: string): string {
    if (!operatorId.trim()) throw new Error("operator id required");
    const token = randomBytes(32).toString("hex");
    this.sessions.set(sessionTokenDigest(token), operatorId);
    return token;
  }
}

export class OperatorSessionValidator {
  constructor(
    private readonly sessions: ReadonlyMap<string, string>,
    private readonly revoked: ReadonlySet<string>,
  ) {}
  valid(token: string, operatorId: string): boolean {
    const digest = sessionTokenDigest(token);
    return !this.revoked.has(digest) && this.sessions.get(digest) === operatorId;
  }
}

export class OperatorSessionRevoker {
  constructor(
    private readonly sessions: Map<string, string>,
    private readonly revoked: Set<string>,
  ) {}
  revoke(token: string): void {
    const digest = sessionTokenDigest(token);
    this.sessions.delete(digest);
    this.revoked.add(digest);
  }
}

export function firstPartySessionCookie(token: string, maxAgeSeconds: number): string {
  if (!token || !Number.isSafeInteger(maxAgeSeconds) || maxAgeSeconds <= 0)
    throw new Error("invalid session cookie");
  if (/[;\s]/.test(token)) throw new Error("invalid session cookie");
  return `__Host-brain_session=${token}; Max-Age=${maxAgeSeconds}; Path=/; HttpOnly; Secure; SameSite=Lax`;
}

export function requireRecentAuthentication(
  authenticatedAt: number,
  now: number,
  maxAgeMs: number,
): void {
  if (maxAgeMs <= 0 || now - authenticatedAt > maxAgeMs)
    throw new Error("recent authentication required");
}
