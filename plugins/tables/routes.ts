import { createManifestRequest, createRouteDispatcher, requiredString } from "../sdk-typescript/src/index.js";

export const dataRequest = createManifestRequest("tables", "table");

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

type TableInput = { table?: unknown; rows?: unknown; row?: unknown };
type TableRead = { table: string; rows: readonly unknown[] };
type TableWrite = { table: string; row: unknown; written: true };
type TableSchema = { table: string; columns: readonly string[] };

export const dispatchRoute = createRouteDispatcher({ read, write, schema });

export function read(input: unknown): TableRead {
  const payload = (input ?? {}) as TableInput;
  const table = requiredString(payload.table, "table is required");
  if (Array.isArray(payload.rows)) return { table, rows: payload.rows };
  return tableStore.execute("read", table) as TableRead;
}

export function write(input: unknown): TableWrite {
  const payload = (input ?? {}) as TableInput;
  const table = requiredString(payload.table, "table is required");
  if (payload.row === undefined || payload.row === null || typeof payload.row !== "object")
    throw new Error("row is required");
  return tableStore.execute("write", table, payload.row as Record<string, unknown>) as TableWrite;
}

export function schema(input: unknown): TableSchema {
  const payload = (input ?? {}) as TableInput;
  return tableStore.execute("schema", requiredString(payload.table, "table is required")) as TableSchema;
}
