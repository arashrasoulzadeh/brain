import { request } from "../sdk-typescript/src/index.js";

export function projectSchemaName(projectId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error("invalid project id");
  return `project_${projectId}`;
}

export function dataRequest(projectId: string, pluginId: string, version: string, operation: "tables.read" | "tables.write" | "tables.schema", table: string, payload: unknown = {}) {
  return request(`tables-${operation}-${projectId}-${table}`, { protocol: 1, projectId, pluginId, version, operation, payload: { table, ...(payload as Record<string, unknown>) } });
}

