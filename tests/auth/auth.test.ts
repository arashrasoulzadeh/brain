import { AuthService, functions, signInRequest, signUpRequest } from "../../plugins/auth/index.js";
import { createFunctionHandler } from "../../plugins/sdk-typescript/src/index.js";

const serviceForHandler = new AuthService();
const handler = createFunctionHandler(functions(serviceForHandler), [
  { operation: "auth.signup", function: "signup", capability: "auth.signup" },
  { operation: "auth.signin", function: "signin", capability: "auth.signin" },
  { operation: "auth.session", function: "session", capability: "auth.session" },
]);
const signup = signUpRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
if (signup.params.pluginId !== "auth" || signup.params.operation !== "auth.signup") throw new Error("auth plugin contract failed");
if (typeof handler(signup) !== "object") throw new Error("auth signup handler contract failed");
const request = signInRequest("dev", "auth", "0.1.0", "guest", "correct horse battery");
if (typeof handler(request) !== "object") throw new Error("auth handler contract failed");

const service = new AuthService();
service.signup("project-a", "alice", "correct horse battery");
const session = service.signin("project-a", "alice", "correct horse battery");
if (service.session("project-a", session.token).username !== "alice") throw new Error("session lookup failed");
try { service.session("project-b", session.token); throw new Error("cross-project session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { service.signin("project-a", "alice", "wrong password"); throw new Error("invalid password accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid credentials") throw error; }
