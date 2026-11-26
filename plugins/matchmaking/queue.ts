import { type MatchmakingTicket, type AssignmentDecision } from "./types.js";
import { type MatchmakingConfigReader, type MatchmakingConfigStatus } from "./config.js";
import { validateProjectId, validateIdentifier } from "../sdk-typescript/src/index.js";

type QueueRuntime = {
  sequence: number;
  decisionSequence: number;
  lifecycleRevision: number;
  pausedProjects: Set<string>;
  queues: Map<string, MatchmakingTicket[]>;
  configs?: { reader: MatchmakingConfigReader; status: MatchmakingConfigStatus };
};
function createQueueRuntime(configs?: QueueRuntime["configs"]): QueueRuntime {
  return {
    sequence: 0,
    decisionSequence: 0,
    lifecycleRevision: 0,
    pausedProjects: new Set(),
    queues: new Map(),
    configs,
  };
}

function isProjectAssignmentsPaused(runtime: QueueRuntime, projectId: string): boolean {
  validateProjectId(projectId);
  return runtime.pausedProjects.has(projectId);
}

function pauseProjectAssignments(runtime: QueueRuntime, projectId: string): void {
  validateProjectId(projectId);
  runtime.pausedProjects.add(projectId);
}

function resumeProjectAssignments(runtime: QueueRuntime, projectId: string): void {
  validateProjectId(projectId);
  runtime.pausedProjects.delete(projectId);
}

export function createMatchmakingQueue(configs?: QueueRuntime["configs"]) {
  const runtime = createQueueRuntime(configs);

  function queueKey(projectId: string, scope: string): string {
    validateProjectId(projectId);
    validateIdentifier(scope, "invalid matchmaking scope");
    return `${projectId}:${scope}`;
  }

  function enqueue(
    projectId: string,
    scope: string,
    playerId: string,
    now = Date.now(),
    pluginVersion = "unknown",
  ) {
    const key = queueKey(projectId, scope);
    if (!scope || !playerId || !pluginVersion) throw new Error("invalid matchmaking ticket");
    if (isProjectAssignmentsPaused(runtime, projectId))
      throw new Error("matchmaking assignments paused");
    if (runtime.configs && !runtime.configs.status.isEnabled(projectId, scope))
      throw new Error("matchmaking scope disabled");
    const queue = runtime.queues.get(key) ?? [];
    if (queue.some((ticket) => ticket.playerId === playerId))
      throw new Error("active ticket already exists");
    const ttl = runtime.configs?.reader.get(projectId, scope).ticketTtlSeconds ?? 60;
    const ticket = {
      id: `${key}:ticket-${++runtime.sequence}`,
      projectId,
      scope,
      playerId,
      pluginVersion,
      createdAt: now,
      expiresAt: now + ttl * 1000,
    };
    runtime.queues.set(key, [...queue, ticket]);
    return { ...ticket };
  }

  function peek(projectId: string, scope: string) {
    const key = queueKey(projectId, scope);
    return (runtime.queues.get(key) ?? []).map((ticket) => ({ ...ticket }));
  }

  function match(projectId: string, scope: string, matchSize: number) {
    const key = queueKey(projectId, scope);
    const queue = runtime.queues.get(key) ?? [];
    if (!Number.isSafeInteger(matchSize) || matchSize < 1 || queue.length < matchSize) return [];
    const matched = queue.slice(0, matchSize);
    runtime.queues.set(key, queue.slice(matchSize));
    return matched.map((ticket) => ({ ...ticket }));
  }

  function expire(now = Date.now()): MatchmakingTicket[] {
    const expired: MatchmakingTicket[] = [];
    for (const [key, queue] of runtime.queues) {
      const active = queue.filter((ticket) => {
        if (ticket.expiresAt <= now) {
          expired.push({ ...ticket });
          return false;
        }
        return true;
      });
      runtime.queues.set(key, active);
    }
    return expired;
  }

  function requeue(ticket: MatchmakingTicket, now = Date.now()): void {
    validateProjectId(ticket.projectId);
    validateIdentifier(ticket.scope, "invalid matchmaking scope");
    if (ticket.expiresAt <= now) throw new Error("ticket expired");
    const key = `${ticket.projectId}:${ticket.scope}`;
    const queue = runtime.queues.get(key) ?? [];
    if (queue.some((active) => active.id === ticket.id || active.playerId === ticket.playerId))
      throw new Error("active ticket already exists");
    runtime.queues.set(
      key,
      [...queue, { ...ticket }].sort((a, b) => a.createdAt - b.createdAt),
    );
  }

  function assign(
    projectId: string,
    scope: string,
    matchSize: number,
    now = Date.now(),
    expectedRevision = runtime.lifecycleRevision,
  ): AssignmentDecision | null {
    queueKey(projectId, scope);
    if (isProjectAssignmentsPaused(runtime, projectId))
      throw new Error("matchmaking assignments paused");
    if (expectedRevision !== runtime.lifecycleRevision)
      throw new Error("stale matchmaking lifecycle revision");
    if (!Number.isSafeInteger(matchSize) || matchSize < 1) return null;
    const tickets = match(projectId, scope, matchSize);
    if (tickets.length !== matchSize) return null;
    if (expectedRevision !== runtime.lifecycleRevision)
      throw new Error("stale matchmaking lifecycle revision");
    return {
      id: `${projectId}:${scope}:assignment-${++runtime.decisionSequence}`,
      projectId,
      scope,
      ticketIds: tickets.map((ticket) => ticket.id),
      createdAt: now,
    };
  }

  function cancelIncompatible(projectId: string, scope: string, compatibleVersion: string) {
    const key = queueKey(projectId, scope);
    if (!scope || !compatibleVersion) throw new Error("invalid matchmaking cancellation");
    const queue = runtime.queues.get(key) ?? [];
    const cancelled = queue.filter((ticket) => ticket.pluginVersion !== compatibleVersion);
    runtime.queues.set(
      key,
      queue.filter((ticket) => ticket.pluginVersion === compatibleVersion),
    );
    return cancelled.map((ticket) => ({ ...ticket }));
  }

  function cancelScope(projectId: string, scope: string) {
    validateProjectId(projectId);
    if (!scope) throw new Error("matchmaking scope required");
    const key = `${projectId}:${scope}`;
    const cancelled = runtime.queues.get(key) ?? [];
    runtime.queues.delete(key);
    return cancelled.map((ticket) => ({ ...ticket }));
  }

  return {
    enqueue, peek, match, expire, requeue, assign, cancelIncompatible, cancelScope,
    pauseAssignments: (projectId: string) => pauseProjectAssignments(runtime, projectId),
    resumeAssignments: (projectId: string) => resumeProjectAssignments(runtime, projectId),
    setLifecycleRevision: (revision: number) => {
      runtime.lifecycleRevision = revision;
    },
  };
}
