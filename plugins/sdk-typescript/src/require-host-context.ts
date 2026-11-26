import type { HostExecutionContext, RpcRequest } from "./index.js";

export function requireHostContext(context: HostExecutionContext, request: RpcRequest): void {
  if (!context.projectId || !context.invocationId || request.params.projectId !== context.projectId)
    throw new Error("host project context mismatch");
}
