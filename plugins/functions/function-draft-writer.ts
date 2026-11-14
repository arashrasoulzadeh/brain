import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionDraftState } from "./function-draft-state.js";
export class FunctionDraftWriter {
  constructor(private readonly state: FunctionDraftState) {}
  save(projectId: string, functionId: string, source: string, savedAt = Date.now()) {
    validateProjectId(projectId);
    if (!functionId || !source || !Number.isFinite(savedAt))
      throw new Error("invalid function draft");
    const draft = { projectId, functionId, source, savedAt };
    this.state.drafts.set(`${projectId}:${functionId}`, draft);
    return { ...draft };
  }
}
