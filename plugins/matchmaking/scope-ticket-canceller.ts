import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";

export class ScopeTicketCanceller {
  constructor(private readonly runtime: QueueRuntime) {}

  cancelScope(projectId: string, scope: string) {
    validateProjectId(projectId);
    if (!scope) throw new Error("matchmaking scope required");
    const key = `${projectId}:${scope}`;
    const cancelled = this.runtime.queues.get(key) ?? [];
    this.runtime.queues.delete(key);
    return cancelled.map((ticket) => ({ ...ticket }));
  }
}
