import { validateIdentifier } from "../sdk-typescript/src/index.js";
import type { SchemaChange, SchemaChangePlan } from "./schema-types.js";
export function planSchemaChange(
  projectId: string,
  table: string,
  change: SchemaChange,
): SchemaChangePlan {
  if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table))
    throw new Error("invalid schema target");
  if (change.kind === "add_column") {
    validateIdentifier(change.name);
    return { projectId, table, change: { ...change } };
  }
  if (change.kind === "create_index") {
    validateIdentifier(change.name);
    if (!change.columns.length) throw new Error("index columns required");
    change.columns.forEach((columnName) => validateIdentifier(columnName));
    return { projectId, table, change: { ...change, columns: [...change.columns] } };
  }
  throw new Error("unsupported schema change");
}
