import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { WorkflowSummary } from "./types.js";

export function workflowManagementPage(projectId: string, workflows: readonly WorkflowSummary[]) {
  validateProjectId(projectId, "invalid workflow management");
  if (
    workflows.some(
      (item) =>
        !item.id ||
        !item.status ||
        (item.lastRunAt !== undefined && !Number.isFinite(item.lastRunAt)),
    )
  )
    throw new Error("invalid workflow management");
  return { projectId, title: "Workflows", workflows: workflows.map((item) => ({ ...item })) };
}
