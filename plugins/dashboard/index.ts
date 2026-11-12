import {
  request,
  createHostHandler,
  type PluginHandler,
  type PluginRoute,
  type RpcRequest,
} from "../sdk-typescript/src/index.js";

export function projectCreateRequest(pluginId: string, version: string, id: string): RpcRequest {
  return request(`dashboard-project-${id}`, {
    protocol: 1,
    projectId: "operator",
    pluginId,
    version,
    operation: "project.create",
    payload: { id },
  });
}

export function adminRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: string,
  payload: unknown,
): RpcRequest {
  return request(`dashboard-${operation}-${projectId}`, {
    protocol: 1,
    projectId,
    pluginId,
    version,
    operation,
    payload,
  });
}

export const createHandler = createHostHandler;

export const handle: PluginHandler = () => { throw new Error("host route registry required"); };
