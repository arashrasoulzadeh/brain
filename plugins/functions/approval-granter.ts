import type { ApprovalState } from "./approval-state.js";
export class ApprovalGranter { constructor(private readonly state: ApprovalState) {} approve(runId: string, approver: string): boolean { const approval = this.state.approvals.get(runId); if (!approval || approval.approved || !approver || approver === approval.requester) return false; approval.approved = true; return true; } }
