import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";
export class TicketReader {
  constructor(private readonly runtime: QueueRuntime) {}
  peek(projectId: string, scope: string) {
    validateProjectId(projectId);
    return [...(this.runtime.queues.get(`${projectId}:${scope}`) ?? [])];
  }
}
