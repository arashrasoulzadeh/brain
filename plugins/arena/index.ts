import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function arenaRequest(projectId: string, pluginId: string, version: string, operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result", roomId: string, payload: unknown = {}) {
  return request(`arena-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { roomId, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
