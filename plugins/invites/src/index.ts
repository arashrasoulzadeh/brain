import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "invites",
  version: "0.1.0",
  hostApi: 1,
  dependencies: ["rooms@0.1.0"],
  capabilities: ["invites.create", "invites.accept", "invites.cancel"],
  extensionPoints: [],
  configSchema: { ttlSeconds: "int" },
} as const;

export function inviteRequest(projectId: string, operation: "invites.create" | "invites.accept" | "invites.cancel", roomId: string, payload: unknown = {}) {
  return request(`invites-${operation}-${projectId}-${roomId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { roomId, ...payload } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("invites.")) throw new Error("unsupported invites operation");
  return { accepted: true, operation: rpc.params.operation };
};
