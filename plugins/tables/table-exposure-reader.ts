import type { TableExposure } from "./table-exposure.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export class TableExposureReader {
  constructor(private readonly records: ReadonlyMap<string, TableExposure>) {}
  get(projectId: string, table: string): TableExposure | undefined {
    validateProjectId(projectId);
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid table exposure");
    const record = this.records.get(`${projectId}:${table}`);
    return record ? { ...record, columns: [...record.columns] } : undefined;
  }
}
