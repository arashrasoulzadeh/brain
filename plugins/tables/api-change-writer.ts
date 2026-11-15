import type { ApiChangeState } from "./api-change-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class ApiChangeWriter {
  constructor(private readonly state: ApiChangeState) {}
  begin(projectId: string, operation: string, payload: unknown) {
    validateProjectId(projectId, "invalid API change");
    if (!operation) throw new Error("invalid API change");
    let finished = false;
    const finish = (commit: boolean) => {
      if (finished) throw new Error("transaction already finished");
      finished = true;
      if (commit) this.state.changes.push({ projectId, operation, payload });
    };
    return { commit: () => finish(true), rollback: () => finish(false) };
  }
}
