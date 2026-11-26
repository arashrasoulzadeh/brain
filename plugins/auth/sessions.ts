import { expiresAt, isActiveToken, randomToken, tokenDigest, validateProjectId } from "../sdk-typescript/src/index.js";
import { projectUserID } from "./accounts.js";
import type { AuthState } from "./state.js";

export function sessionTokenDigest(token: string): string {
  return tokenDigest(token);
}

export function newSession(
  state: AuthState,
  projectId: string,
  username: string,
  now: number,
): { token: string; username: string } {
  const token = randomToken(32, "hex");
  state.sessions.set(sessionTokenDigest(token), {
    projectId,
    username,
    expiresAt: expiresAt(now, state.sessionTtlMs, "invalid session"),
  });
  return { token, username };
}

export class AccountSessionReader {
  constructor(private readonly state: AuthState) {}
  session(projectId: string, token: string, now = Date.now()) {
    validateProjectId(projectId);
    const session = this.state.sessions.get(sessionTokenDigest(token));
    if (!isActiveToken(session, now) || session.projectId !== projectId)
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
    if (!isActiveToken(session, now) || session.projectId !== projectId)
      throw new Error("invalid session");
    this.state.sessions.delete(digest);
    this.state.rotated.add(digest);
    return newSession(this.state, projectId, session.username, now).token;
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
    const username = `guest-${randomToken(8, "hex")}`;
    const session = newSession(this.state, projectId, username, Date.now());
    return { ...session, userId: projectUserID(projectId, username), guest: true };
  }
}

export class OperatorSessionIssuer {
  constructor(private readonly sessions: Map<string, string>) {}
  issue(operatorId: string): string {
    if (!operatorId.trim()) throw new Error("operator id required");
    const token = randomToken(32, "hex");
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
