import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "tables",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["tables.read", "tables.write", "tables.schema"],
  extensionPoints: ["data.policy"],
  configSchema: { exposure: "string" },
} as const;

export function dataRequest(projectId: string, operation: "tables.read" | "tables.write" | "tables.schema", table: string, payload: unknown = {}) {
  return request(`tables-${operation}-${projectId}-${table}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { table, ...payload } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("tables.")) throw new Error("unsupported tables operation");
  return { accepted: true, operation: rpc.params.operation };
};
