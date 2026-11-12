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

type Account = { username: string; userId: string; salt: Buffer; passwordHash: Buffer };

export function projectUserID(projectId: string, userId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(userId)) throw new Error("invalid project user id");
  return `${projectId}:${userId}`;
}

export function sessionTokenDigest(token: string): string { return createHash("sha256").update(token).digest("hex"); }
export function createPKCEChallenge(verifier: string): string { return createHash("sha256").update(verifier).digest("base64url"); }
export function verifyPKCE(verifier: string, challenge: string, method: string): boolean { return method === "S256" && timingSafeEqual(Buffer.from(createPKCEChallenge(verifier)), Buffer.from(challenge)); }
export class OIDCStateStore {
  private readonly states = new Map<string, { projectId: string; clientId: string; redirectUri: string }>();
  issue(projectId: string, clientId: string, redirectUri: string): string {
    const state = randomBytes(24).toString("base64url");
    this.states.set(state, { projectId, clientId, redirectUri });
    return state;
  }
  consume(state: string, projectId: string, clientId: string, redirectUri: string): void {
    const record = this.states.get(state);
    if (!record || record.projectId !== projectId || record.clientId !== clientId || record.redirectUri !== redirectUri) throw new Error("callback state mismatch");
    this.states.delete(state);
  }
}
export class TestDeliveryInbox {
  private readonly messages = new Map<string, string[]>();
  deliver(destination: string, body: string): void {
    if (!destination.trim() || !body.trim()) throw new Error("invalid test delivery");
    this.messages.set(destination, [...(this.messages.get(destination) ?? []), body]);
  }
  read(destination: string): string | undefined { return this.messages.get(destination)?.at(-1); }
}

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

export class OperatorSessionRegistry {
  private readonly sessions = new Map<string, string>();
  private readonly revoked = new Set<string>();
  issue(operatorId: string): string {
    if (!operatorId.trim()) throw new Error("operator id required");
    const token = randomBytes(32).toString("hex");
    this.sessions.set(sessionTokenDigest(token), operatorId);
    return token;
  }
  valid(token: string, operatorId: string): boolean {
    const digest = sessionTokenDigest(token);
    return !this.revoked.has(digest) && this.sessions.get(digest) === operatorId;
  }
  revoke(token: string): void {
    const digest = sessionTokenDigest(token);
    this.sessions.delete(digest);
    this.revoked.add(digest);
  }
}

export class IdentityLinker {
  private readonly links = new Map<string, string>();
  link(userId: string, provider: string, subject: string): void {
    if (!userId.trim() || !provider.trim() || !subject.trim()) throw new Error("invalid identity link");
    const key = `${provider}:${subject}`;
    const existing = this.links.get(key);
    if (existing && existing !== userId) throw new Error("identity already linked");
    this.links.set(key, userId);
  }
}

export type OIDCConfig = { issuer: string; audience: string; clientId: string; redirectUri: string };
export class OIDCConfigStore {
  private readonly configs = new Map<string, OIDCConfig>();
  set(projectId: string, config: OIDCConfig): void {
    if (!projectId || !config.audience || !config.clientId || !/^https:\/\//.test(config.issuer) || !/^https:\/\//.test(config.redirectUri)) throw new Error("invalid OIDC configuration");
    this.configs.set(projectId, { ...config });
  }
  get(projectId: string): OIDCConfig {
    const config = this.configs.get(projectId);
    if (!config) throw new Error("OIDC configuration not found");
    return { ...config };
  }
}

export type OIDCTokenClaims = { issuer: string; audience: string; expiresAt: number; nonce: string; signature: string };
export function validateOIDCToken(claims: OIDCTokenClaims, config: OIDCConfig, expectedNonce: string, now: number, verifySignature: (signature: string) => boolean): void {
  if (claims.issuer !== config.issuer || claims.audience !== config.audience || claims.expiresAt <= now || claims.nonce !== expectedNonce || !verifySignature(claims.signature)) throw new Error("invalid OIDC token");
}
export function requireRecentAuthentication(authenticatedAt: number, now: number, maxAgeMs: number): void {
  if (maxAgeMs <= 0 || now - authenticatedAt > maxAgeMs) throw new Error("recent authentication required");
}

type AuthorizationCode = { projectId: string; clientId: string; redirectUri: string; userId: string; expiresAt: number; used: boolean };
export class AuthorizationCodeStore {
  private readonly codes = new Map<string, AuthorizationCode>();
  issue(projectId: string, clientId: string, redirectUri: string, userId: string, now: number, ttlSeconds: number): string {
    if (!projectId || !clientId || !redirectUri || !userId || ttlSeconds <= 0) throw new Error("invalid authorization code");
    const code = randomBytes(24).toString("base64url");
    this.codes.set(code, { projectId, clientId, redirectUri, userId, expiresAt: now + ttlSeconds * 1000, used: false });
    return code;
  }
  consume(code: string, projectId: string, clientId: string, redirectUri: string, now: number): { userId: string } {
    const record = this.codes.get(code);
    if (!record || record.used || record.expiresAt <= now || record.projectId !== projectId || record.clientId !== clientId || record.redirectUri !== redirectUri) throw new Error("authorization code unavailable");
    record.used = true;
    return { userId: record.userId };
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
    project.set(username, { username, userId: projectUserID(projectId, username), salt, passwordHash: this.hash(password, salt) });
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

  logout(projectId: string, token: string): void {
    const digest = sessionTokenDigest(token);
    const session = this.sessions.get(digest);
    if (!session || session.projectId !== projectId) throw new Error("invalid session");
    this.sessions.delete(digest);
  }

  guest(projectId: string) {
    if (!projectId) throw new Error("project id required");
    const token = randomBytes(32).toString("hex");
    const username = `guest-${randomBytes(8).toString("hex")}`;
    this.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: Date.now() + this.sessionTtlMs });
    return { token, username, userId: projectUserID(projectId, username), guest: true };
  }

  upgradeGuest(projectId: string, token: string, username: string, password: string) {
    const session = this.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || !session.username.startsWith("guest-")) throw new Error("guest session required");
    const userId = projectUserID(projectId, session.username);
    this.validate(projectId, username, password);
    const project = this.accounts.get(projectId) ?? new Map<string, Account>();
    if (project.has(username)) throw new Error("account already exists");
    const salt = randomBytes(16);
    project.set(username, { username, userId, salt, passwordHash: this.hash(password, salt) });
    this.accounts.set(projectId, project);
    session.username = username;
    return { username, userId };
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
