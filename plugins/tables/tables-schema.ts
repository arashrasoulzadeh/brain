export function schema(input: unknown): { table: string; columns: readonly string[] } {
  const payload = (input ?? {}) as { table?: unknown; columns?: unknown };
  const table = typeof payload.table === "string" ? payload.table.trim() : "";
  if (!table) throw new Error("table is required");
  if (
    payload.columns !== undefined &&
    (!Array.isArray(payload.columns) ||
      payload.columns.some((column) => typeof column !== "string"))
  )
    throw new Error("columns must be strings");
  return { table, columns: (payload.columns as readonly string[] | undefined) ?? [] };
}
