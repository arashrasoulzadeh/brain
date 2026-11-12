import { AuthService, functions, guestRequest, projectUserID, signInRequest, signUpRequest } from "../../plugins/auth/index.js";
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

const service = new AuthService();
const registration = service.signup("project-a", "alice", "correct horse battery");
if (registration.username !== "alice" || "email" in registration) throw new Error("password registration attempted email delivery");
if (!projectUserID("project-a", "user-1").startsWith("project-a:")) throw new Error("project-scoped user ID failed");
try { service.signup("project-a", "alice", "another password"); throw new Error("duplicate username accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "account already exists") throw error; }
service.signup("project-b", "alice", "correct horse battery");
const session = service.signin("project-a", "alice", "correct horse battery");
if (service.session("project-a", session.token).username !== "alice") throw new Error("session lookup failed");
try { service.session("project-b", session.token); throw new Error("cross-project session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { service.signin("project-a", "alice", "wrong password"); throw new Error("invalid password accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid credentials") throw error; }
