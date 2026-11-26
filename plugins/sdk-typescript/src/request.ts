import type { RpcRequest, WorkerParams } from "./index.js";
export function request(id: string, params: WorkerParams): RpcRequest {
  if (!id || !params.projectId || !params.pluginId || !params.version || !params.operation)
    throw new Error("invalid plugin request");
  return { jsonrpc: "2.0", id, method: "invoke", params };
}

export function pluginRequest(
  prefix: string,
  projectId: string,
  pluginId: string,
  version: string,
  operation: string,
  payload: unknown = {},
  identity = "",
): RpcRequest {
  return request(`${prefix}-${operation}-${projectId}${identity ? `-${identity}` : ""}`, {
    protocol: 1,
    projectId,
    pluginId,
    version,
    operation,
    payload,
  });
}

/** A request serializer only; host authorization must bind the resource and project. */
export function createResourceRequest<Operation extends string>(prefix: string, field: string) {
  return (
    projectId: string, pluginId: string, version: string,
    operation: Operation, identity: string, payload: unknown = {},
  ): RpcRequest => pluginRequest(
    prefix, projectId, pluginId, version, operation,
    { [field]: identity, ...(payload as Record<string, unknown>) }, identity,
  );
}
