import type { InvocationCommitState } from "./invocation-commit-state.js";
export class InvocationExistenceReader { constructor(private readonly state: InvocationCommitState) {} has(idempotencyKey: string): boolean { return this.state.records.has(idempotencyKey); } }
