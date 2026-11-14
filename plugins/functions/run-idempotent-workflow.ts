import { runCappedRetries } from "./run-capped-retries.js";
import type { createInvocationCommitStore } from "./create-invocation-commit-store.js";

const inFlightWorkflows = new WeakMap<object, Map<string, Promise<unknown>>>();

export async function runIdempotentWorkflow<T>(
  store: ReturnType<typeof createInvocationCommitStore>,
  idempotencyKey: string,
  task: () => T | Promise<T>,
  maxAttempts: number,
): Promise<T> {
  if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T;
  let storeWorkflows = inFlightWorkflows.get(store);
  if (!storeWorkflows) {
    storeWorkflows = new Map();
    inFlightWorkflows.set(store, storeWorkflows);
  }
  const existing = storeWorkflows.get(idempotencyKey);
  if (existing) return (await existing) as T;
  const execution = (async () => {
    const result = await runCappedRetries(task, maxAttempts);
    return store.commit(idempotencyKey, [], result) as T;
  })();
  storeWorkflows.set(idempotencyKey, execution);
  try {
    return await execution;
  } finally {
    storeWorkflows.delete(idempotencyKey);
  }
}
