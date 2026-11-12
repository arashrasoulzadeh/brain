import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "subscriptions",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["subscriptions.open", "subscriptions.receive", "subscriptions.resync"],
  extensionPoints: ["realtime.transport"],
  configSchema: { maxSubscriptions: "int" },
} as const;

export function subscriptionRequest(projectId: string, operation: "subscriptions.open" | "subscriptions.receive" | "subscriptions.resync", channel: string, payload: unknown = {}) {
  return request(`subscriptions-${operation}-${projectId}-${channel}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload: { channel, ...(payload as Record<string, unknown>) } });
}

export const handle: PluginHandler = (rpc) => {
  if (!rpc.params.operation.startsWith("subscriptions.")) throw new Error("unsupported subscriptions operation");
  return { accepted: true, operation: rpc.params.operation };
};
