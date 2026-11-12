import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function inviteRequest(projectId: string, pluginId: string, version: string, operation: "invites.create" | "invites.accept" | "invites.cancel", roomId: string, payload: unknown = {}) {
  return request(`invites-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
