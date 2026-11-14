import { pluginRequest } from "../sdk-typescript/src/index.js";

export function arenaRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "arena.start" | "arena.input" | "arena.snapshot" | "arena.result",
  roomId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "arena",
    projectId,
    pluginId,
    version,
    operation,
    { roomId, ...(payload as Record<string, unknown>) },
    roomId,
  );
}
