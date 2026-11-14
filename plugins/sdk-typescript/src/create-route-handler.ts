import type { PluginHandler, PluginRoute } from "./index.js";
export function createRouteHandler(routes: readonly PluginRoute[]): PluginHandler {
  return (rpc) => {
    const route = routes.find((candidate) => candidate.operation === rpc.params.operation);
    if (!route || !route.function || !route.capability)
      throw new Error("operation not declared in plugin manifest");
    return {
      accepted: true,
      operation: rpc.params.operation,
      function: route.function,
      capability: route.capability,
    };
  };
}
