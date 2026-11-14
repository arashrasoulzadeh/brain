import { pluginRequest } from "../sdk-typescript/src/index.js";

export function subscriptionRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "subscriptions.open" | "subscriptions.receive" | "subscriptions.resync",
  channel: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "subscriptions",
    projectId,
    pluginId,
    version,
    operation,
    { channel, ...(payload as Record<string, unknown>) },
    channel,
  );
}
