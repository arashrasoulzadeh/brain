type Approval = { requester: string; approved: boolean };

export function createApprovalStore() {
  const approvals = new Map<string, Approval>();
  return {
    request(runId: string, requester: string): void {
      if (!runId || !requester || approvals.has(runId))
        throw new Error("invalid approval request");
      approvals.set(runId, { requester, approved: false });
    },
    approve(runId: string, approver: string): boolean {
      const approval = approvals.get(runId);
      if (!approval || approval.approved || !approver || approver === approval.requester)
        return false;
      approval.approved = true;
      return true;
    },
    isApproved(runId: string): boolean {
      return approvals.get(runId)?.approved === true;
    },
  };
}
