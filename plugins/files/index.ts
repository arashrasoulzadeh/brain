import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function fileRequest(projectId: string, pluginId: string, version: string, operation: "files.upload" | "files.download" | "files.metadata" | "files.publish", objectId: string, payload: unknown = {}) {
  return request(`files-${operation}-${projectId}-${objectId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { objectId, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
