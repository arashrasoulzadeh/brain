import { request, type RpcRequest } from "../../sdk-typescript/src/index.js";

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
