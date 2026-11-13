import { validateProjectId } from "../sdk-typescript/src/index.js";

export class UploadGrantStore {
  constructor(private readonly grants: Set<string>) {}
  authorize(projectId: string, objectId: string, actorId: string): void {
    validateProjectId(projectId, "invalid upload project");
    if (!objectId || !actorId) throw new Error("invalid upload grant");
    this.grants.add(`${projectId}:${objectId}:${actorId}`);
  }
}
