import type { ApprovalState } from "./approval-state.js";
export class ApprovalReader {
  constructor(private readonly state: ApprovalState) {}
  isApproved(runId: string): boolean {
    return this.state.approvals.get(runId)?.approved === true;
  }
}
