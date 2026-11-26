import { createCommitRollback, validateProjectId } from "../sdk-typescript/src/index.js";
import type { RowChange, ScopedRowChange } from "./index.js";

export function createTransactionalChangeCapture<T>() {
  const changes: RowChange<T>[] = [];
  const scopedChanges: ScopedRowChange<T>[] = [];
  return {
    begin(projectId: string, table: string, oldValue: T, newValue: T) {
      validateProjectId(projectId, "invalid change target");
      validateTable(table);
      return createCommitRollback(() => { changes.push({ projectId, table, oldValue, newValue }); });
    },
    beginScoped(
      projectId: string,
      table: string,
      oldValue: T,
      newValue: T,
      oldScope: readonly string[],
      newScope: readonly string[],
      approvedScopes: readonly string[] = ["public", "member", "owner"],
    ) {
      validateProjectId(projectId, "invalid change target");
      const allowed = new Set(approvedScopes);
      if (![...oldScope, ...newScope].every((scope) => scope && allowed.has(scope)))
        throw new Error("unapproved authorization scope");
      validateTable(table);
      return createCommitRollback(() => {
        scopedChanges.push({ projectId, table, oldValue, newValue, oldScope: [...oldScope], newScope: [...newScope] });
      });
    },
    list(projectId: string) {
      return changes.filter((change) => change.projectId === projectId).map((change) => ({ ...change }));
    },
    listScoped(projectId: string) {
      validateProjectId(projectId);
      return scopedChanges.filter((change) => change.projectId === projectId).map((change) => ({
        ...change,
        oldScope: [...change.oldScope],
        newScope: [...change.newScope],
      }));
    },
  };
}

function validateTable(table: string): void {
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid change target");
}
