import type { PluginRoute } from "./index.js";

export function requireRoute(routes: readonly PluginRoute[], operation: string): PluginRoute {
  const route = routes.find((candidate) => candidate.operation === operation);
  if (!route) throw new Error("operation not declared in plugin manifest");
  if (!route.function || !route.capability) throw new Error("invalid plugin route");
  return route;
}
