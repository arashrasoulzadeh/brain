import type { PendingUpload } from "./types.js";
import { projectObjectKey } from "./object-key.js";

export class PendingUploadReader {
  constructor(private readonly uploads: ReadonlyMap<string, PendingUpload>) {}
  get(projectId: string, objectId: string): PendingUpload | undefined {
    projectObjectKey(projectId, objectId);
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    return upload ? { ...upload } : undefined;
  }
}
