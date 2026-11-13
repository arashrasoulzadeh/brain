export type ApprovalState = { approvals: Map<string, { requester: string; approved: boolean }> };
export function createApprovalState(): ApprovalState { return { approvals: new Map() }; }
