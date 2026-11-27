import type { ManifestOperation, ManifestOperations } from "./manifest-operations.generated.js";
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
    { ...(payload as Record<string, unknown>), [field]: identity }, identity,
  );
}

/** Serialize a declared operation for a fixed plugin ID. Host authorization remains authoritative. */
export function createManifestRequest<Id extends keyof ManifestOperations>(plugin: Id, field: string) {
  const serialize = createResourceRequest<ManifestOperation<Id>>(plugin, field);
  return (projectId: string, version: string, operation: ManifestOperation<Id>, identity: string, payload: unknown = {}): RpcRequest =>
    serialize(projectId, plugin, version, operation, identity, payload);
}
