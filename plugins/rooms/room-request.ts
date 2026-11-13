import { pluginRequest } from "../sdk-typescript/src/index.js";

export function roomRequest(projectId: string, pluginId: string, version: string, operation: "rooms.create" | "rooms.join" | "rooms.reconnect" | "rooms.result", roomId: string, payload: unknown = {}) {
  return pluginRequest("rooms", projectId, pluginId, version, operation, { roomId, ...(payload as Record<string, unknown>) }, roomId);
}
