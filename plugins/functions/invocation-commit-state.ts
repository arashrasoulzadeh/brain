export type InvocationCommitState = {
  records: Map<string, unknown>;
  effects: unknown[];
  lifecycleRevision: number;
};
export function createInvocationCommitState(): InvocationCommitState {
  return { records: new Map(), effects: [], lifecycleRevision: 0 };
}
