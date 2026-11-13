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

export function createRouteHandler(routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = routes.find((candidate) => candidate.operation === rpc.params.operation);
    if (!route || !route.function || !route.capability) throw new Error("operation not declared in plugin manifest");
    return { accepted: true, operation: rpc.params.operation, function: route.function, capability: route.capability };
  };
}

export function createHostHandler(host: HostInvoker, routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = routes.find((candidate) => candidate.operation === rpc.params.operation);
    if (!route) throw new Error("operation not declared in plugin manifest");
    if (!route.function || !route.capability) throw new Error("invalid plugin route");
    return {
      accepted: true,
      operation: rpc.params.operation,
      payload: host.invoke(rpc.params.projectId, rpc.params.operation, route.capability, rpc.params.payload ?? null),
    };
  };
}

export type PluginFunction = (request: RpcRequest) => unknown;

export function createFunctionHandler(functions: Readonly<Record<string, PluginFunction>>, routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = routes.find((candidate) => candidate.operation === rpc.params.operation);
    if (!route) throw new Error("operation not declared in plugin manifest");
    if (!route.function || !route.capability) throw new Error("invalid plugin route");
    const fn = functions[route.function];
    if (!fn) throw new Error("route function not implemented");
    return { accepted: true, operation: rpc.params.operation, payload: fn(rpc) };
  };
}

export function request(id: string, params: WorkerParams): RpcRequest {
  if (!id || !params.projectId || !params.pluginId || !params.version || !params.operation) {
    throw new Error("invalid plugin request");
  }
  return { jsonrpc: "2.0", id, method: "invoke", params };
}

export function pluginRequest(prefix: string, projectId: string, pluginId: string, version: string, operation: string, payload: unknown = {}, identity = ""): RpcRequest {
  return request(`${prefix}-${operation}-${projectId}${identity ? `-${identity}` : ""}`, { protocol: 1, projectId, pluginId, version, operation, payload });
}

export function validateProjectId(projectId: string, errorMessage = "invalid project id"): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error(errorMessage);
  return projectId;
}

export function validateIdentifier(identifier: string, errorMessage = "invalid identifier"): string {
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(identifier)) throw new Error(errorMessage);
  return identifier;
}
