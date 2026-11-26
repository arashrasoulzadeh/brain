import { createCommitRollback, validateIdentifier, validateProjectId } from "../sdk-typescript/src/index.js";
import type { DirectSQLChange } from "./index.js";

export function createDirectSQLChangeCapture() {
  const changes: DirectSQLChange[] = [];
  return {
    begin(projectId: string, table: string, statement: string) {
      validateIdentifier(table);
      validateProjectId(projectId, "invalid SQL change");
      if (!statement) throw new Error("invalid SQL change");
      return createCommitRollback(() => { changes.push({ projectId, table, statement }); });
    },
    list(projectId: string) {
      return changes.filter((change) => change.projectId === projectId).map((change) => ({ ...change }));
    },
  };
}
