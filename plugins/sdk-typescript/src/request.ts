import type { RpcRequest, WorkerParams } from "./index.js";
export function request(id: string, params: WorkerParams): RpcRequest {
  if (!id || !params.projectId || !params.pluginId || !params.version || !params.operation)
    throw new Error("invalid plugin request");
  return { jsonrpc: "2.0", id, method: "invoke", params };
}
