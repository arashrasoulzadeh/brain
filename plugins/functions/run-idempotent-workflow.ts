import { runCappedRetries } from "./run-capped-retries.js";
import type { createInvocationCommitStore } from "./create-invocation-commit-store.js";
export async function runIdempotentWorkflow<T>(store: ReturnType<typeof createInvocationCommitStore>, idempotencyKey: string, task: () => T | Promise<T>, maxAttempts: number): Promise<T> { if (store.has(idempotencyKey)) return store.result(idempotencyKey) as T; const result = await runCappedRetries(task, maxAttempts); return store.commit(idempotencyKey, [], result) as T; }
