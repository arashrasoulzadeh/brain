import { createHash, createHmac, randomBytes, timingSafeEqual } from "node:crypto";
import { hash as argon2Hash, verify as argon2Verify } from "@node-rs/argon2";
import { type PluginFunction, type RpcRequest, validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";

export { guestRequest } from "./guest-request.js";
export { signInRequest } from "./sign-in-request.js";
export { signUpRequest } from "./sign-up-request.js";

type Account = { username: string; userId: string; passwordHash: string };

export function projectUserID(projectId: string, userId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(userId)) throw new Error("invalid project user id");
  return `${projectId}:${userId}`;
}

export { sessionTokenDigest } from "./session-token-digest.js";
export function createPKCEChallenge(verifier: string): string { return createHash("sha256").update(verifier).digest("base64url"); }
export function verifyPKCE(verifier: string, challenge: string, method: string): boolean {
  if (method !== "S256") return false;
  const expected = Buffer.from(createPKCEChallenge(verifier));
  const received = Buffer.from(challenge);
  return expected.length === received.length && timingSafeEqual(expected, received);
}
export { OIDCStateIssuer } from "./oidc-state-issuer.js";
export { OIDCStateConsumer } from "./oidc-state-consumer.js";
export { DeliveryWriter } from "./delivery-writer.js";
export { DeliveryReader } from "./delivery-reader.js";
export class EmailOTPService {
  private readonly challenges = new Map<string, { digest: string; expiresAt: number; used: boolean }>();
  constructor(private readonly secret: string, private readonly inbox: { deliver(destination: string, body: string): void }) {}
  issue(projectId: string, destination: string, purpose: string, now: number, ttlSeconds: number): void {
    validateProjectId(projectId);
    if (!destination.includes("@") || !purpose || ttlSeconds <= 0) throw new Error("invalid email OTP");
    const code = String(100000 + (randomBytes(4).readUInt32BE(0) % 900000));
    this.challenges.set(`${projectId}:${destination}:${purpose}`, { digest: this.digest(code), expiresAt: now + ttlSeconds * 1000, used: false });
    this.inbox.deliver(destination, code);
  }
  verify(projectId: string, destination: string, purpose: string, code: string, now: number): boolean {
    const challenge = this.challenges.get(`${projectId}:${destination}:${purpose}`);
    if (!challenge || challenge.used || challenge.expiresAt <= now || !timingSafeEqual(Buffer.from(challenge.digest), Buffer.from(this.digest(code)))) return false;
    challenge.used = true;
    return true;
  }
  private digest(code: string): string { return createHmac("sha256", this.secret).update(code).digest("hex"); }
}
export class SMSOTPService {
  private readonly challenges = new Map<string, { digest: string; expiresAt: number; used: boolean }>();
  constructor(private readonly secret: string, private readonly inbox: { deliver(destination: string, body: string): void }) {}
  issue(projectId: string, destination: string, purpose: string, now: number, ttlSeconds: number): void {
    validateProjectId(projectId);
    if (!/^\+[1-9][0-9]{7,14}$/.test(destination) || !purpose || ttlSeconds <= 0) throw new Error("invalid SMS OTP");
    const code = String(100000 + (randomBytes(4).readUInt32BE(0) % 900000));
    this.challenges.set(`${projectId}:${destination}:${purpose}`, { digest: this.digest(code), expiresAt: now + ttlSeconds * 1000, used: false });
    this.inbox.deliver(destination, code);
  }
  verify(projectId: string, destination: string, purpose: string, code: string, now: number): boolean {
    const challenge = this.challenges.get(`${projectId}:${destination}:${purpose}`);
    if (!challenge || challenge.used || challenge.expiresAt <= now || !timingSafeEqual(Buffer.from(challenge.digest), Buffer.from(this.digest(code)))) return false;
    challenge.used = true;
    return true;
  }
  private digest(code: string): string { return createHmac("sha256", this.secret).update(code).digest("hex"); }
}
export function normalizePhoneNumber(value: string): string {
  const compact = value.replace(/[\s().-]/g, "");
  const normalized = compact.startsWith("00") ? `+${compact.slice(2)}` : compact;
  if (!/^\+[1-9][0-9]{7,14}$/.test(normalized)) throw new Error("invalid phone number");
  return normalized;
}
export class OTPLimiter {
  private readonly counts = new Map<string, { startedAt: number; count: number }>();
  constructor(private readonly limit: number, private readonly windowMs: number) {}
  allow(projectId: string, destination: string, now: number): boolean {
    validateProjectId(projectId);
    const key = `${projectId}:${destination}`;
    const current = this.counts.get(key);
    if (!current || now - current.startedAt >= this.windowMs) { this.counts.set(key, { startedAt: now, count: 1 }); return true; }
    if (current.count >= this.limit) return false;
    current.count += 1;
    return true;
  }
}
export class EmailRecoveryService {
  private readonly tokens = new Map<string, { digest: string; expiresAt: number; used: boolean }>();
  constructor(private readonly secret: string, private readonly inbox: { deliver(destination: string, body: string): void }) {}
  issue(projectId: string, destination: string, now: number, ttlSeconds: number): void {
    validateProjectId(projectId);
    if (!destination.includes("@") || ttlSeconds <= 0) throw new Error("invalid recovery request");
    const token = randomBytes(24).toString("base64url");
    this.tokens.set(`${projectId}:${destination}`, { digest: createHmac("sha256", this.secret).update(token).digest("hex"), expiresAt: now + ttlSeconds * 1000, used: false });
    this.inbox.deliver(destination, token);
  }
  consume(projectId: string, destination: string, token: string, now: number): boolean {
    const record = this.tokens.get(`${projectId}:${destination}`);
    if (!record || record.used || record.expiresAt <= now) return false;
    const digest = createHmac("sha256", this.secret).update(token).digest("hex");
    if (!timingSafeEqual(Buffer.from(record.digest), Buffer.from(digest))) return false;
    record.used = true;
    return true;
  }
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

export { OperatorSessionIssuer } from "./operator-session-issuer.js";
export { OperatorSessionValidator } from "./operator-session-validator.js";
export { OperatorSessionRevoker } from "./operator-session-revoker.js";

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
export { OIDCConfigWriter } from "./oidc-config-writer.js";
export { OIDCConfigReader } from "./oidc-config-reader.js";

export type OIDCTokenClaims = { issuer: string; audience: string; expiresAt: number; nonce: string; signature: string };
export function validateOIDCToken(claims: OIDCTokenClaims, config: OIDCConfig, expectedNonce: string, now: number, verifySignature: (signature: string) => boolean): void {
  if (claims.issuer !== config.issuer || claims.audience !== config.audience || claims.expiresAt <= now || claims.nonce !== expectedNonce || !verifySignature(claims.signature)) throw new Error("invalid OIDC token");
}
export function requireRecentAuthentication(authenticatedAt: number, now: number, maxAgeMs: number): void {
  if (maxAgeMs <= 0 || now - authenticatedAt > maxAgeMs) throw new Error("recent authentication required");
}
export function publicAuthError(error: unknown): string {
  if (error instanceof Error && ["account already exists", "invalid credentials", "account not found"].includes(error.message)) return "invalid credentials";
  return "authentication failed";
}
export function validateDeliveryConfig(environment: string, provider: string): void {
  if (environment === "production" && ["local", "fake", "test"].includes(provider)) throw new Error("fake delivery forbidden in production");
  if (!environment || !provider) throw new Error("invalid delivery configuration");
}

export { AuthorizationCodeIssuer } from "./authorization-code-issuer.js";
export { AuthorizationCodeConsumer } from "./authorization-code-consumer.js";
export type { AuthorizationCode } from "./authorization-code-types.js";

export class AuthService {
  private readonly accounts = new Map<string, Map<string, Account>>();
  private readonly sessions = new Map<string, { projectId: string; username: string; expiresAt: number }>();
  private readonly rotated = new Set<string>();
  constructor(private readonly sessionTtlMs = 24 * 60 * 60 * 1000) {
    if (sessionTtlMs <= 0) throw new Error("invalid session ttl");
  }

