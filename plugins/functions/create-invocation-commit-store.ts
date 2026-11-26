export function createInvocationCommitStore() {
  const records = new Map<string, unknown>();
  const committedEffects: unknown[] = [];
  let lifecycleRevision = 0;
  return {
    commit(
      idempotencyKey: string,
      effects: readonly unknown[],
      result: unknown,
      success = true,
      expectedRevision = lifecycleRevision,
    ): unknown {
      if (!idempotencyKey) throw new Error("idempotency key required");
      if (records.has(idempotencyKey)) return records.get(idempotencyKey);
      if (expectedRevision !== lifecycleRevision)
        throw new Error("stale invocation lifecycle revision");
      if (!success) throw new Error("invocation commit failed");
      committedEffects.push(...effects);
      records.set(idempotencyKey, result);
      return result;
    },
    effects(): unknown[] {
      return [...committedEffects];
    },
    result(idempotencyKey: string): unknown {
      return records.get(idempotencyKey);
    },
    has(idempotencyKey: string): boolean {
      return records.has(idempotencyKey);
    },
    setLifecycleRevision(revision: number): void {
      lifecycleRevision = revision;
    },
  };
}
