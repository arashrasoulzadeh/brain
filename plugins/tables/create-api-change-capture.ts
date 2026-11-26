import { createCommitRollback, validateProjectId } from "../sdk-typescript/src/index.js";
import type { ApiChange } from "./index.js";

export function createApiChangeCapture() {
  const changes: ApiChange[] = [];
  return {
    begin(projectId: string, operation: string, payload: unknown) {
      validateProjectId(projectId, "invalid API change");
      if (!operation) throw new Error("invalid API change");
      return createCommitRollback(() => { changes.push({ projectId, operation, payload }); });
    },
    list(projectId: string) {
      return changes.filter((change) => change.projectId === projectId).map((change) => ({ ...change }));
    },
  };
}
