import { pluginRequest, type RpcRequest } from "../sdk-typescript/src/index.js";

export function functionRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "functions.invoke" | "functions.publish",
  payload: unknown = {},
): RpcRequest {
  return pluginRequest("functions", projectId, pluginId, version, operation, payload);
}
