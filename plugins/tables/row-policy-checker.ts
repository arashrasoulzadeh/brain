import { validateProjectId } from "../sdk-typescript/src/index.js";

export class RowPolicyChecker {
  constructor(private readonly grants: ReadonlySet<string>) {}
  allows(projectId: string, actorId: string): boolean { validateProjectId(projectId); return this.grants.has(`${projectId}:${actorId}`); }
}
