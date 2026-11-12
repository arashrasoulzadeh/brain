import { createHash, randomBytes, scryptSync, timingSafeEqual } from "node:crypto";
import { request, type PluginFunction, type RpcRequest } from "../sdk-typescript/src/index.js";

export function signInRequest(projectId: string, pluginId: string, version: string, provider: string, credential: unknown) {
  return request(`auth-signin-${projectId}`, { protocol: 1, projectId, pluginId, version, operation: "auth.signin", payload: { provider, credential } });
}

export function signUpRequest(projectId: string, pluginId: string, version: string, username: string, password: string) {
  return request(`auth-signup-${projectId}-${username}`, { protocol: 1, projectId, pluginId, version, operation: "auth.signup", payload: { username, password } });
}

export function guestRequest(projectId: string, pluginId: string, version: string) {
  return request(`auth-guest-${projectId}`, { protocol: 1, projectId, pluginId, version, operation: "auth.guest", payload: {} });
}

type Account = { username: string; salt: Buffer; passwordHash: Buffer };

export function projectUserID(projectId: string, userId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(userId)) throw new Error("invalid project user id");
  return `${projectId}:${userId}`;
}

export function sessionTokenDigest(token: string): string { return createHash("sha256").update(token).digest("hex"); }

export class AccountRateLimiter {
  private readonly attempts = new Map<string, { startedAt: number; count: number }>();
  constructor(private readonly limit: number, private readonly windowMs: number) {
    if (!Number.isInteger(limit) || limit < 1 || windowMs <= 0) throw new Error("invalid account rate limit");
  }
  allow(projectId: string, username: string, now: number): boolean {
    const key = projectUserID(projectId, username);
    const current = this.attempts.get(key);
    if (!current || now - current.startedAt >= this.windowMs) { this.attempts.set(key, { startedAt: now, count: 1 }); return true; }
    if (current.count >= this.limit) return false;
    current.count += 1;
    return true;
  }
}

export class ProjectRateLimiter {
  private readonly attempts = new Map<string, { startedAt: number; count: number }>();
  constructor(private readonly limit: number, private readonly windowMs: number) {
    if (!Number.isInteger(limit) || limit < 1 || windowMs <= 0) throw new Error("invalid project rate limit");
  }
  allow(projectId: string, now: number): boolean {
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error("invalid project id");
    const current = this.attempts.get(projectId);
    if (!current || now - current.startedAt >= this.windowMs) { this.attempts.set(projectId, { startedAt: now, count: 1 }); return true; }
    if (current.count >= this.limit) return false;
    current.count += 1;
    return true;
  }
}

export class AuthService {
  private readonly accounts = new Map<string, Map<string, Account>>();
  private readonly sessions = new Map<string, { projectId: string; username: string; expiresAt: number }>();
  private readonly rotated = new Set<string>();
  constructor(private readonly sessionTtlMs = 24 * 60 * 60 * 1000) {
    if (sessionTtlMs <= 0) throw new Error("invalid session ttl");
  }

  signup(projectId: string, username: string, password: string) {
    this.validate(projectId, username, password);
    const project = this.accounts.get(projectId) ?? new Map<string, Account>();
    if (project.has(username)) throw new Error("account already exists");
    const salt = randomBytes(16);
    project.set(username, { username, salt, passwordHash: this.hash(password, salt) });
    this.accounts.set(projectId, project);
    return { username, userId: projectUserID(projectId, username) };
  }

  signin(projectId: string, username: string, password: string, now = Date.now()) {
    const account = this.accounts.get(projectId)?.get(username);
    if (!account || !timingSafeEqual(account.passwordHash, this.hash(password, account.salt))) throw new Error("invalid credentials");
    const token = randomBytes(32).toString("hex");
    this.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: now + this.sessionTtlMs });
    return { token, username };
  }

  session(projectId: string, token: string, now = Date.now()) {
    const session = this.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || session.expiresAt <= now) throw new Error("invalid session");
    return { username: session.username };
  }

  rotate(projectId: string, token: string, now = Date.now()): string {
    const digest = sessionTokenDigest(token);
    const session = this.sessions.get(digest);
    if (!session && this.rotated.has(digest)) throw new Error("refresh token reuse detected");
    if (!session || session.projectId !== projectId || session.expiresAt <= now) throw new Error("invalid session");
    this.sessions.delete(digest);
    this.rotated.add(digest);
    const rotated = randomBytes(32).toString("hex");
    this.sessions.set(sessionTokenDigest(rotated), { projectId, username: session.username, expiresAt: now + this.sessionTtlMs });
    return rotated;
  }

  guest(projectId: string) {
    if (!projectId) throw new Error("project id required");
    const token = randomBytes(32).toString("hex");
    const username = `guest-${randomBytes(8).toString("hex")}`;
    this.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: Date.now() + this.sessionTtlMs });
    return { token, username, guest: true };
  }

  private hash(password: string, salt: Buffer) { return scryptSync(password, salt, 32); }
  private validate(projectId: string, username: string, password: string) {
    if (!projectId || !username || password.length < 8) throw new Error("invalid signup");
  }
}

export function functions(service = new AuthService()): Readonly<Record<string, PluginFunction>> {
  const payload = (rpc: RpcRequest) => rpc.params.payload as Record<string, string> | undefined;
  return {
    signup: (rpc) => { const p = payload(rpc); return service.signup(rpc.params.projectId, p?.username ?? "", p?.password ?? ""); },
    signin: (rpc) => { const p = payload(rpc); return service.signin(rpc.params.projectId, p?.username ?? p?.provider ?? "", p?.password ?? String(p?.credential ?? "")); },
    session: (rpc) => { const p = payload(rpc); return service.session(rpc.params.projectId, p?.token ?? ""); },
    guest: (rpc) => service.guest(rpc.params.projectId),
  };
}
