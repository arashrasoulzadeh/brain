import { AuthService, createHandler, manifest, signInRequest, signUpRequest } from "../../plugins/auth/src/index.js";

const serviceForHandler = new AuthService();
const handler = createHandler(serviceForHandler);
const signup = signUpRequest("dev", "guest", "correct horse battery");
if (manifest.id !== "auth" || signup.params.operation !== "auth.signup") throw new Error("auth plugin contract failed");
if (typeof handler(signup) !== "object") throw new Error("auth signup handler contract failed");
const request = signInRequest("dev", "guest", "correct horse battery");
if (typeof handler(request) !== "object") throw new Error("auth handler contract failed");

const service = new AuthService();
service.signup("project-a", "alice", "correct horse battery");
const session = service.signin("project-a", "alice", "correct horse battery");
if (service.session("project-a", session.token).username !== "alice") throw new Error("session lookup failed");
try { service.session("project-b", session.token); throw new Error("cross-project session accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid session") throw error; }
try { service.signin("project-a", "alice", "wrong password"); throw new Error("invalid password accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid credentials") throw error; }
