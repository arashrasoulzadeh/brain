import { request, createRouteHandler, type PluginHandler } from "../sdk-typescript/src/index.js";

export function dataRequest(projectId: string, pluginId: string, version: string, operation: "tables.read" | "tables.write" | "tables.schema", table: string, payload: unknown = {}) {
  return request(`tables-${operation}-${projectId}-${table}`, { protocol: 1, projectId, pluginId, version, operation, payload: { table, ...(payload as Record<string, unknown>) } });
}

export const createHandler = createRouteHandler;
export const handle: PluginHandler = createHandler([]);
