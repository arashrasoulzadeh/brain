import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "files",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["files.upload", "files.download", "files.metadata", "files.publish"],
  extensionPoints: ["files.storage"],
  configSchema: { bucket: "string" },
} as const;

export function fileRequest(projectId: string, operation: "files.upload" | "files.download" | "files.metadata" | "files.publish", objectId: string, payload: unknown = {}) {
  return request(`files-${operation}-${projectId}-${objectId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { objectId, ...payload } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("files.")) throw new Error("unsupported files operation");
  return { accepted: true, operation: rpc.params.operation };
};
