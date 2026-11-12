import { request } from "../sdk-typescript/src/index.js";

export function invoke(projectId: string, pluginId: string, version: string, operation: string) {
  return request(`${pluginId}-${operation}`, { protocol: 1, projectId, pluginId, version, operation });
}
