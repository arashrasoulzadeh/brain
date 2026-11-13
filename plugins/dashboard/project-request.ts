import { pluginRequest, type RpcRequest } from "../sdk-typescript/src/index.js";

export function projectCreateRequest(pluginId: string, version: string, id: string): RpcRequest {
  return pluginRequest("dashboard", "operator", pluginId, version, "project.create", { id }, `project-${id}`);
}
