import { projectSchemaName } from "./project-schema-name.js";

export function projectRoleName(projectId: string): string {
  return `runtime_${projectSchemaName(projectId)}`;
}
