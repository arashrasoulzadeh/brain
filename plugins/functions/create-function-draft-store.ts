import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionDraft } from "./index.js";

export function createFunctionDraftStore() {
  const drafts = createProjectMap<FunctionDraft>();
  return {
    save(projectId: string, functionId: string, source: string, savedAt = Date.now()) {
      validateProjectId(projectId);
      if (!functionId || !source || !Number.isFinite(savedAt))
        throw new Error("invalid function draft");
      const draft = { projectId, functionId, source, savedAt };
      drafts.set(projectId, functionId, draft);
      return { ...draft };
    },
    get(projectId: string, functionId: string) {
      const draft = drafts.get(projectId, functionId);
      return draft ? { ...draft } : undefined;
    },
  };
}
