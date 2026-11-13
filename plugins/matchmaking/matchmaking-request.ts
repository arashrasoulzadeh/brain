import { pluginRequest } from "../sdk-typescript/src/index.js";

export function matchmakingRequest(projectId: string, pluginId: string, version: string, operation: "matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign", config: string, payload: unknown = {}) {
  return pluginRequest("matchmaking", projectId, pluginId, version, operation, { config, ...(payload as Record<string, unknown>) }, config);
}
