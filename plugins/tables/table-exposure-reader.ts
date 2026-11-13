import type { TableExposure } from "./table-exposure.js";

export class TableExposureReader {
  constructor(private readonly records: ReadonlyMap<string, TableExposure>) {}
  get(projectId: string, table: string): TableExposure | undefined { const record = this.records.get(`${projectId}:${table}`); return record ? { ...record, columns: [...record.columns] } : undefined; }
}
