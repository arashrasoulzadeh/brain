import { projectObjectKey } from "./access.js";
import { validateProjectId, validateBoundedInteger } from "../sdk-typescript/src/index.js";
import { type PendingUpload } from "./types.js";

export class AuthorizedUploadWriter {
  constructor(
    private readonly grants: ReadonlySet<string>,
    private readonly bytes: Map<string, number>,
  ) {}
  write(projectId: string, objectId: string, actorId: string, chunk: Uint8Array): number {
    projectObjectKey(projectId, objectId);
    if (!actorId.trim()) throw new Error("invalid upload actor");
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("upload denied");
    const key = `${projectId}:${objectId}`;
    const total = (this.bytes.get(key) ?? 0) + chunk.byteLength;
    this.bytes.set(key, total);
    return total;
  }
}

export class UploadGrantStore {
  constructor(private readonly grants: Set<string>) {}
  authorize(projectId: string, objectId: string, actorId: string): void {
    validateProjectId(projectId, "invalid upload project");
    if (!objectId || !actorId) throw new Error("invalid upload grant");
    this.grants.add(`${projectId}:${objectId}:${actorId}`);
  }
}

export class PendingUploadCreator {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  create(
    projectId: string,
    objectId: string,
    contentType: string,
    size: number,
    createdAt = Date.now(),
  ): PendingUpload {
    projectObjectKey(projectId, objectId);
    if (!contentType || !Number.isSafeInteger(size) || size < 0 || !Number.isFinite(createdAt))
      throw new Error("invalid upload metadata");
    const upload = {
      projectId,
      objectId,
      contentType,
      size,
      status: "pending" as const,
      createdAt,
    };
    this.uploads.set(`${projectId}:${objectId}`, upload);
    return { ...upload };
  }
}

export class PendingUploadReader {
  constructor(private readonly uploads: ReadonlyMap<string, PendingUpload>) {}
  get(projectId: string, objectId: string): PendingUpload | undefined {
    projectObjectKey(projectId, objectId);
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    return upload ? { ...upload } : undefined;
  }
}

export class PendingUploadPublisher {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  publish(projectId: string, objectId: string, etag: string): PendingUpload {
    projectObjectKey(projectId, objectId);
    const upload = this.uploads.get(`${projectId}:${objectId}`);
    if (!upload || upload.status !== "pending" || !etag)
      throw new Error("object upload not confirmed");
    const published = { ...upload, status: "published" as const, etag };
    this.uploads.set(`${projectId}:${objectId}`, published);
    return { ...published };
  }
}

export class PendingUploadReconciler {
  constructor(private readonly uploads: Map<string, PendingUpload>) {}
  reconcile(now: number, maxAgeMs: number): number {
    let count = 0;
    for (const [key, upload] of this.uploads)
      if (upload.status === "pending" && now - (upload.createdAt ?? now) > maxAgeMs) {
        this.uploads.set(key, { ...upload, status: "aborted" });
        count += 1;
      }
    return count;
  }
}

export function reconcileOrphanedObjects(
  projectId: string,
  storageKeys: readonly string[],
  publishedObjectIds: readonly string[],
): string[] {
  const published = new Set(
    publishedObjectIds.map((objectId) => projectObjectKey(projectId, objectId)),
  );
  return storageKeys.filter(
    (key) => key.startsWith(`projects/${projectId}/objects/`) && !published.has(key),
  );
}

export function enforceUploadSize(size: number, maxSize: number): void {
  validateBoundedInteger(size, maxSize, "upload size exceeded");
}
