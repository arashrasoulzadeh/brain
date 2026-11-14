import { validateProjectId } from "../sdk-typescript/src/index.js";

export class RowPolicyGranter {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, actorId: string): boolean {
    validateProjectId(projectId);
    if (!actorId) return false;
    this.grants.add(`${projectId}:${actorId}`);
    return true;
  }
}
