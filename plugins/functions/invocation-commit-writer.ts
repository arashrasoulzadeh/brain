import type { InvocationCommitState } from "./invocation-commit-state.js";
export class InvocationCommitWriter {
  constructor(private readonly state: InvocationCommitState) {}
  commit(
    idempotencyKey: string,
    effects: readonly unknown[],
    result: unknown,
    success = true,
    expectedRevision = this.state.lifecycleRevision,
  ): unknown {
    if (!idempotencyKey) throw new Error("idempotency key required");
    const existing = this.state.records.get(idempotencyKey);
    if (existing !== undefined) return existing;
    if (expectedRevision !== this.state.lifecycleRevision)
      throw new Error("stale invocation lifecycle revision");
    if (!success) throw new Error("invocation commit failed");
    this.state.effects.push(...effects);
    this.state.records.set(idempotencyKey, result);
    return result;
  }
}
