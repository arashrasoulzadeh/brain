import { validateProjectId } from "../sdk-typescript/src/index.js";

export class AuthorizedUploadStream {
  private readonly grants = new Set<string>();
  private readonly bytes = new Map<string, number>();
  authorize(projectId: string, objectId: string, actorId: string): void { validateProjectId(projectId, "invalid upload project"); if (!objectId || !actorId) throw new Error("invalid upload grant"); this.grants.add(`${projectId}:${objectId}:${actorId}`); }
  write(projectId: string, objectId: string, actorId: string, chunk: Uint8Array): number {
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("upload denied");
    const key = `${projectId}:${objectId}`;
    const total = (this.bytes.get(key) ?? 0) + chunk.byteLength;
    this.bytes.set(key, total);
    return total;
  }
}
