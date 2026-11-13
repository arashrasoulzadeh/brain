import { validateProjectId } from "../sdk-typescript/src/index.js";

export class FileGrantStore {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, objectId: string, actorId: string): void {
    validateProjectId(projectId, "invalid file project");
    if (!objectId || !actorId) throw new Error("invalid file grant");
    this.grants.add(`${projectId}:${objectId}:${actorId}`);
  }
}
