export function recordBrowserPage(
  projectId: string,
  table: string,
  rows: readonly Record<string, unknown>[],
  approvedColumns: readonly string[],
): {
  projectId: string;
  table: string;
  columns: readonly string[];
  rows: Record<string, unknown>[];
} {
  if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !approvedColumns.length)
    throw new Error("invalid record browser");
  return {
    projectId,
    table,
    columns: [...approvedColumns],
    rows: rows.map((row) =>
      Object.fromEntries(
        approvedColumns
          .filter((column) => Object.hasOwn(row, column))
          .map((column) => [column, row[column]]),
      ),
    ),
  };
}
