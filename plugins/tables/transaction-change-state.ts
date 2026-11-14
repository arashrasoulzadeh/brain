import type { RowChange, ScopedRowChange } from "./index.js";
export type TransactionChangeState<T> = {
  changes: RowChange<T>[];
  scopedChanges: ScopedRowChange<T>[];
};
export function createTransactionChangeState<T>(): TransactionChangeState<T> {
  return { changes: [], scopedChanges: [] };
}
