import type { ApiChangeState } from "./api-change-state.js";
export class ApiChangeReader {
  constructor(private readonly state: ApiChangeState) {}
  list(projectId: string) {
    return this.state.changes
      .filter((change) => change.projectId === projectId)
      .map((change) => ({ ...change }));
  }
}
