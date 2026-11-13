export function validateFilter(filter: Record<string, unknown>, approvedColumns: readonly string[]): void {
  for (const [column, expression] of Object.entries(filter)) {
    if (!approvedColumns.includes(column) || typeof expression !== "object" || expression === null || !Object.keys(expression as object).every((operator) => ["eq", "lt", "gt", "in"].includes(operator))) throw new Error("unapproved filter");
  }
}
