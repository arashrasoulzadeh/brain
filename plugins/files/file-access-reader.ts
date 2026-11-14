import { projectObjectKey } from "./object-key.js";

export class FileAccessReader {
  constructor(private readonly grants: ReadonlySet<string>) {}
  read(projectId: string, objectId: string, actorId: string): void {
    projectObjectKey(projectId, objectId);
    if (!actorId.trim()) throw new Error("invalid file actor");
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`))
      throw new Error("file access denied");
  }
}
