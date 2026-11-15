export function read(input: unknown): { table: string; rows: readonly unknown[] } {
  const payload = (input ?? {}) as { table?: unknown; rows?: unknown };
  const table = typeof payload.table === "string" ? payload.table.trim() : "";
  if (!table) throw new Error("table is required");
  if (payload.rows !== undefined && !Array.isArray(payload.rows))
    throw new Error("rows must be an array");
  return { table, rows: (payload.rows as readonly unknown[] | undefined) ?? [] };
}
