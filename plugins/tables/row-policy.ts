import { validateProjectId } from "../sdk-typescript/src/index.js";

export class RowPolicy {
  private readonly grants = new Set<string>();
  grant(projectId: string, actorId: string): boolean { validateProjectId(projectId); if (!actorId) return false; this.grants.add(`${projectId}:${actorId}`); return true; }
  allows(projectId: string, actorId: string): boolean { validateProjectId(projectId); return this.grants.has(`${projectId}:${actorId}`); }
}
