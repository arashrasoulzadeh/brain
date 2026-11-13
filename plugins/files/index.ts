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

export type PendingUpload = { projectId: string; objectId: string; contentType: string; size: number; status: "pending" };
export class PendingUploadStore {
  private readonly uploads = new Map<string, PendingUpload>();
  create(projectId: string, objectId: string, contentType: string, size: number): PendingUpload {
    projectObjectKey(projectId, objectId);
    if (!contentType || !Number.isSafeInteger(size) || size < 0) throw new Error("invalid upload metadata");
    const upload = { projectId, objectId, contentType, size, status: "pending" as const };
    this.uploads.set(`${projectId}:${objectId}`, upload);
    return { ...upload };
  }
  get(projectId: string, objectId: string): PendingUpload | undefined { const upload = this.uploads.get(`${projectId}:${objectId}`); return upload ? { ...upload } : undefined; }
}

