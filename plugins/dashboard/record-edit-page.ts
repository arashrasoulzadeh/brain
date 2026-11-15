import { validateProjectId } from "../sdk-typescript/src/index.js";

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
    !recordId ||
    !Number.isInteger(expectedVersion) ||
    expectedVersion < 0 ||
    !approvedColumns.length
  )
    throw new Error("invalid record edit");
  try {
    validateProjectId(projectId);
  } catch {
    throw new Error("invalid record edit");
  }
  const filtered = Object.fromEntries(
    Object.entries(patch).filter(([column]) => approvedColumns.includes(column)),
  );
  return { projectId, table, recordId, expectedVersion, patch: filtered };
}
