import type { InvocationCommitState } from "./invocation-commit-state.js";
export class InvocationResultReader {
  constructor(private readonly state: InvocationCommitState) {}
  result(idempotencyKey: string): unknown {
    return this.state.records.get(idempotencyKey);
  }
}
