import { type MatchmakingTicket, type AssignmentDecision } from "./types.js";
import { type MatchmakingConfigReader, type MatchmakingConfigStatus } from "./config.js";
import { validateProjectId, validateIdentifier } from "../sdk-typescript/src/index.js";

export type QueueRuntime = {
  sequence: number;
  decisionSequence: number;
  lifecycleRevision: number;
  pausedProjects: Set<string>;
  queues: Map<string, MatchmakingTicket[]>;
  configs?: { reader: MatchmakingConfigReader; status: MatchmakingConfigStatus };
};
export function createQueueRuntime(configs?: QueueRuntime["configs"]): QueueRuntime {
  return {
    sequence: 0,
    decisionSequence: 0,
    lifecycleRevision: 0,
    pausedProjects: new Set(),
    queues: new Map(),
    configs,
  };
}

export function isProjectAssignmentsPaused(runtime: QueueRuntime, projectId: string): boolean {
  validateProjectId(projectId);
  return runtime.pausedProjects.has(projectId);
}

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

export class TicketReader {
  constructor(private readonly runtime: QueueRuntime) {}
  peek(projectId: string, scope: string) {
    validateProjectId(projectId);
    validateIdentifier(scope, "invalid matchmaking scope");
    return [...(this.runtime.queues.get(`${projectId}:${scope}`) ?? [])];
  }
}

export class TicketMatcher {
  constructor(private readonly runtime: QueueRuntime) {}
  match(projectId: string, scope: string, matchSize: number) {
    validateProjectId(projectId);
    validateIdentifier(scope, "invalid matchmaking scope");
    const key = `${projectId}:${scope}`;
    const queue = this.runtime.queues.get(key) ?? [];
    if (matchSize < 1 || queue.length < matchSize) return [];
    const matched = queue.slice(0, matchSize);
    this.runtime.queues.set(key, queue.slice(matchSize));
    return matched.map((ticket) => ({ ...ticket }));
  }
}

export class TicketExpirer {
  constructor(private readonly runtime: QueueRuntime) {}
  expire(now = Date.now()): MatchmakingTicket[] {
    const expired: MatchmakingTicket[] = [];
    for (const [key, queue] of this.runtime.queues) {
      const active = queue.filter((ticket) => {
        if (ticket.expiresAt <= now) {
          expired.push({ ...ticket });
          return false;
        }
        return true;
      });
      this.runtime.queues.set(key, active);
    }
    return expired;
  }
}

export class TicketRequeuer {
  constructor(private readonly runtime: QueueRuntime) {}
  requeue(ticket: MatchmakingTicket, now = Date.now()): void {
    validateProjectId(ticket.projectId);
    validateIdentifier(ticket.scope, "invalid matchmaking scope");
    if (ticket.expiresAt <= now) throw new Error("ticket expired");
    const key = `${ticket.projectId}:${ticket.scope}`;
    const queue = this.runtime.queues.get(key) ?? [];
    if (queue.some((active) => active.id === ticket.id || active.playerId === ticket.playerId))
      throw new Error("active ticket already exists");
    this.runtime.queues.set(
      key,
      [...queue, { ...ticket }].sort((a, b) => a.createdAt - b.createdAt),
    );
  }
}

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
    validateIdentifier(scope, "invalid matchmaking scope");
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

export class TicketCanceller {
  constructor(private readonly runtime: QueueRuntime) {}

  cancelIncompatible(projectId: string, scope: string, compatibleVersion: string) {
    validateProjectId(projectId);
    validateIdentifier(scope, "invalid matchmaking scope");
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

export function pauseProjectAssignments(runtime: QueueRuntime, projectId: string): void {
  validateProjectId(projectId);
  runtime.pausedProjects.add(projectId);
}

export function resumeProjectAssignments(runtime: QueueRuntime, projectId: string): void {
  validateProjectId(projectId);
  runtime.pausedProjects.delete(projectId);
}

export function createMatchmakingQueue(configs?: {
  reader: MatchmakingConfigReader;
  status: MatchmakingConfigStatus;
}) {
  const runtime = createQueueRuntime(configs);
  const enqueue = new TicketEnqueuer(runtime);
  const reader = new TicketReader(runtime);
  const matcher = new TicketMatcher(runtime);
  const expirer = new TicketExpirer(runtime);
  const requeuer = new TicketRequeuer(runtime);
  const assigner = new AssignmentCreator(runtime);
  const canceller = new TicketCanceller(runtime);
  const scopeCanceller = new ScopeTicketCanceller(runtime);
  return {
    enqueue: enqueue.enqueue.bind(enqueue),
    peek: reader.peek.bind(reader),
    match: matcher.match.bind(matcher),
    expire: expirer.expire.bind(expirer),
    requeue: requeuer.requeue.bind(requeuer),
    assign: assigner.assign.bind(assigner),
    cancelIncompatible: canceller.cancelIncompatible.bind(canceller),
    cancelScope: scopeCanceller.cancelScope.bind(scopeCanceller),
    pauseAssignments: (projectId: string) => pauseProjectAssignments(runtime, projectId),
    resumeAssignments: (projectId: string) => resumeProjectAssignments(runtime, projectId),
    setLifecycleRevision: (revision: number) => {
      runtime.lifecycleRevision = revision;
    },
  };
}
