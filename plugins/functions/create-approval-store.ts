import { createApprovalState } from "./approval-state.js";
import { ApprovalRequester } from "./approval-requester.js";
import { ApprovalGranter } from "./approval-granter.js";
import { ApprovalReader } from "./approval-reader.js";
export function createApprovalStore() { const state = createApprovalState(); const requester = new ApprovalRequester(state); const granter = new ApprovalGranter(state); const reader = new ApprovalReader(state); return { request: requester.request.bind(requester), approve: granter.approve.bind(granter), isApproved: reader.isApproved.bind(reader) }; }
