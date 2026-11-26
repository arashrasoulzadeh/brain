import { createRouteDispatcher, requiredString } from "../sdk-typescript/src/index.js";
import { tableStore } from "./table-store.js";

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
