import { validateProjectId } from "../sdk-typescript/src/index.js";

export class FileAccessPolicy {
  private readonly grants = new Set<string>();
  grant(projectId: string, objectId: string, actorId: string): void { validateProjectId(projectId, "invalid file project"); if (!objectId || !actorId) throw new Error("invalid file grant"); this.grants.add(`${projectId}:${objectId}:${actorId}`); }
  read(projectId: string, objectId: string, actorId: string): void { this.check(projectId, objectId, actorId); }
  delete(projectId: string, objectId: string, actorId: string): void { this.check(projectId, objectId, actorId); }
  private check(projectId: string, objectId: string, actorId: string): void { if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("file access denied"); }
}
