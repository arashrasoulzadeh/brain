import { createSchemaAuditState } from "./schema-audit-state.js";
import { SchemaPublicationRecorder } from "./schema-publication-recorder.js";
import { SchemaPublicationReader } from "./schema-publication-reader.js";
export function createSchemaPublicationAudit() { const state = createSchemaAuditState(); const recorder = new SchemaPublicationRecorder(state); const reader = new SchemaPublicationReader(state); return { record: recorder.record.bind(recorder), list: reader.list.bind(reader) }; }
