export type WorkerParams = {
  protocol: number;
  projectId: string;
  pluginId: string;
  version: string;
  operation: string;
  payload?: unknown;
};

export type RpcRequest = {
  jsonrpc: "2.0";
  id: string;
  method: "invoke";
  params: WorkerParams;
};

export type RpcResponse = {
  jsonrpc: "2.0";
  id: string;
  result?: { protocol: number; success: boolean; payload?: unknown };
  errorCode?: string;
};

export type PluginHandler = (request: RpcRequest) => Promise<unknown> | unknown;

export type PluginRoute = {
  operation: string;
  function: string;
  capability: string;
};

export type HostInvoker = {
  invoke(projectId: string, operation: string, capability: string, payload: unknown): unknown;
};

export type HostExecutionContext = {
  readonly projectId: string;
  readonly invocationId: string;
};

export type PluginFunction = (request: RpcRequest) => unknown;
export { createRouteHandler } from "./create-route-handler.js";
export { createHostHandler } from "./create-host-handler.js";
export { createFunctionHandler } from "./create-function-handler.js";
export { request } from "./request.js";
export { pluginRequest } from "./plugin-request.js";
export { validateProjectId } from "./validate-project-id.js";
export { validateIdentifier } from "./validate-identifier.js";
