import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "arena",
  version: "0.1.0",
  hostApi: 1,
  dependencies: ["rooms@0.1.0", "matchmaking@0.1.0"],
  capabilities: ["arena.start", "arena.input", "arena.snapshot", "arena.result"],
  extensionPoints: [],
  configSchema: { tickRate: "int", maxPlayers: "int" },
} as const;

export function arenaRequest(projectId: string, operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result", roomId: string, payload: unknown = {}) {
  return request(`arena-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { roomId, ...payload } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("arena.")) throw new Error("unsupported arena operation");
  return { accepted: true, operation: rpc.params.operation };
};
