import { validateProjectId } from "../sdk-typescript/src/index.js";

export function projectSchemaName(projectId: string): string {
  validateProjectId(projectId);
  return `project_${projectId}`;
}

export function projectRoleName(projectId: string): string {
  return `runtime_${projectSchemaName(projectId)}`;
}

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

export class RowPolicyChecker {
  constructor(private readonly grants: ReadonlySet<string>) {}
  allows(projectId: string, actorId: string): boolean {
    validateProjectId(projectId);
    validateProjectId(actorId, "invalid actor id");
    return this.grants.has(`${projectId}:${actorId}`);
  }
}

export class RowPolicyGranter {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, actorId: string): boolean {
    validateProjectId(projectId);
    validateProjectId(actorId, "invalid actor id");
    this.grants.add(`${projectId}:${actorId}`);
    return true;
  }
}

export type TableExposure = { projectId: string; table: string; columns: readonly string[] };

export class TableExposureReader {
  constructor(private readonly records: ReadonlyMap<string, TableExposure>) {}
  get(projectId: string, table: string): TableExposure | undefined {
    validateProjectId(projectId);
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid table exposure");
    const record = this.records.get(`${projectId}:${table}`);
    return record ? { ...record, columns: [...record.columns] } : undefined;
  }
}

export class TableExposureWriter {
  constructor(private readonly records: Map<string, TableExposure>) {}
  expose(projectId: string, table: string, columns: readonly string[]): void {
    validateProjectId(projectId);
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !columns.length)
      throw new Error("invalid table exposure");
    this.records.set(`${projectId}:${table}`, { projectId, table, columns: [...columns] });
  }
}
