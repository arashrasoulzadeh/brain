import { request, type PluginHandler, type RpcRequest } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "dashboard",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["project.manage", "plugin.install", "plugin.grant", "plugin.enable"],
  extensionPoints: ["admin.navigation", "admin.page"],
  configSchema: { title: "string" },
} as const;

export function projectCreateRequest(id: string): RpcRequest {
  return request(`dashboard-project-${id}`, {
    protocol: 1, projectId: "operator", pluginId: manifest.id, version: manifest.version,
    operation: "project.create", payload: { id },
  });
}

export function adminRequest(projectId: string, operation: string, payload: unknown): RpcRequest {
  return request(`dashboard-${operation}-${projectId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload });
}

export const handle: PluginHandler = (rpc) => {
  const supported = ["project.create", "plugin.install", "plugin.grant", "plugin.enable"];
  if (!supported.includes(rpc.params.operation)) throw new Error("unsupported dashboard operation");
  return { accepted: true, operation: rpc.params.operation, payload: rpc.params.payload ?? null };
};
