import type { TransactionChangeState } from "./transaction-change-state.js";
export class TransactionChangeWriter<T> {
  constructor(private readonly state: TransactionChangeState<T>) {}
  begin(projectId: string, table: string, oldValue: T, newValue: T) {
    if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table))
      throw new Error("invalid change target");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit) this.state.changes.push({ projectId, table, oldValue, newValue });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
}
