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
export function paginate<T>(items: readonly T[], cursor: string | undefined, limit: number): { items: T[]; nextCursor?: string } {
  if (!Number.isInteger(limit) || limit < 1 || limit > 100) throw new Error("invalid page limit");
  const start = cursor ? Number(Buffer.from(cursor, "base64url").toString("utf8")) : 0;
  if (!Number.isInteger(start) || start < 0 || start > items.length) throw new Error("invalid page cursor");
  const page = [...items.slice(start, start + limit)];
  return { items: page, ...(start + page.length < items.length ? { nextCursor: Buffer.from(String(start + page.length)).toString("base64url") } : {}) };
}
export function validateIdentifier(value: string): void { if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(value)) throw new Error("invalid identifier"); }
export function validateValue(column: string, value: unknown): void { validateIdentifier(column); if (value === undefined || typeof value === "function" || (typeof value === "string" && value.length > 4096)) throw new Error("invalid value"); }
export class VersionedRecord<T extends Record<string, unknown>> {
  private value: T;
  version = 0;
  constructor(value: T) { this.value = { ...value }; }
  update(expectedVersion: number, next: T): void { if (expectedVersion !== this.version) throw new Error("stale record version"); this.value = { ...next }; this.version += 1; }
  read(): { value: T; version: number } { return { value: { ...this.value }, version: this.version }; }
}
export class IdempotentCreateStore<T> {
  private readonly results = new Map<string, T>();
  create(projectId: string, key: string, factory: () => T): T {
    if (!projectId || !key) throw new Error("idempotency key required");
    const id = `${projectId}:${key}`;
    const existing = this.results.get(id);
    if (existing) return existing;
    const value = factory(); this.results.set(id, value); return value;
  }
}

export type SchemaChange =
  | { kind: "add_column"; name: string; type: "string" | "number" | "boolean" | "date" }
  | { kind: "create_index"; name: string; columns: readonly string[] };

export type SchemaChangePlan = { projectId: string; table: string; change: SchemaChange };

export function planSchemaChange(projectId: string, table: string, change: SchemaChange): SchemaChangePlan {
  if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid schema target");
  if (change.kind === "add_column") {
    validateIdentifier(change.name);
    return { projectId, table, change: { ...change } };
  }
  if (change.kind === "create_index") {
    validateIdentifier(change.name);
    if (!change.columns.length) throw new Error("index columns required");
    change.columns.forEach(validateIdentifier);
    return { projectId, table, change: { ...change, columns: [...change.columns] } };
  }
  throw new Error("unsupported schema change");
}

export function previewSchemaSQL(plan: SchemaChangePlan): string {
  const checked = planSchemaChange(plan.projectId, plan.table, plan.change);
  const schema = projectSchemaName(checked.projectId);
  if (checked.change.kind === "add_column") {
    const type = { string: "TEXT", number: "DOUBLE PRECISION", boolean: "BOOLEAN", date: "TIMESTAMPTZ" }[checked.change.type];
    return `ALTER TABLE "${schema}"."${checked.table}" ADD COLUMN "${checked.change.name}" ${type};`;
  }
  return `CREATE INDEX "${checked.change.name}" ON "${schema}"."${checked.table}" (${checked.change.columns.map((column) => `"${column}"`).join(", ")});`;
}

export function dataRequest(projectId: string, pluginId: string, version: string, operation: "tables.read" | "tables.write" | "tables.schema", table: string, payload: unknown = {}) {
  return request(`tables-${operation}-${projectId}-${table}`, { protocol: 1, projectId, pluginId, version, operation, payload: { table, ...(payload as Record<string, unknown>) } });
}

