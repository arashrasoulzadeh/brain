import type { TransactionChangeState } from "./transaction-change-state.js";
export class TransactionChangeReader<T> {
  constructor(private readonly state: TransactionChangeState<T>) {}
  list(projectId: string) {
    return this.state.changes
      .filter((change) => change.projectId === projectId)
      .map((change) => ({ ...change }));
  }
}
