export function read(input: unknown): { table: string; rows: readonly unknown[] } {
  const payload = (input ?? {}) as { table?: unknown; rows?: unknown };
  const table = typeof payload.table === "string" ? payload.table.trim() : "";
  if (!table) throw new Error("table is required");
  return tableStore.execute("read", table) as { table: string; rows: readonly unknown[] };
}
import { tableStore } from "./table-store.js";
