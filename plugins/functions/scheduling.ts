import { createProjectLeaseStore, createProjectSet, validateProjectId, requireProjectInput, type PluginContext } from "../sdk-typescript/src/index.js";
import { type FunctionSchedule, type RecordChangeTrigger } from "./index.js";

export function createScheduleStore() {
  const schedules: FunctionSchedule[] = [];
  const claims = createProjectLeaseStore("invalid schedule claim");
  const recordOccurrence = createProjectSet<string>();
  const recordEnqueue = createProjectSet<number>();

  return {
    add(projectId: string, functionId: string, expression: string, dueAt: number) {
      validateProjectId(projectId);
      if (!functionId || expression.trim().split(/\s+/).length !== 5 || !Number.isSafeInteger(dueAt) || dueAt < 0)
        throw new Error("invalid schedule");
      const schedule = { projectId, functionId, expression: expression.trim(), dueAt };
      schedules.push(schedule);
      return { ...schedule };
    },
    due(now: number) {
      return schedules.filter((schedule) => schedule.dueAt <= now).map((schedule) => ({ ...schedule }));
    },
    claim(projectId: string, functionId: string, workerId: string, now: number, leaseMs: number): boolean {
      validateProjectId(projectId);
      if (!workerId || !Number.isSafeInteger(now) || !Number.isSafeInteger(leaseMs) || leaseMs < 1)
        throw new Error("invalid schedule claim");
      if (!schedules.some((candidate) => candidate.projectId === projectId && candidate.functionId === functionId && candidate.dueAt <= now))
        return false;
      return claims.acquire(projectId, functionId, workerId, now, leaseMs);
    },
    recordOccurrence(projectId: string, functionId: string, occurrenceUtc: string): boolean {
      validateProjectId(projectId);
      if (!functionId || !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/.test(occurrenceUtc) || Number.isNaN(Date.parse(occurrenceUtc)))
        throw new Error("invalid UTC occurrence");
      return recordOccurrence(projectId, functionId, occurrenceUtc);
    },
    enqueueLatestMissed(projectId: string, functionId: string, now: number) {
      validateProjectId(projectId);
      if (!Number.isSafeInteger(now) || now < 0) throw new Error("invalid schedule time");
      const schedule = schedules.filter((candidate) => candidate.projectId === projectId && candidate.functionId === functionId && candidate.dueAt <= now)
        .sort((left, right) => right.dueAt - left.dueAt)[0];
      if (!schedule) return undefined;
      return recordEnqueue(projectId, functionId, schedule.dueAt) ? { ...schedule } : undefined;
    },
    enqueueVersion(projectId: string, functionId: string, functionVersion: number, enqueuedAt: number) {
      validateProjectId(projectId);
      if (!functionId || !Number.isSafeInteger(functionVersion) || functionVersion < 1 || !Number.isSafeInteger(enqueuedAt) || enqueuedAt < 0)
        throw new Error("invalid function enqueue");
      return { projectId, functionId, functionVersion, enqueuedAt };
    },
  };
}

export function createRecordChangeTriggerStore() {
  const triggers: RecordChangeTrigger[] = [];

  function match(projectId: string, collection: string, operation: string): RecordChangeTrigger[] {
    validateProjectId(projectId);
    return triggers
      .filter(
        (trigger) =>
          trigger.projectId === projectId &&
          trigger.collection === collection &&
          trigger.operation === operation,
      )
      .map((trigger) => ({ ...trigger, target: { ...trigger.target } }));
  }

  return {
    add(
      projectId: string,
      collection: string,
      operation: string,
      target: { functionId: string; version: number },
    ): RecordChangeTrigger {
      validateProjectId(projectId);
      if (!collection || !operation || !target.functionId || !Number.isSafeInteger(target.version) || target.version < 1)
        throw new Error("invalid record trigger");
      const trigger = { projectId, collection, operation, target: { ...target } };
      triggers.push(trigger);
      return { ...trigger, target: { ...trigger.target } };
    },
    match,
    dispatch(
      projectId: string,
      collection: string,
      operation: string,
      committed: boolean,
      ancestry: readonly string[] = [],
    ): RecordChangeTrigger[] {
      return committed
        ? match(projectId, collection, operation).filter((trigger) => !ancestry.includes(trigger.target.functionId))
        : [];
    },
  };
}

export const createJobLeaseStore = () => createProjectLeaseStore();

/** Bind scheduling, triggers and leases to the trusted project once, not per job. */
export function bindProjectSchedule(context: PluginContext, store: ReturnType<typeof createScheduleStore>) {
  requireProjectInput(context, {});
  const id = context.project.id;
  return {
    add: (functionId: string, expression: string, dueAt: number) => store.add(id, functionId, expression, dueAt),
    due: (now: number) => store.due(now).filter(schedule => schedule.projectId === id),
    claim: (functionId: string, workerId: string, now: number, leaseMs: number) => store.claim(id, functionId, workerId, now, leaseMs),
    recordOccurrence: (functionId: string, utc: string) => store.recordOccurrence(id, functionId, utc),
    enqueueLatestMissed: (functionId: string, now: number) => store.enqueueLatestMissed(id, functionId, now),
    enqueueVersion: (functionId: string, version: number, now: number) => store.enqueueVersion(id, functionId, version, now),
  };
}

export function bindProjectTriggers(context: PluginContext, store: ReturnType<typeof createRecordChangeTriggerStore>) {
  requireProjectInput(context, {});
  const id = context.project.id;
  return {
    add: (collection: string, operation: string, target: { functionId: string; version: number }) => store.add(id, collection, operation, target),
    match: (collection: string, operation: string) => store.match(id, collection, operation),
    dispatch: (collection: string, operation: string, committed: boolean, ancestry: readonly string[] = []) => store.dispatch(id, collection, operation, committed, ancestry),
  };
}

export function bindProjectJobLeases(context: PluginContext, store: ReturnType<typeof createJobLeaseStore>) {
  requireProjectInput(context, {});
  const id = context.project.id;
  return {
    acquire: (jobId: string, workerId: string, now: number, leaseMs: number) => store.acquire(id, jobId, workerId, now, leaseMs),
    renew: (jobId: string, workerId: string, now: number, leaseMs: number) => store.renew(id, jobId, workerId, now, leaseMs),
  };
}
