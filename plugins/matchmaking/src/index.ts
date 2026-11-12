import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "matchmaking",
  version: "0.1.0",
  hostApi: 1,
  dependencies: ["rooms@0.1.0"],
  capabilities: ["matchmaking.queue", "matchmaking.cancel", "matchmaking.status", "matchmaking.assign"],
  extensionPoints: ["matchmaking.policy"],
  configSchema: { ticketTtlSeconds: "int", matchSize: "int" },
} as const;

export function matchmakingRequest(projectId: string, operation: "matchmaking.queue" | "matchmaking.cancel" | "matchmaking.status" | "matchmaking.assign", config: string, payload: unknown = {}) {
  return request(`matchmaking-${operation}-${projectId}-${config}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { config, ...(payload as Record<string, unknown>) } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("matchmaking.")) throw new Error("unsupported matchmaking operation");
  return { accepted: true, operation: rpc.params.operation };
};
