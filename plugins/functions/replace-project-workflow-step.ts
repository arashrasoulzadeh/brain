import type { ProjectWorkflowAction } from "./register-project-workflow-action.js";

export function replaceProjectWorkflowStep(
  actions: Map<string, ProjectWorkflowAction>,
  projectId: string,
  stepId: string,
  replacement: ProjectWorkflowAction,
): ProjectWorkflowAction {
  const key = `${projectId}:${stepId}`;
  if (
    !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId) ||
    replacement.projectId !== projectId ||
    replacement.id !== stepId ||
    !actions.has(key) ||
    !replacement.operation ||
    replacement.operation.startsWith("deploy.") ||
    !replacement.capability
  ) {
    throw new Error("invalid workflow step replacement");
  }
  const updated = { ...replacement };
  actions.set(key, updated);
  return updated;
}
