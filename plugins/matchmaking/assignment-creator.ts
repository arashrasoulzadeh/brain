import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { AssignmentDecision } from "./types.js";
import type { QueueRuntime } from "./queue-runtime.js";
import { TicketMatcher } from "./ticket-matcher.js";
import { isProjectAssignmentsPaused } from "./is-project-assignments-paused.js";
export class AssignmentCreator {
  private readonly matcher: TicketMatcher;
  constructor(private readonly runtime: QueueRuntime) {
    this.matcher = new TicketMatcher(runtime);
  }
  assign(
    projectId: string,
    scope: string,
    matchSize: number,
    now = Date.now(),
    expectedRevision = this.runtime.lifecycleRevision,
  ): AssignmentDecision | null {
    validateProjectId(projectId);
    if (isProjectAssignmentsPaused(this.runtime, projectId))
      throw new Error("matchmaking assignments paused");
    if (expectedRevision !== this.runtime.lifecycleRevision)
      throw new Error("stale matchmaking lifecycle revision");
    const tickets = this.matcher.match(projectId, scope, matchSize);
    if (tickets.length !== matchSize) return null;
    if (expectedRevision !== this.runtime.lifecycleRevision)
      throw new Error("stale matchmaking lifecycle revision");
    return {
      id: `${projectId}:${scope}:assignment-${++this.runtime.decisionSequence}`,
      projectId,
      scope,
      ticketIds: tickets.map((ticket) => ticket.id),
      createdAt: now,
    };
  }
}
