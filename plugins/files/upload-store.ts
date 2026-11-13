import type { PendingUpload } from "./types.js";
import { projectObjectKey } from "./object-key.js";

export class PendingUploadStore {
  private readonly uploads = new Map<string, PendingUpload>();
  create(projectId: string, objectId: string, contentType: string, size: number, createdAt = Date.now()): PendingUpload {
    projectObjectKey(projectId, objectId);
    if (!contentType || !Number.isSafeInteger(size) || size < 0 || !Number.isFinite(createdAt)) throw new Error("invalid upload metadata");
    const upload = { projectId, objectId, contentType, size, status: "pending" as const, createdAt };
    this.uploads.set(`${projectId}:${objectId}`, upload);
    return { ...upload };
  }
  get(projectId: string, objectId: string): PendingUpload | undefined { const upload = this.uploads.get(`${projectId}:${objectId}`); return upload ? { ...upload } : undefined; }
  publish(projectId: string, objectId: string, etag: string): PendingUpload {
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    if (!upload || upload.status !== "pending" || !etag) throw new Error("object upload not confirmed");
    const published = { ...upload, status: "published" as const, etag };
    this.uploads.set(`${projectId}:${objectId}`, published);
    return { ...published };
  }
  reconcile(now: number, maxAgeMs: number): number {
    let count = 0;
    for (const [key, upload] of this.uploads) if (upload.status === "pending" && now - (upload.createdAt ?? now) > maxAgeMs) { this.uploads.set(key, { ...upload, status: "aborted" }); count += 1; }
    return count;
  }
}
