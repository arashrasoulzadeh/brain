import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "rooms",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["rooms.create", "rooms.join", "rooms.reconnect", "rooms.result"],
  extensionPoints: ["rooms.runtime"],
  configSchema: { maxPlayers: "int" },
} as const;

export function roomRequest(projectId: string, operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result", roomId: string, payload: unknown = {}) {
  return request(`rooms-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { roomId, ...payload } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("rooms.")) throw new Error("unsupported rooms operation");
  return { accepted: true, operation: rpc.params.operation };
};
