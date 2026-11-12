import { AccountRateLimiter, AuthService, IdentityLinker, OperatorSessionRegistry, ProjectRateLimiter, functions, guestRequest, projectUserID, sessionTokenDigest, signInRequest, signUpRequest } from "../../plugins/auth/index.js";
import { createFunctionHandler } from "../../plugins/sdk-typescript/src/index.js";

const serviceForHandler = new AuthService();
const handler = createFunctionHandler(functions(serviceForHandler), [
  { operation: "auth.signup", function: "signup", capability: "auth.signup" },
  { operation: "auth.signin", function: "signin", capability: "auth.signin" },
  { operation: "auth.session", function: "session", capability: "auth.session" },
  { operation: "auth.guest", function: "guest", capability: "auth.guest" },
]);
const signup = signUpRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
if (signup.params.pluginId !== "auth" || signup.params.operation !== "auth.signup") throw new Error("auth plugin contract failed");
if (typeof handler(signup) !== "object") throw new Error("auth signup handler contract failed");
const request = signInRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
if (typeof handler(request) !== "object") throw new Error("auth handler contract failed");
const guest = (handler(guestRequest("dev", "auth", "0.1.0")) as { payload: { token: string; username: string; guest: boolean } }).payload;
if (!guest.token || !guest.username.startsWith("guest-") || !guest.guest) throw new Error("guest session failed");
const guestService = new AuthService();
const guestAccount = guestService.guest("project-a");
const upgraded = guestService.upgradeGuest("project-a", guestAccount.token, "dana", "correct horse battery");
if (upgraded.userId !== guestAccount.userId || upgraded.username !== "dana") throw new Error("guest upgrade changed user id");
const links = new IdentityLinker();
links.link("project-a:user-1", "google", "subject-1");
try { links.link("project-a:user-2", "google", "subject-1"); throw new Error("identity-link conflict accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "identity already linked") throw error; }

const service = new AuthService();
const registration = service.signup("project-a", "alice", "correct horse battery");
if (registration.username !== "alice" || "email" in registration) throw new Error("password registration attempted email delivery");
if (!projectUserID("project-a", "user-1").startsWith("project-a:")) throw new Error("project-scoped user ID failed");
try { service.signup("project-a", "alice", "another password"); throw new Error("duplicate username accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "account already exists") throw error; }
service.signup("project-b", "alice", "correct horse battery");
const session = service.signin("project-a", "alice", "correct horse battery");
if (!session.token || session.username !== "alice") throw new Error("password sign-in failed");
if (service.session("project-a", session.token).username !== "alice") throw new Error("session lookup failed");
if (sessionTokenDigest(session.token) === session.token) throw new Error("session token was not digested");
const expiring = new AuthService(1000);
expiring.signup("project-a", "bob", "correct horse battery");
const expiringSession = expiring.signin("project-a", "bob", "correct horse battery", 100);
if (expiring.session("project-a", expiringSession.token, 1099).username !== "bob") throw new Error("valid session expired early");
try { expiring.session("project-a", expiringSession.token, 1100); throw new Error("expired session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
const rotating = new AuthService();
rotating.signup("project-a", "carol", "correct horse battery");
const originalToken = rotating.signin("project-a", "carol", "correct horse battery", 200).token;
const rotatedToken = rotating.rotate("project-a", originalToken, 201);
if (rotatedToken === originalToken || rotating.session("project-a", rotatedToken, 202).username !== "carol") throw new Error("session rotation failed");
try { rotating.session("project-a", originalToken, 202); throw new Error("rotated token remained valid"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { rotating.rotate("project-a", originalToken, 203); throw new Error("refresh token reuse accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "refresh token reuse detected") throw error; }
const logoutToken = rotating.signin("project-a", "carol", "correct horse battery", 300).token;
rotating.logout("project-a", logoutToken);
try { rotating.session("project-a", logoutToken, 301); throw new Error("logged out session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
const operators = new OperatorSessionRegistry();
const operatorToken = operators.issue("admin");
if (!operators.valid(operatorToken, "admin") || operators.valid(operatorToken, "other")) throw new Error("operator session isolation failed");
operators.revoke(operatorToken);
if (operators.valid(operatorToken, "admin")) throw new Error("operator session revocation failed");
try { service.session("project-b", session.token); throw new Error("cross-project session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { service.signin("project-a", "alice", "wrong password"); throw new Error("invalid password accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid credentials") throw error; }
const accountLimiter = new AccountRateLimiter(2, 1000);
if (!accountLimiter.allow("project-a", "alice", 100) || !accountLimiter.allow("project-a", "alice", 101) || accountLimiter.allow("project-a", "alice", 102)) throw new Error("account rate limit failed");
const projectLimiter = new ProjectRateLimiter(1, 1000);
if (!projectLimiter.allow("project-a", 100) || projectLimiter.allow("project-a", 101) || !projectLimiter.allow("project-b", 101)) throw new Error("project rate limit failed");
