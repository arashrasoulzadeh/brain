export function write(input: unknown): { table: string; row: unknown; written: true } {
  const payload = (input ?? {}) as { table?: unknown; row?: unknown };
  const table = typeof payload.table === "string" ? payload.table.trim() : "";
  if (!table) throw new Error("table is required");
  if (payload.row === undefined || payload.row === null || typeof payload.row !== "object")
    throw new Error("row is required");
  return tableStore.execute("write", table, payload.row as Record<string, unknown>) as {
    table: string;
    row: unknown;
    written: true;
  };
}
import { tableStore } from "./table-store.js";
