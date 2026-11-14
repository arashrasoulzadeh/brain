import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { TableExposure } from "./table-exposure.js";

export class TableExposureWriter {
  constructor(private readonly records: Map<string, TableExposure>) {}
  expose(projectId: string, table: string, columns: readonly string[]): void {
    validateProjectId(projectId);
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !columns.length)
      throw new Error("invalid table exposure");
    this.records.set(`${projectId}:${table}`, { projectId, table, columns: [...columns] });
  }
}
