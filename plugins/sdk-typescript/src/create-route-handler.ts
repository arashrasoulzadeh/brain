import type { PluginHandler, PluginRoute } from "./index.js";
import { requireRoute } from "./require-route.js";

export function createRouteHandler(routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = requireRoute(routes, rpc.params.operation);
    return {
      accepted: true,
      operation: rpc.params.operation,
      function: route.function,
      capability: route.capability,
    };
  };
}
