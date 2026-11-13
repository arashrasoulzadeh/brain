import { validateProjectId } from "../sdk-typescript/src/index.js";

export type TableExposure = { projectId: string; table: string; columns: readonly string[] };
export class TableExposureRegistry {
  private readonly records = new Map<string, TableExposure>();
  expose(projectId: string, table: string, columns: readonly string[]): void {
    validateProjectId(projectId);
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !columns.length) throw new Error("invalid table exposure");
    this.records.set(`${projectId}:${table}`, { projectId, table, columns: [...columns] });
  }
  get(projectId: string, table: string): TableExposure | undefined { const record = this.records.get(`${projectId}:${table}`); return record ? { ...record, columns: [...record.columns] } : undefined; }
}
