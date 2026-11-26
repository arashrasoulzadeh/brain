import { validateIdentifier, validateProjectId } from "../sdk-typescript/src/index.js";
import { projectSchemaName } from "./access.js";

export type SchemaChange =
  | { kind: "add_column"; name: string; type: "string" | "number" | "boolean" | "date" }
  | { kind: "create_index"; name: string; columns: readonly string[] };
export type SchemaChangePlan = { projectId: string; table: string; change: SchemaChange };
export type SchemaPublication = SchemaChangePlan & { actorId: string; publishedAt: number };

export function planSchemaChange(
  projectId: string,
  table: string,
  change: SchemaChange,
): SchemaChangePlan {
  validateProjectId(projectId, "invalid schema target");
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid schema target");
  if (change.kind === "add_column") {
    validateIdentifier(change.name);
    return { projectId, table, change: { ...change } };
  }
  if (change.kind === "create_index") {
    validateIdentifier(change.name);
    if (!change.columns.length) throw new Error("index columns required");
    change.columns.forEach((columnName) => validateIdentifier(columnName));
    return { projectId, table, change: { ...change, columns: [...change.columns] } };
  }
  throw new Error("unsupported schema change");
}

export function previewSchemaSQL(plan: SchemaChangePlan): string {
  const checked = planSchemaChange(plan.projectId, plan.table, plan.change);
  const schema = projectSchemaName(checked.projectId);
  if (checked.change.kind === "add_column") {
    const type = {
      string: "TEXT",
      number: "DOUBLE PRECISION",
      boolean: "BOOLEAN",
      date: "TIMESTAMPTZ",
    }[checked.change.type];
    return `ALTER TABLE "${schema}"."${checked.table}" ADD COLUMN "${checked.change.name}" ${type};`;
  }
  return `CREATE INDEX "${checked.change.name}" ON "${schema}"."${checked.table}" (${checked.change.columns.map((column) => `"${column}"`).join(", ")});`;
}

export function requireRecentSchemaAuth(
  lastAuthenticatedAt: number,
  now: number,
  maxAgeMs = 5 * 60 * 1000,
): void {
  if (
    !Number.isFinite(lastAuthenticatedAt) ||
    !Number.isFinite(now) ||
    now < lastAuthenticatedAt ||
    now - lastAuthenticatedAt > maxAgeMs
  )
    throw new Error("recent authentication required");
}

export function createSchemaPublicationAudit() {
  const entries: SchemaPublication[] = [];
  return {
    record(projectId: string, actorId: string, plan: SchemaChangePlan, publishedAt: number): void {
      if (!actorId || plan.projectId !== projectId || !Number.isFinite(publishedAt))
        throw new Error("invalid schema publication audit");
      entries.push({ ...plan, actorId, publishedAt, change: copyChange(plan.change) });
    },
    list(projectId: string) {
      return entries.filter((entry) => entry.projectId === projectId).map((entry) => ({
        ...entry,
        change: copyChange(entry.change),
      }));
    },
  };
}

function copyChange(change: SchemaChange): SchemaChange {
  return change.kind === "create_index" ? { ...change, columns: [...change.columns] } : { ...change };
}

export function requirePrimaryKey(columns: readonly string[], primaryKey: string): void {
  if (!primaryKey || !columns.includes(primaryKey))
    throw new Error("supported primary key required");
}

export function validateColumnTypes(columns: Record<string, string>): void {
  const supported = new Set(["string", "number", "boolean", "date"]);
  if (Object.values(columns).some((type) => !supported.has(type)))
    throw new Error("unsupported column type");
}
