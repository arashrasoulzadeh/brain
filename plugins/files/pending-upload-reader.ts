import type { PendingUpload } from "./types.js";

export class PendingUploadReader {
  constructor(private readonly uploads: ReadonlyMap<string, PendingUpload>) {}
  get(projectId: string, objectId: string): PendingUpload | undefined {
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    return upload ? { ...upload } : undefined;
  }
}
