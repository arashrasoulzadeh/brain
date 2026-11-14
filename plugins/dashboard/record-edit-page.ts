export function recordEditPage(
  projectId: string,
  table: string,
  recordId: string,
  expectedVersion: number,
  patch: Record<string, unknown>,
  approvedColumns: readonly string[],
): {
  projectId: string;
  table: string;
  recordId: string;
  expectedVersion: number;
  patch: Record<string, unknown>;
} {
  if (
    !projectId ||
    !recordId ||
    !Number.isInteger(expectedVersion) ||
    expectedVersion < 0 ||
    !approvedColumns.length
  )
    throw new Error("invalid record edit");
  const filtered = Object.fromEntries(
    Object.entries(patch).filter(([column]) => approvedColumns.includes(column)),
  );
  return { projectId, table, recordId, expectedVersion, patch: filtered };
}
