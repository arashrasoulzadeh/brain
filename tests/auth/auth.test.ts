import { handle, manifest, signInRequest } from "../../plugins/auth/src/index.js";

const request = signInRequest("dev", "guest", {});
if (manifest.id !== "auth" || request.params.operation !== "auth.signin") throw new Error("auth plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("auth handler contract failed");
