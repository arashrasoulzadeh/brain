import { projectSchemaName } from "./project-schema-name.js";
import { planSchemaChange } from "./plan-schema-change.js";
import type { SchemaChangePlan } from "./schema-types.js";
export function previewSchemaSQL(plan: SchemaChangePlan): string {
  const checked = planSchemaChange(plan.projectId, plan.table, plan.change);
  const schema = projectSchemaName(checked.projectId);
  if (checked.change.kind === "add_column") {
    const type = {
      string: "TEXT",
      number: "DOUBLE PRECISION",
      boolean: "BOOLEAN",
      date: "TIMESTAMPTZ",
    }[checked.change.type];
    return `ALTER TABLE "${schema}"."${checked.table}" ADD COLUMN "${checked.change.name}" ${type};`;
  }
  return `CREATE INDEX "${checked.change.name}" ON "${schema}"."${checked.table}" (${checked.change.columns.map((column) => `"${column}"`).join(", ")});`;
}
