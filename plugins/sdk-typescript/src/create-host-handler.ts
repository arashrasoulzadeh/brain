import type { HostExecutionContext, HostInvoker, PluginHandler, PluginRoute } from "./index.js";
import { requireHostContext } from "./require-host-context.js";
import { requireRoute } from "./require-route.js";

export function createHostHandler(
  host: HostInvoker,
  routes: readonly PluginRoute[],
  context: HostExecutionContext,
): PluginHandler {
  return (rpc) => {
    requireHostContext(context, rpc);
    const route = requireRoute(routes, rpc.params.operation);
    return {
      accepted: true,
      operation: rpc.params.operation,
      payload: host.invoke(
        rpc.params.projectId,
        rpc.params.operation,
        route.capability,
        rpc.params.payload ?? null,
      ),
    };
  };
}