  async signup(projectId: string, username: string, password: string) {
    this.validate(projectId, username, password);
    const project = this.accounts.get(projectId) ?? new Map<string, Account>();
    if (project.has(username)) throw new Error("account already exists");
    const passwordHash = await this.hash(password);
    project.set(username, { username, userId: projectUserID(projectId, username), passwordHash });
    this.accounts.set(projectId, project);
    return { username, userId: projectUserID(projectId, username) };
  }

  async signin(projectId: string, username: string, password: string, now = Date.now()) {
    validateProjectId(projectId);
    const account = this.accounts.get(projectId)?.get(username);
    if (!account || !(await argon2Verify(account.passwordHash, password))) throw new Error("invalid credentials");
    const token = randomBytes(32).toString("hex");
    this.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: now + this.sessionTtlMs });
    return { token, username };
  }

  session(projectId: string, token: string, now = Date.now()) {
    validateProjectId(projectId);
    const session = this.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || session.expiresAt <= now) throw new Error("invalid session");
    return { username: session.username };
  }

  rotate(projectId: string, token: string, now = Date.now()): string {
    validateProjectId(projectId);
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
    validateProjectId(projectId);
    const digest = sessionTokenDigest(token);
    const session = this.sessions.get(digest);
    if (!session || session.projectId !== projectId) throw new Error("invalid session");
    this.sessions.delete(digest);
  }

  async resetPassword(projectId: string, username: string, password: string): Promise<void> {
    this.validate(projectId, username, password);
    const account = this.accounts.get(projectId)?.get(username);
    if (!account) throw new Error("account not found");
    account.passwordHash = await this.hash(password);
    for (const [digest, session] of this.sessions) if (session.projectId === projectId && session.username === username) this.sessions.delete(digest);
  }

  guest(projectId: string) {
    validateProjectId(projectId);
    const token = randomBytes(32).toString("hex");
    const username = `guest-${randomBytes(8).toString("hex")}`;
    this.sessions.set(sessionTokenDigest(token), { projectId, username, expiresAt: Date.now() + this.sessionTtlMs });
    return { token, username, userId: projectUserID(projectId, username), guest: true };
  }

  async upgradeGuest(projectId: string, token: string, username: string, password: string) {
    validateProjectId(projectId);
    const session = this.sessions.get(sessionTokenDigest(token));
    if (!session || session.projectId !== projectId || !session.username.startsWith("guest-")) throw new Error("guest session required");
    const userId = projectUserID(projectId, session.username);
    this.validate(projectId, username, password);
    const project = this.accounts.get(projectId) ?? new Map<string, Account>();
    if (project.has(username)) throw new Error("account already exists");
    const salt = randomBytes(16);
    project.set(username, { username, userId, passwordHash: await this.hash(password) });
    this.accounts.set(projectId, project);
    session.username = username;
    return { username, userId };
  }

  private hash(password: string) { return argon2Hash(password, { memoryCost: 19456, timeCost: 2, parallelism: 1, outputLen: 32 }); }
  private validate(projectId: string, username: string, password: string) {
    validateProjectId(projectId);
    if (!username || password.length < 8) throw new Error("invalid signup");
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
