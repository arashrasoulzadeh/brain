import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";

export class TicketCanceller {
  constructor(private readonly runtime: QueueRuntime) {}

  cancelIncompatible(projectId: string, scope: string, compatibleVersion: string) {
    validateProjectId(projectId);
    if (!scope || !compatibleVersion) throw new Error("invalid matchmaking cancellation");
    const key = `${projectId}:${scope}`;
    const queue = this.runtime.queues.get(key) ?? [];
    const cancelled = queue.filter((ticket) => ticket.pluginVersion !== compatibleVersion);
    this.runtime.queues.set(
      key,
      queue.filter((ticket) => ticket.pluginVersion === compatibleVersion),
    );
    return cancelled.map((ticket) => ({ ...ticket }));
  }
}
