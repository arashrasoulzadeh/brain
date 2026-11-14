import { pluginRequest } from "../sdk-typescript/src/index.js";

export function inviteRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "invites.create" | "invites.accept" | "invites.cancel",
  roomId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "invites",
    projectId,
    pluginId,
    version,
    operation,
    { roomId, ...(payload as Record<string, unknown>) },
    roomId,
  );
}
