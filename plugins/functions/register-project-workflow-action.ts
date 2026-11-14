import { validateProjectId } from "../sdk-typescript/src/index.js";

export type ProjectWorkflowAction = {
  readonly projectId: string;
  readonly id: string;
  readonly operation: string;
  readonly capability: string;
};

export function registerProjectWorkflowAction(
  actions: Map<string, ProjectWorkflowAction>,
  action: ProjectWorkflowAction,
): ProjectWorkflowAction {
  try {
    validateProjectId(action.projectId);
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(action.id)) throw new Error("invalid action id");
  } catch {
    throw new Error("invalid project workflow action");
  }
  if (
    !action.operation ||
    !action.capability ||
    action.operation.startsWith("deploy.") ||
    action.operation.startsWith("release.") ||
    actions.has(`${action.projectId}:${action.id}`)
  ) {
    throw new Error("invalid project workflow action");
  }
  const registered = { ...action };
  actions.set(`${action.projectId}:${action.id}`, registered);
  return registered;
}
