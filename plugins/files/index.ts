import { request } from "../sdk-typescript/src/index.js";

export function fileRequest(projectId: string, pluginId: string, version: string, operation: "files.upload" | "files.download" | "files.metadata" | "files.publish", objectId: string, payload: unknown = {}) {
  return request(`files-${operation}-${projectId}-${objectId}`, { protocol: 1, projectId, pluginId, version, operation, payload: { objectId, ...(payload as Record<string, unknown>) } });
}

export type PrivateBucketConfig = { endpoint: string; bucket: string; region: string; publicAccess: false };
export function configurePrivateBucket(config: Omit<PrivateBucketConfig, "publicAccess">): PrivateBucketConfig {
  if (!config.endpoint || !config.bucket || !config.region) throw new Error("invalid bucket configuration");
  const endpoint = new URL(config.endpoint);
  if (endpoint.protocol !== "https:") throw new Error("private bucket requires HTTPS");
  return { ...config, publicAccess: false };
}

export function projectObjectKey(projectId: string, objectId: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) || !/^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(objectId)) throw new Error("invalid object identity");
  return `projects/${projectId}/objects/${objectId}`;
}

export type PendingUpload = { projectId: string; objectId: string; contentType: string; size: number; status: "pending" | "published" | "aborted"; createdAt?: number; etag?: string };
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

export class FileQuotaLedger {
  private readonly reserved = new Map<string, number>();
  constructor(private readonly quotaByProject: ReadonlyMap<string, number>) {}
  reserve(projectId: string, size: number): void {
    if (!Number.isSafeInteger(size) || size < 0) throw new Error("invalid quota amount");
    const next = (this.reserved.get(projectId) ?? 0) + size;
    const quota = this.quotaByProject.get(projectId);
    if (quota === undefined || next > quota) throw new Error("file quota exceeded");
    this.reserved.set(projectId, next);
  }
  release(projectId: string, size: number): void { this.reserved.set(projectId, Math.max(0, (this.reserved.get(projectId) ?? 0) - size)); }
  reservedFor(projectId: string): number { return this.reserved.get(projectId) ?? 0; }
}

export class AuthorizedUploadStream {
  private readonly grants = new Set<string>();
  private readonly bytes = new Map<string, number>();
  authorize(projectId: string, objectId: string, actorId: string): void { this.grants.add(`${projectId}:${objectId}:${actorId}`); }
  write(projectId: string, objectId: string, actorId: string, chunk: Uint8Array): number {
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("upload denied");
    const key = `${projectId}:${objectId}`;
    const total = (this.bytes.get(key) ?? 0) + chunk.byteLength;
    this.bytes.set(key, total);
    return total;
  }
}

export class FileAccessPolicy {
  private readonly grants = new Set<string>();
  grant(projectId: string, objectId: string, actorId: string): void { this.grants.add(`${projectId}:${objectId}:${actorId}`); }
  read(projectId: string, objectId: string, actorId: string): void { this.check(projectId, objectId, actorId); }
  delete(projectId: string, objectId: string, actorId: string): void { this.check(projectId, objectId, actorId); }
  private check(projectId: string, objectId: string, actorId: string): void { if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("file access denied"); }
}

export function enforceUploadSize(size: number, maxSize: number): void {
  if (!Number.isSafeInteger(size) || !Number.isSafeInteger(maxSize) || size < 0 || maxSize < 0 || size > maxSize) throw new Error("upload size exceeded");
}

export function reconcileOrphanedObjects(projectId: string, storageKeys: readonly string[], publishedObjectIds: readonly string[]): string[] {
  const published = new Set(publishedObjectIds.map((objectId) => projectObjectKey(projectId, objectId)));
  return storageKeys.filter((key) => key.startsWith(`projects/${projectId}/objects/`) && !published.has(key));
}

export function protectedDownloadHeaders(filename: string): Record<string, string> {
  if (!filename || filename.includes("\r") || filename.includes("\n")) throw new Error("invalid download filename");
  return { "content-disposition": `attachment; filename="${filename.replaceAll('"', "")}"`, "cache-control": "private, no-store" };
}

export function requireExplicitPublication(publish: boolean): void { if (publish !== true) throw new Error("explicit publication required"); }

