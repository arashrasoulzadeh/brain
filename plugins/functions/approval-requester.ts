import type { ApprovalState } from "./approval-state.js";
export class ApprovalRequester {
  constructor(private readonly state: ApprovalState) {}
  request(runId: string, requester: string): void {
    if (!runId || !requester || this.state.approvals.has(runId))
      throw new Error("invalid approval request");
    this.state.approvals.set(runId, { requester, approved: false });
  }
}
