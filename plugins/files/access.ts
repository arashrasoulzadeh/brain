import { validateProjectId } from "../sdk-typescript/src/index.js";
import { type PrivateBucketConfig } from "./types.js";

export class FileGrantStore {
  constructor(private readonly grants: Set<string>) {}
  grant(projectId: string, objectId: string, actorId: string): void {
    validateProjectId(projectId, "invalid file project");
    if (!objectId || !actorId) throw new Error("invalid file grant");
    this.grants.add(`${projectId}:${objectId}:${actorId}`);
  }
}

export function projectObjectKey(projectId: string, objectId: string): string {
  validateProjectId(projectId, "invalid object identity");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(objectId)) throw new Error("invalid object identity");
  return `projects/${projectId}/objects/${objectId}`;
}

export class FileAccessReader {
  constructor(private readonly grants: ReadonlySet<string>) {}
  read(projectId: string, objectId: string, actorId: string): void {
    projectObjectKey(projectId, objectId);
    if (!actorId.trim()) throw new Error("invalid file actor");
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`))
      throw new Error("file access denied");
  }
}

export class FileAccessDeleter {
  constructor(private readonly grants: ReadonlySet<string>) {}
  delete(projectId: string, objectId: string, actorId: string): void {
    projectObjectKey(projectId, objectId);
    if (!actorId.trim()) throw new Error("invalid file actor");
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`))
      throw new Error("file access denied");
  }
}

export function configurePrivateBucket(
  config: Omit<PrivateBucketConfig, "publicAccess">,
): PrivateBucketConfig {
  if (!config.endpoint || !config.bucket || !config.region)
    throw new Error("invalid bucket configuration");
  const endpoint = new URL(config.endpoint);
  if (endpoint.protocol !== "https:") throw new Error("private bucket requires HTTPS");
  return { ...config, publicAccess: false };
}

export function protectedDownloadHeaders(filename: string): Record<string, string> {
  if (!filename || filename.includes("\r") || filename.includes("\n"))
    throw new Error("invalid download filename");
  return {
    "content-disposition": `attachment; filename="${filename.replaceAll('"', "")}"`,
    "cache-control": "private, no-store",
  };
}

export function requireExplicitPublication(publish: boolean): void {
  if (publish !== true) throw new Error("explicit publication required");
}
