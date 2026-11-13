import { createHash, createHmac, timingSafeEqual } from "node:crypto";
import { type PluginFunction, type RpcRequest, validateProjectId } from "../sdk-typescript/src/index.js";
import { sessionTokenDigest } from "./session-token-digest.js";
import { createAuthState, type AuthState } from "./auth-state.js";
import { AccountSignup } from "./account-signup.js";
import { AccountSignin } from "./account-signin.js";
import { AccountSessionReader } from "./account-session-reader.js";
import { GuestSessionIssuer } from "./guest-session-issuer.js";

export { guestRequest } from "./guest-request.js";
export { signInRequest } from "./sign-in-request.js";
export { signUpRequest } from "./sign-up-request.js";

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
export { EmailOTPIssuer } from "./email-otp-issuer.js";
export { EmailOTPVerifier } from "./email-otp-verifier.js";
export { SMSOTPIssuer } from "./sms-otp-issuer.js";
export { SMSOTPVerifier } from "./sms-otp-verifier.js";
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
export { RecoveryTokenIssuer } from "./recovery-token-issuer.js";
export { RecoveryTokenConsumer } from "./recovery-token-consumer.js";

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

export { createAuthState } from "./auth-state.js";
export { AccountSignup } from "./account-signup.js";
export { AccountSignin } from "./account-signin.js";
export { AccountSessionReader } from "./account-session-reader.js";
export { AccountSessionRotator } from "./account-session-rotator.js";
export { AccountSessionRevoker } from "./account-session-revoker.js";
export { AccountPasswordResetter } from "./account-password-resetter.js";
export { GuestSessionIssuer } from "./guest-session-issuer.js";
export { GuestAccountUpgrader } from "./guest-account-upgrader.js";
export type AuthOperations = ReturnType<typeof createAuthOperations>;
export function createAuthOperations(state: AuthState = createAuthState()) {
  const signup = new AccountSignup(state); const signin = new AccountSignin(state); const session = new AccountSessionReader(state); const guest = new GuestSessionIssuer(state);
  return { signup: signup.signup.bind(signup), signin: signin.signin.bind(signin), session: session.session.bind(session), guest: guest.guest.bind(guest) };
}

export function functions(service = createAuthOperations()): Readonly<Record<string, PluginFunction>> {
  const payload = (rpc: RpcRequest) => rpc.params.payload as Record<string, string> | undefined;
  return {
    signup: (rpc) => { const p = payload(rpc); return service.signup(rpc.params.projectId, p?.username ?? "", p?.password ?? ""); },
    signin: (rpc) => { const p = payload(rpc); return service.signin(rpc.params.projectId, p?.username ?? p?.provider ?? "", p?.password ?? String(p?.credential ?? "")); },
    session: (rpc) => { const p = payload(rpc); return service.session(rpc.params.projectId, p?.token ?? ""); },
    guest: (rpc) => service.guest(rpc.params.projectId),
  };
}
