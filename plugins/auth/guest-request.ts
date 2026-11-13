import { pluginRequest } from "../sdk-typescript/src/index.js";

export function guestRequest(projectId: string, pluginId: string, version: string) {
  return pluginRequest("auth", projectId, pluginId, version, "auth.guest", {}, "guest");
}
