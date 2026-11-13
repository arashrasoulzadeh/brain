import type { FunctionDraftState } from "./function-draft-state.js";
export class FunctionDraftReader { constructor(private readonly state: FunctionDraftState) {} get(projectId: string, functionId: string) { const draft = this.state.drafts.get(`${projectId}:${functionId}`); return draft ? { ...draft } : undefined; } }
