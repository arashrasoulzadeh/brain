import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function subscriptionRequest(projectId: string, pluginId: string, version: string, operation: "subscriptions.open" | "subscriptions.receive" | "subscriptions.resync", channel: string, payload: unknown = {}) {
  return request(`subscriptions-${operation}-${projectId}-${channel}`, { protocol: 1, projectId, pluginId, version, operation, payload: { channel, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
