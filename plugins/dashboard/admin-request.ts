import { pluginRequest, type RpcRequest } from "../sdk-typescript/src/index.js";

export function adminRequest(projectId: string, pluginId: string, version: string, operation: string, payload: unknown): RpcRequest {
  return pluginRequest("dashboard", projectId, pluginId, version, operation, payload);
}
