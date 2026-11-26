import type { HostExecutionContext, HostInvoker, PluginFunction, PluginHandler, PluginRoute, RpcRequest } from "./index.js";

export type PayloadHandler = (input: unknown) => unknown;

export function createRouteDispatcher(handlers: Readonly<Record<string, PayloadHandler>>) {
  return createDispatcher<unknown>(handlers);
}

export function createRequestDispatcher(
  handlers: Readonly<Record<string, PluginFunction>>,
  missingHandlerError?: string,
) {
  return createDispatcher<RpcRequest>(handlers, missingHandlerError);
}

function createDispatcher<Input>(
  handlers: Readonly<Record<string, (input: Input) => unknown>>,
  missingHandlerError?: string,
) {
  const registered = Object.freeze({ ...handlers });
  return (name: string, input: Input): unknown => {
    if (!Object.hasOwn(registered, name) || typeof registered[name] !== "function")
      throw new Error(missingHandlerError ?? "declared plugin function is not exported: " + name);
    return registered[name](input);
  };
}

export function createOperationRoutes<Operation extends string>(
  operations: readonly Operation[],
  execute: (operation: Operation, input: unknown) => unknown,
) {
  const handlers = Object.freeze(
    Object.fromEntries(
      operations.map((operation) => [operation, (input: unknown) => execute(operation, input)]),
    ) as Record<Operation, PayloadHandler>,
  );
  return { handlers, dispatch: createRouteDispatcher(handlers) };
}

function resolveRoute(rpc: RpcRequest, routes: readonly PluginRoute[], context?: HostExecutionContext): PluginRoute {
  if (context && (!context.projectId || !context.invocationId || rpc.params.projectId !== context.projectId))
    throw new Error("host project context mismatch");
  const route = routes.find((candidate) => candidate.operation === rpc.params.operation);
  if (!route) throw new Error("operation not declared in plugin manifest");
  if (!route.function || !route.capability)
    throw new Error(context ? "invalid plugin route" : "operation not declared in plugin manifest");
  return route;
}

export function createRouteHandler(routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = resolveRoute(rpc, routes);
    return { accepted: true, operation: rpc.params.operation, function: route.function, capability: route.capability };
  };
}

export function createHostHandler(host: HostInvoker, routes: readonly PluginRoute[], context: HostExecutionContext): PluginHandler {
  return (rpc) => {
    const route = resolveRoute(rpc, routes, context);
    return {
      accepted: true,
      operation: rpc.params.operation,
      payload: host.invoke(rpc.params.projectId, rpc.params.operation, route.capability, rpc.params.payload ?? null),
    };
  };
}

export function createFunctionHandler(functions: Readonly<Record<string, PluginFunction>>, routes: readonly PluginRoute[], context: HostExecutionContext): PluginHandler {
  const dispatch = createRequestDispatcher(functions, "route function not implemented");
  return (rpc) => {
    const route = resolveRoute(rpc, routes, context);
    return { accepted: true, operation: rpc.params.operation, payload: dispatch(route.function, rpc) };
  };
}
