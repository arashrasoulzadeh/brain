import type { TransactionChangeState } from "./transaction-change-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class ScopedTransactionChangeWriter<T> {
  constructor(private readonly state: TransactionChangeState<T>) {}
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
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)) throw new Error("invalid change target");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit)
        this.state.scopedChanges.push({
          projectId,
          table,
          oldValue,
          newValue,
          oldScope: [...oldScope],
          newScope: [...newScope],
        });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
}
