type TableRecord = Record<string, unknown>;

export class TableStore {
  private readonly rows = new Map<string, TableRecord[]>();
  private readonly columns = new Map<string, string[]>();

  execute(operation: "read" | "write" | "schema", table: string, value?: TableRecord): unknown {
    if (operation === "write") {
      const rows = this.rows.get(table) ?? [];
      rows.push({ ...value });
      this.rows.set(table, rows);
      const columns = this.columns.get(table) ?? [];
      for (const column of Object.keys(value ?? {}))
        if (!columns.includes(column)) columns.push(column);
      this.columns.set(table, columns);
      return { table, row: value, written: true };
    }
    if (operation === "schema") return { table, columns: this.columns.get(table) ?? [] };
    return { table, rows: this.rows.get(table) ?? [] };
  }
}

export const tableStore = new TableStore();
