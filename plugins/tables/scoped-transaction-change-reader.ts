import type { TransactionChangeState } from "./transaction-change-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class ScopedTransactionChangeReader<T> {
  constructor(private readonly state: TransactionChangeState<T>) {}
  listScoped(projectId: string) {
    validateProjectId(projectId);
    return this.state.scopedChanges
      .filter((change) => change.projectId === projectId)
      .map((change) => ({
        ...change,
        oldScope: [...change.oldScope],
        newScope: [...change.newScope],
      }));
  }
}
