import type { ProjectWorkflowAction } from "./register-project-workflow-action.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export function authorizeProjectWorkflowAction(
  actions: ReadonlyMap<string, ProjectWorkflowAction>,
  grants: ReadonlySet<string>,
  projectId: string,
  actionId: string,
): ProjectWorkflowAction {
  validateProjectId(projectId);
  const action = actions.get(`${projectId}:${actionId}`);
  if (!action || !grants.has(`${projectId}:${actionId}:${action.capability}`)) {
    throw new Error("workflow grant revoked");
  }
  return { ...action };
}
