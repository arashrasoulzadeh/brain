import type { FunctionDraftState } from "./function-draft-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class FunctionDraftReader {
  constructor(private readonly state: FunctionDraftState) {}
  get(projectId: string, functionId: string) {
    validateProjectId(projectId);
    const draft = this.state.drafts.get(`${projectId}:${functionId}`);
    return draft ? { ...draft } : undefined;
  }
}
