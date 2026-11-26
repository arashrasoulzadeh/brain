import { deepEqual, rejects, throws } from "node:assert/strict";
import { dispatchRequest, guestRequest, signInRequest, signUpRequest } from "../../plugins/auth/index.js";
import { createFunctionHandler, pluginRequest } from "../../plugins/sdk-typescript/src/index.js";

const project = "sdk-auth-project";
const signup = signUpRequest(project, "auth", "0.1.0", "sdk-user", "correct horse battery");
await dispatchRequest("signup", signup);
const session = await dispatchRequest("signin", signInRequest(project, "auth", "0.1.0", "sdk-user", "correct horse battery")) as { token: string };
const request = pluginRequest("auth", project, "auth", "0.1.0", "auth.session", { token: session.token });
deepEqual(dispatchRequest("session", request), { username: "sdk-user" });
throws(() => dispatchRequest("session", { ...request, params: { ...request.params, projectId: "other-project" } }), /invalid session/);
await rejects(async () => dispatchRequest("signin", signInRequest("other-project", "auth", "0.1.0", "sdk-user", "correct horse battery")), /invalid credentials/);
const guest = dispatchRequest("guest", guestRequest(project, "auth", "0.1.0")) as { token: string; username: string };
deepEqual(dispatchRequest("session", pluginRequest("auth", project, "auth", "0.1.0", "auth.session", { token: guest.token })), { username: guest.username });
throws(() => dispatchRequest("constructor", request), /not exported/);

const handler = createFunctionHandler({}, [{ operation: "auth.session", function: "constructor", capability: "auth.session" }], { projectId: project, invocationId: "sdk-test" });
throws(() => handler(request), /route function not implemented/);
