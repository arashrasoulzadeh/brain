import type { PendingUpload } from "./types.js";
import { projectObjectKey } from "./object-key.js";

export class PendingUploadCreator {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  create(projectId: string, objectId: string, contentType: string, size: number, createdAt = Date.now()): PendingUpload { projectObjectKey(projectId, objectId); if (!contentType || !Number.isSafeInteger(size) || size < 0 || !Number.isFinite(createdAt)) throw new Error("invalid upload metadata"); const upload = { projectId, objectId, contentType, size, status: "pending" as const, createdAt }; this.uploads.set(`${projectId}:${objectId}`, upload); return { ...upload }; }
}
