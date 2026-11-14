import { pluginRequest } from "../sdk-typescript/src/index.js";

export function fileRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "files.upload" | "files.download" | "files.metadata" | "files.publish",
  objectId: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "files",
    projectId,
    pluginId,
    version,
    operation,
    { objectId, ...(payload as Record<string, unknown>) },
    objectId,
  );
}
