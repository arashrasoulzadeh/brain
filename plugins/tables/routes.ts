import { createResourceRequest, createRouteDispatcher, requiredString, requireProjectInput, requireOperator, type PluginContext } from "../sdk-typescript/src/index.js";

export const dataRequest = createResourceRequest<"tables.read" | "tables.write" | "tables.schema">("tables", "table");

type TableRecord = Record<string, unknown>;

export class TableStore {
  private readonly rows = new Map<string, TableRecord[]>();
  private readonly columns = new Map<string, string[]>();

  execute(operation: "read" | "write" | "schema", table: string, value?: TableRecord): unknown {
    if (operation === "write") {
      const rows = this.rows.get(table) ?? [];
      rows.push({ ...value });
      this.rows.set(table, rows);
      const columns = this.columns.get(table) ?? [];
      for (const column of Object.keys(value ?? {}))
        if (!columns.includes(column)) columns.push(column);
      this.columns.set(table, columns);
      return { table, row: value, written: true };
    }
    if (operation === "schema") return { table, columns: this.columns.get(table) ?? [] };
    return { table, rows: this.rows.get(table) ?? [] };
  }
}

export const tableStore = new TableStore();

type TablePage = { revision: string; total_rows: number; offset: number; next_offset?: number };
type TableInput = { table?: unknown; rows?: unknown; row?: unknown; page?: TablePage };
type TableRead = { table: string; rows: readonly unknown[] } & Partial<TablePage>;
type TableWrite = { table: string; row: unknown; written: true };
type TableSchema = { table: string; columns: readonly string[] };

export const dispatchRoute = createRouteDispatcher({ read, write, schema });

export function read(input: unknown): TableRead {
  const payload = (input ?? {}) as TableInput;
  const table = requiredString(payload.table, "table is required");
  if (Array.isArray(payload.rows)) return { table, rows: payload.rows, ...(payload.page ? { revision: payload.page.revision, total_rows: payload.page.total_rows, offset: payload.page.offset, ...(payload.page.next_offset !== undefined ? { next_offset: payload.page.next_offset } : {}) } : {}) };
  return tableStore.execute("read", table) as TableRead;
}

export function write(input: unknown, context?: PluginContext): TableWrite {
  const payload = (input ?? {}) as TableInput;
  const table = requiredString(payload.table, "table is required");
  if (payload.row === undefined || payload.row === null || typeof payload.row !== "object")
    throw new Error("row is required");
  const result = tableStore.execute("write", table, payload.row as Record<string, unknown>) as TableWrite;
  context?.plugins.observe.fire("record.written", { written: true });
  return result;
}

export function schema(input: unknown): TableSchema {
  const payload = (input ?? {}) as TableInput;
  const table = requiredString(payload.table, "table is required");
  if (Array.isArray(payload.rows)) {
    const columns = [...new Set(payload.rows.flatMap((row) => row && typeof row === "object" && !Array.isArray(row) ? Object.keys(row) : []))].sort();
    return { table, columns };
  }
  return tableStore.execute("schema", table) as TableSchema;
}

/** The host commits the revision-checked mutation only after this succeeds. */
export function mutateTableRow(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const table = requiredString(payload.table, "table is required");
  const result = payload.result as Record<string, unknown> | undefined;
  if (!result || typeof result !== "object" || result.table !== table || typeof result.revision !== "string" ||
      (context.execution.operation === "tables.update" ? result.updated !== true : context.execution.operation !== "tables.delete" || result.deleted !== true))
    throw new Error("host table mutation snapshot required");
  context.plugins.observe.fire("record.written", { written: true });
  return result;
}

export function administerSchema(context: PluginContext, input: unknown) {
  const payload = requireProjectInput(context, input);
  const table = requiredString(payload.table, "table is required");
  const read = context.execution.operation === "tables.schema.profile";
  requireOperator(context, read ? "tables.schema" : "tables.admin.schema", table);
  const result = payload.result as Record<string, unknown> | undefined;
  if (!result || typeof result !== "object" || result.table !== table || typeof result.revision !== "string" || typeof result.exists !== "boolean") throw new Error("host schema snapshot required");
  if (!read) context.plugins.observe.fire("record.written", { written: true });
  return result;
}
