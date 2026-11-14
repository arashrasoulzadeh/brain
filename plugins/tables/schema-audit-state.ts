import type { SchemaPublication } from "./index.js";
export type SchemaAuditState = { entries: SchemaPublication[] };
export function createSchemaAuditState(): SchemaAuditState {
  return { entries: [] };
}
