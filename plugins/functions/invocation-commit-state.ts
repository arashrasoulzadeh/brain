export type InvocationCommitState = { records: Map<string, unknown>; effects: unknown[] };
export function createInvocationCommitState(): InvocationCommitState { return { records: new Map(), effects: [] }; }
