import { projectObjectKey } from "./object-key.js";

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
