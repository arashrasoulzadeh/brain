import { projectObjectKey } from "./object-key.js";

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
