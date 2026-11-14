import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";

export function isProjectAssignmentsPaused(runtime: QueueRuntime, projectId: string): boolean {
  validateProjectId(projectId);
  return runtime.pausedProjects.has(projectId);
}
