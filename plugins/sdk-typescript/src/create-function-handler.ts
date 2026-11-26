import type { HostExecutionContext, PluginFunction, PluginHandler, PluginRoute } from "./index.js";
import { requireHostContext } from "./require-host-context.js";
import { requireRoute } from "./require-route.js";

export function createFunctionHandler(
  functions: Readonly<Record<string, PluginFunction>>,
  routes: readonly PluginRoute[],
  context: HostExecutionContext,
): PluginHandler {
  return (rpc) => {
    requireHostContext(context, rpc);
    const route = requireRoute(routes, rpc.params.operation);
    const fn = functions[route.function];
    if (!Object.prototype.hasOwnProperty.call(functions, route.function) || typeof fn !== "function")
      throw new Error("route function not implemented");
    return { accepted: true, operation: rpc.params.operation, payload: fn(rpc) };
  };
}
