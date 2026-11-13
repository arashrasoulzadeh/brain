import { pluginRequest } from "../sdk-typescript/src/index.js";

export function invoke(projectId: string, pluginId: string, version: string, operation: string) {
  return pluginRequest(pluginId, projectId, pluginId, version, operation);
}
