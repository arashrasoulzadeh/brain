import { validateIdentifier as validateSdkIdentifier, validateProjectId } from "../sdk-typescript/src/index.js";
import { projectSchemaName } from "./project-schema-name.js";
import { createSchemaAuditState } from "./schema-audit-state.js";
import { SchemaPublicationRecorder } from "./schema-publication-recorder.js";
import { SchemaPublicationReader } from "./schema-publication-reader.js";
import { createApiChangeState } from "./api-change-state.js";
import { ApiChangeWriter } from "./api-change-writer.js";
import { ApiChangeReader } from "./api-change-reader.js";
import { createSQLChangeState } from "./sql-change-state.js";
import { SQLChangeWriter } from "./sql-change-writer.js";
import { SQLChangeReader } from "./sql-change-reader.js";

export { dataRequest } from "./data-request.js";
export { projectSchemaName } from "./project-schema-name.js";
export { projectRoleName } from "./project-role-name.js";
export { RowPolicyGranter } from "./row-policy-granter.js";
export { RowPolicyChecker } from "./row-policy-checker.js";
export { TableExposureWriter } from "./table-exposure-writer.js";
export { TableExposureReader } from "./table-exposure-reader.js";
export type { TableExposure } from "./table-exposure.js";
export { requirePrimaryKey } from "./require-primary-key.js";
export { validateColumnTypes } from "./validate-column-types.js";
export { projectRow } from "./project-row.js";
export { validateFilter } from "./validate-filter.js";
export { paginate } from "./paginate.js";
export const validateIdentifier = validateSdkIdentifier;
export { validateValue } from "./validate-value.js";
export { createVersionedRecord } from "./versioned-record.js";
export { IdempotentCreateStore } from "./idempotent-create-store.js";
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
    change.columns.forEach((columnName) => validateIdentifier(columnName));
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

export function requireRecentSchemaAuth(lastAuthenticatedAt: number, now: number, maxAgeMs = 5 * 60 * 1000): void {
  if (!Number.isFinite(lastAuthenticatedAt) || !Number.isFinite(now) || now < lastAuthenticatedAt || now - lastAuthenticatedAt > maxAgeMs) throw new Error("recent authentication required");
}

export type SchemaPublication = SchemaChangePlan & { actorId: string; publishedAt: number };
export function createSchemaPublicationAudit() { const state = createSchemaAuditState(); const recorder = new SchemaPublicationRecorder(state); const reader = new SchemaPublicationReader(state); return { record: recorder.record.bind(recorder), list: reader.list.bind(reader) }; }

export type RowChange<T> = { projectId: string; table: string; oldValue: T; newValue: T };
export type ScopedRowChange<T> = RowChange<T> & { oldScope: readonly string[]; newScope: readonly string[] };

export class TransactionalChangeCapture<T> {
  private readonly changes: RowChange<T>[] = [];
  private readonly scopedChanges: ScopedRowChange<T>[] = [];
  begin(projectId: string, table: string, oldValue: T, newValue: T): { commit: () => void; rollback: () => void } {
    if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid change target");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit) this.changes.push({ projectId, table, oldValue, newValue });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
  beginScoped(projectId: string, table: string, oldValue: T, newValue: T, oldScope: readonly string[], newScope: readonly string[], approvedScopes: readonly string[] = ["public", "member", "owner"]): { commit: () => void; rollback: () => void } {
    const allowed = new Set(approvedScopes);
    if (![...oldScope, ...newScope].every((scope) => scope && allowed.has(scope))) throw new Error("unapproved authorization scope");
    if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid change target");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit) this.scopedChanges.push({ projectId, table, oldValue, newValue, oldScope: [...oldScope], newScope: [...newScope] });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
  list(projectId: string): RowChange<T>[] { return this.changes.filter((change) => change.projectId === projectId).map((change) => ({ ...change })); }
  listScoped(projectId: string): ScopedRowChange<T>[] { return this.scopedChanges.filter((change) => change.projectId === projectId).map((change) => ({ ...change, oldScope: [...change.oldScope], newScope: [...change.newScope] })); }
}

export type ApiChange = { projectId: string; operation: string; payload: unknown };

export function createApiChangeCapture() { const state = createApiChangeState(); const writer = new ApiChangeWriter(state); const reader = new ApiChangeReader(state); return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) }; }

export type DirectSQLChange = { projectId: string; table: string; statement: string };

export function createDirectSQLChangeCapture() { const state = createSQLChangeState(); const writer = new SQLChangeWriter(state); const reader = new SQLChangeReader(state); return { begin: writer.begin.bind(writer), list: reader.list.bind(reader) }; }
