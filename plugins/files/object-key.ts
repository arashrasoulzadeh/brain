import { validateProjectId } from "../sdk-typescript/src/index.js";

export function projectObjectKey(projectId: string, objectId: string): string {
  validateProjectId(projectId, "invalid object identity");
  if (!/^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(objectId)) throw new Error("invalid object identity");
  return `projects/${projectId}/objects/${objectId}`;
}
