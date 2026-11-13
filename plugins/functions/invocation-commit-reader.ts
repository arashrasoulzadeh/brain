import type { InvocationCommitState } from "./invocation-commit-state.js";
export class InvocationCommitReader { constructor(private readonly state: InvocationCommitState) {} effects(): unknown[] { return [...this.state.effects]; } }
