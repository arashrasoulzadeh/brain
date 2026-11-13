import { validateProjectId } from "../sdk-typescript/src/index.js";

export function projectSchemaName(projectId: string): string {
  validateProjectId(projectId);
  return `project_${projectId}`;
}
