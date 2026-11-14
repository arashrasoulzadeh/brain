import { pluginRequest } from "../sdk-typescript/src/index.js";

export function dataRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: "tables.read" | "tables.write" | "tables.schema",
  table: string,
  payload: unknown = {},
) {
  return pluginRequest(
    "tables",
    projectId,
    pluginId,
    version,
    operation,
    { table, ...(payload as Record<string, unknown>) },
    table,
  );
}
