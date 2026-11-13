import { pluginRequest } from "../sdk-typescript/src/index.js";

export function signInRequest(projectId: string, pluginId: string, version: string, provider: string, credential: unknown) {
  return pluginRequest("auth", projectId, pluginId, version, "auth.signin", { provider, credential }, "signin");
}
