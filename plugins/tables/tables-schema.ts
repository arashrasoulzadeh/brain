export function schema(input: unknown): { table: string; columns: readonly string[] } {
  const payload = (input ?? {}) as { table?: unknown; columns?: unknown };
  const table = typeof payload.table === "string" ? payload.table.trim() : "";
  if (!table) throw new Error("table is required");
  return tableStore.execute("schema", table) as { table: string; columns: readonly string[] };
}
import { tableStore } from "./table-store.js";
