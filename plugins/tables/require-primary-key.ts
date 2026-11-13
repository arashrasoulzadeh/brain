export function requirePrimaryKey(columns: readonly string[], primaryKey: string): void {
  if (!primaryKey || !columns.includes(primaryKey)) throw new Error("supported primary key required");
}
