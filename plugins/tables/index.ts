export { dataRequest, read, write, schema, dispatchRoute, TableStore, tableStore } from "./routes.js";
export {
  projectSchemaName, projectRoleName, RowPolicyGranter, RowPolicyChecker,
  TableExposureWriter, TableExposureReader, type TableExposure, projectRow,
} from "./access.js";
export {
  requirePrimaryKey, validateColumnTypes, type SchemaChange, type SchemaChangePlan,
  type SchemaPublication, planSchemaChange, previewSchemaSQL, requireRecentSchemaAuth,
  createSchemaPublicationAudit,
} from "./schema.js";
export {
  validateFilter, paginate, validateValue, createVersionedRecord, IdempotentCreateStore,
} from "./records.js";
export { validateIdentifier } from "../sdk-typescript/src/index.js";
export {
  createTransactionalChangeCapture, createApiChangeCapture,
  createDirectSQLChangeCapture,
} from "./capture.js";
export type RowChange<T> = { projectId: string; table: string; oldValue: T; newValue: T };
export type ScopedRowChange<T> = RowChange<T> & {
  oldScope: readonly string[];
  newScope: readonly string[];
};
export type ApiChange = { projectId: string; operation: string; payload: unknown };
export type DirectSQLChange = { projectId: string; table: string; statement: string };
