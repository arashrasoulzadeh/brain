import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function roomRequest(projectId: string, pluginId: string, version: string, operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result", roomId: string, payload: unknown = {}) {
  return request(`rooms-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
