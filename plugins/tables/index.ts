import { request } from "../sdk-typescript/src/index.js";

export function projectSchemaName(projectId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error("invalid project id");
  return `project_${projectId}`;
}
export function projectRoleName(projectId: string): string { return `runtime_${projectSchemaName(projectId)}`; }
export class RowPolicy {
  private readonly grants = new Set<string>();
  grant(projectId: string, actorId: string): boolean { if (!projectId || !actorId) return false; this.grants.add(`${projectId}:${actorId}`); return true; }
  allows(projectId: string, actorId: string): boolean { return this.grants.has(`${projectId}:${actorId}`); }
}
export type TableExposure = { projectId: string; table: string; columns: readonly string[] };
export class TableExposureRegistry {
  private readonly records = new Map<string, TableExposure>();
  expose(projectId: string, table: string, columns: readonly string[]): void {
    if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !columns.length) throw new Error("invalid table exposure");
    this.records.set(`${projectId}:${table}`, { projectId, table, columns: [...columns] });
  }
  get(projectId: string, table: string): TableExposure | undefined { const record = this.records.get(`${projectId}:${table}`); return record ? { ...record, columns: [...record.columns] } : undefined; }
}
export function requirePrimaryKey(columns: readonly string[], primaryKey: string): void {
  if (!primaryKey || !columns.includes(primaryKey)) throw new Error("supported primary key required");
}
export function validateColumnTypes(columns: Record<string, string>): void {
  const supported = new Set(["string", "number", "boolean", "date"]);
  if (Object.values(columns).some((type) => !supported.has(type))) throw new Error("unsupported column type");
}
export function projectRow(row: Record<string, unknown>, approvedColumns: readonly string[]): Record<string, unknown> {
  return Object.fromEntries(approvedColumns.filter((column) => Object.prototype.hasOwnProperty.call(row, column)).map((column) => [column, row[column]]));
}
export function validateFilter(filter: Record<string, unknown>, approvedColumns: readonly string[]): void {
  for (const [column, expression] of Object.entries(filter)) {
    if (!approvedColumns.includes(column) || typeof expression !== "object" || expression === null || !Object.keys(expression as object).every((operator) => ["eq", "lt", "gt", "in"].includes(operator))) throw new Error("unapproved filter");
  }
}

export function dataRequest(projectId: string, pluginId: string, version: string, operation: "tables.read" | "tables.write" | "tables.schema", table: string, payload: unknown = {}) {
  return request(`tables-${operation}-${projectId}-${table}`, { protocol: 1, projectId, pluginId, version, operation, payload: { table, ...(payload as Record<string, unknown>) } });
}

