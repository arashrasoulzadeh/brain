import { validateProjectId } from "../sdk-typescript/src/index.js";
import { validateIdentifier } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";
import { isProjectAssignmentsPaused } from "./is-project-assignments-paused.js";
export class TicketEnqueuer {
  constructor(private readonly runtime: QueueRuntime) {}
  enqueue(
    projectId: string,
    scope: string,
    playerId: string,
    now = Date.now(),
    pluginVersion = "unknown",
  ) {
    validateProjectId(projectId);
    validateIdentifier(scope, "invalid matchmaking scope");
    if (!scope || !playerId || !pluginVersion) throw new Error("invalid matchmaking ticket");
    if (isProjectAssignmentsPaused(this.runtime, projectId))
      throw new Error("matchmaking assignments paused");
    if (this.runtime.configs && !this.runtime.configs.status.isEnabled(projectId, scope))
      throw new Error("matchmaking scope disabled");
    const key = `${projectId}:${scope}`;
    const queue = this.runtime.queues.get(key) ?? [];
    if (queue.some((ticket) => ticket.playerId === playerId))
      throw new Error("active ticket already exists");
    const ttl = this.runtime.configs?.reader.get(projectId, scope).ticketTtlSeconds ?? 60;
    const ticket = {
      id: `${key}:ticket-${++this.runtime.sequence}`,
      projectId,
      scope,
      playerId,
      pluginVersion,
      createdAt: now,
      expiresAt: now + ttl * 1000,
    };
    this.runtime.queues.set(key, [...queue, ticket]);
    return { ...ticket };
  }
}
