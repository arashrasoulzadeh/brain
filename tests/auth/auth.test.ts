import { AccountRateLimiter, AuthService, AuthorizationCodeIssuer, AuthorizationCodeConsumer, DeliveryReader, DeliveryWriter, EmailOTPIssuer, EmailOTPVerifier, EmailRecoveryService, IdentityLinker, OIDCConfigWriter, OIDCConfigReader, OIDCStateIssuer, OIDCStateConsumer, OTPLimiter, OperatorSessionIssuer, OperatorSessionValidator, OperatorSessionRevoker, ProjectRateLimiter, SMSOTPIssuer, SMSOTPVerifier, createPKCEChallenge, functions, guestRequest, normalizePhoneNumber, projectUserID, publicAuthError, requireRecentAuthentication, sessionTokenDigest, signInRequest, signUpRequest, validateDeliveryConfig, validateOIDCToken, verifyPKCE } from "../../plugins/auth/index.js";
import { createFunctionHandler } from "../../plugins/sdk-typescript/src/index.js";

const serviceForHandler = new AuthService();
const handler = createFunctionHandler(functions(serviceForHandler), [
  { operation: "auth.signup", function: "signup", capability: "auth.signup" },
  { operation: "auth.signin", function: "signin", capability: "auth.signin" },
  { operation: "auth.session", function: "session", capability: "auth.session" },
  { operation: "auth.guest", function: "guest", capability: "auth.guest" },
], { projectId: "dev", invocationId: "auth-test" });
const signup = signUpRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
if (signup.params.pluginId !== "auth" || signup.params.operation !== "auth.signup") throw new Error("auth plugin contract failed");
const signupResponse = await handler(signup) as { payload: Promise<unknown> };
await signupResponse.payload;
const request = signInRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
const signinResponse = await handler(request) as { payload: Promise<unknown> };
await signinResponse.payload;
const guest = (handler(guestRequest("dev", "auth", "0.1.0")) as { payload: { token: string; username: string; guest: boolean } }).payload;
if (!guest.token || !guest.username.startsWith("guest-") || !guest.guest) throw new Error("guest session failed");
const guestService = new AuthService();
const guestAccount = guestService.guest("project-a");
const upgraded = await guestService.upgradeGuest("project-a", guestAccount.token, "dana", "correct horse battery");
if (upgraded.userId !== guestAccount.userId || upgraded.username !== "dana") throw new Error("guest upgrade changed user id");
const links = new IdentityLinker();
links.link("project-a:user-1", "google", "subject-1");
try { links.link("project-a:user-2", "google", "subject-1"); throw new Error("identity-link conflict accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "identity already linked") throw error; }
const oidcMap = new Map(); const oidcWriter = new OIDCConfigWriter(oidcMap); const oidcReader = new OIDCConfigReader(oidcMap);
oidcWriter.set("project-a", { issuer: "https://accounts.google.com", audience: "client-id", clientId: "client-id", redirectUri: "https://app.example/callback" });
if (oidcReader.get("project-a").issuer !== "https://accounts.google.com") throw new Error("OIDC configuration failed");
validateOIDCToken({ issuer: "https://accounts.google.com", audience: "client-id", expiresAt: 2000, nonce: "nonce-1", signature: "sig" }, oidcReader.get("project-a"), "nonce-1", 1000, (signature) => signature === "sig");
requireRecentAuthentication(1000, 1200, 300);
try { requireRecentAuthentication(1000, 1301, 300); throw new Error("stale authentication accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "recent authentication required") throw error; }
const codeMap = new Map(); const codeIssuer = new AuthorizationCodeIssuer(codeMap); const codeConsumer = new AuthorizationCodeConsumer(codeMap);
const code = codeIssuer.issue("project-a", "client-id", "https://app.example/callback", "user-1", 1000, 60);
if (codeConsumer.consume(code, "project-a", "client-id", "https://app.example/callback", 1001).userId !== "user-1") throw new Error("authorization code exchange failed");
try { codeConsumer.consume(code, "project-a", "client-id", "https://app.example/callback", 1002); throw new Error("authorization code replay accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "authorization code unavailable") throw error; }
if (!verifyPKCE("verifier", createPKCEChallenge("verifier"), "S256") || verifyPKCE("wrong", createPKCEChallenge("verifier"), "S256")) throw new Error("PKCE verification failed");
if (verifyPKCE("verifier", "short", "S256")) throw new Error("malformed PKCE challenge accepted");
const states = new Map(); const stateIssuer = new OIDCStateIssuer(states); const stateConsumer = new OIDCStateConsumer(states);
const state = stateIssuer.issue("project-a", "client-id", "https://app.example/callback");
stateConsumer.consume(state, "project-a", "client-id", "https://app.example/callback");
try { stateConsumer.consume(state, "project-a", "client-id", "https://app.example/callback"); throw new Error("callback state replay accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "callback state mismatch") throw error; }
const inboxMessages = new Map<string, string[]>(); const deliveryWriter = new DeliveryWriter(inboxMessages); const deliveryReader = new DeliveryReader(inboxMessages);
deliveryWriter.deliver("alice@example.com", "code-1");
if (deliveryReader.read("alice@example.com") !== "code-1") throw new Error("test delivery inbox failed");
const delivery = { deliver: (destination: string, body: string) => deliveryWriter.deliver(destination, body) };
const emailChallenges = new Map(); const otpIssuer = new EmailOTPIssuer(emailChallenges, "test-secret", delivery); const otpVerifier = new EmailOTPVerifier(emailChallenges, "test-secret");
otpIssuer.issue("project-a", "alice@example.com", "signin", 1000, 300);
const emailCode = deliveryReader.read("alice@example.com");
if (!emailCode || !otpVerifier.verify("project-a", "alice@example.com", "signin", emailCode, 1001)) throw new Error("email OTP failed");
const smsChallenges = new Map(); const smsIssuer = new SMSOTPIssuer(smsChallenges, "test-secret", delivery); const smsVerifier = new SMSOTPVerifier(smsChallenges, "test-secret");
smsIssuer.issue("project-a", "+989121234567", "signin", 1000, 300);
const smsCode = deliveryReader.read("+989121234567");
if (!smsCode || !smsVerifier.verify("project-a", "+989121234567", "signin", smsCode, 1001)) throw new Error("SMS OTP failed");
if (normalizePhoneNumber("0098 912-123-4567") !== "+989121234567") throw new Error("phone normalization failed");
const otpLimiter = new OTPLimiter(2, 1000);
if (!otpLimiter.allow("project-a", "alice@example.com", 100) || !otpLimiter.allow("project-a", "alice@example.com", 101) || otpLimiter.allow("project-a", "alice@example.com", 102)) throw new Error("OTP send limit failed");
const recovery = new EmailRecoveryService("recovery-secret", delivery);
try { recovery.issue("bad/project", "alice@example.com", 1000, 60); throw new Error("invalid recovery project accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid project id") throw error; }
recovery.issue("project-a", "alice@example.com", 1000, 60);
const recoveryCode = deliveryReader.read("alice@example.com");
if (!recovery.consume("project-a", "alice@example.com", recoveryCode ?? "", 1001)) throw new Error("email recovery failed");
const resetService = new AuthService();
await resetService.signup("project-a", "erin", "old password");
const resetSession = await resetService.signin("project-a", "erin", "old password", 1000);
await resetService.resetPassword("project-a", "erin", "new password");
try { resetService.session("project-a", resetSession.token, 1001); throw new Error("session survived password reset"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
if (publicAuthError(new Error("account already exists")) !== "invalid credentials" || publicAuthError(new Error("invalid credentials")) !== "invalid credentials") throw new Error("account enumeration signal leaked");
validateDeliveryConfig("test", "local");
try { validateDeliveryConfig("production", "local"); throw new Error("fake production delivery accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "fake delivery forbidden in production") throw error; }

const service = new AuthService();
const registration = await service.signup("project-a", "alice", "correct horse battery");
if (registration.username !== "alice" || "email" in registration) throw new Error("password registration attempted email delivery");
if (!projectUserID("project-a", "user-1").startsWith("project-a:")) throw new Error("project-scoped user ID failed");
try { await service.signup("project-a", "alice", "another password"); throw new Error("duplicate username accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "account already exists") throw error; }
await service.signup("project-b", "alice", "correct horse battery");
const session = await service.signin("project-a", "alice", "correct horse battery");
if (!session.token || session.username !== "alice") throw new Error("password sign-in failed");
if (service.session("project-a", session.token).username !== "alice") throw new Error("session lookup failed");
if (sessionTokenDigest(session.token) === session.token) throw new Error("session token was not digested");
const expiring = new AuthService(1000);
await expiring.signup("project-a", "bob", "correct horse battery");
const expiringSession = await expiring.signin("project-a", "bob", "correct horse battery", 100);
if (expiring.session("project-a", expiringSession.token, 1099).username !== "bob") throw new Error("valid session expired early");
try { expiring.session("project-a", expiringSession.token, 1100); throw new Error("expired session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
const rotating = new AuthService();
await rotating.signup("project-a", "carol", "correct horse battery");
const originalToken = (await rotating.signin("project-a", "carol", "correct horse battery", 200)).token;
const rotatedToken = rotating.rotate("project-a", originalToken, 201);
if (rotatedToken === originalToken || rotating.session("project-a", rotatedToken, 202).username !== "carol") throw new Error("session rotation failed");
try { rotating.session("project-a", originalToken, 202); throw new Error("rotated token remained valid"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { rotating.rotate("project-a", originalToken, 203); throw new Error("refresh token reuse accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "refresh token reuse detected") throw error; }
const logoutToken = (await rotating.signin("project-a", "carol", "correct horse battery", 300)).token;
rotating.logout("project-a", logoutToken);
try { rotating.session("project-a", logoutToken, 301); throw new Error("logged out session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
const operatorSessions = new Map<string, string>(); const revokedOperators = new Set<string>();
const operatorToken = new OperatorSessionIssuer(operatorSessions).issue("admin");
if (!new OperatorSessionValidator(operatorSessions, revokedOperators).valid(operatorToken, "admin") || new OperatorSessionValidator(operatorSessions, revokedOperators).valid(operatorToken, "other")) throw new Error("operator session isolation failed");
new OperatorSessionRevoker(operatorSessions, revokedOperators).revoke(operatorToken);
if (new OperatorSessionValidator(operatorSessions, revokedOperators).valid(operatorToken, "admin")) throw new Error("operator session revocation failed");
try { service.session("project-b", session.token); throw new Error("cross-project session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { await service.signin("project-a", "alice", "wrong password"); throw new Error("invalid password accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid credentials") throw error; }
const accountLimiter = new AccountRateLimiter(2, 1000);
if (!accountLimiter.allow("project-a", "alice", 100) || !accountLimiter.allow("project-a", "alice", 101) || accountLimiter.allow("project-a", "alice", 102)) throw new Error("account rate limit failed");
const projectLimiter = new ProjectRateLimiter(1, 1000);
if (!projectLimiter.allow("project-a", 100) || projectLimiter.allow("project-a", 101) || !projectLimiter.allow("project-b", 101)) throw new Error("project rate limit failed");
