import { request } from "../sdk-typescript/src/index.js";

export function matchmakingRequest(projectId: string, pluginId: string, version: string, operation: "matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign", config: string, payload: unknown = {}) {
  return request(`matchmaking-${operation}-${projectId}-${config}`, { protocol: 1, projectId, pluginId, version, operation, payload: { config, ...(payload as Record<string, unknown>) } });
}

