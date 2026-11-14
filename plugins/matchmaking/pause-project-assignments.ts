import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";

export function pauseProjectAssignments(runtime: QueueRuntime, projectId: string): void {
  validateProjectId(projectId);
  runtime.pausedProjects.add(projectId);
}
