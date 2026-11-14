import type { PendingUpload } from "./types.js";

export class PendingUploadPublisher {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  publish(projectId: string, objectId: string, etag: string): PendingUpload {
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    if (!upload || upload.status !== "pending" || !etag)
      throw new Error("object upload not confirmed");
    const published = { ...upload, status: "published" as const, etag };
    this.uploads.set(`${projectId}:${objectId}`, published);
    return { ...published };
  }
}
