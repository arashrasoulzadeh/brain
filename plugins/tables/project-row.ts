export function projectRow(
  row: Record<string, unknown>,
  approvedColumns: readonly string[],
): Record<string, unknown> {
  return Object.fromEntries(
    approvedColumns
      .filter((column) => Object.prototype.hasOwnProperty.call(row, column))
      .map((column) => [column, row[column]]),
  );
}
