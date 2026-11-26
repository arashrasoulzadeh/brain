import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";
import type { FunctionSchedule } from "./index.js";

type ScheduleLease = { workerId: string; expiresAt: number };

export function createScheduleStore() {
  const schedules: FunctionSchedule[] = [];
  const claims = createProjectMap<ScheduleLease>();
  const occurrences = createProjectMap<Set<string>>();
  const enqueued = createProjectMap<Set<number>>();

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
      const current = claims.get(projectId, functionId);
      if (current && current.expiresAt > now) return false;
      claims.set(projectId, functionId, { workerId, expiresAt: now + leaseMs });
      return true;
    },
    recordOccurrence(projectId: string, functionId: string, occurrenceUtc: string): boolean {
      validateProjectId(projectId);
      if (!functionId || !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/.test(occurrenceUtc) || Number.isNaN(Date.parse(occurrenceUtc)))
        throw new Error("invalid UTC occurrence");
      const seen = occurrences.get(projectId, functionId) ?? new Set<string>();
      if (seen.has(occurrenceUtc)) return false;
      seen.add(occurrenceUtc);
      occurrences.set(projectId, functionId, seen);
      return true;
    },
    enqueueLatestMissed(projectId: string, functionId: string, now: number) {
      validateProjectId(projectId);
      if (!Number.isSafeInteger(now) || now < 0) throw new Error("invalid schedule time");
      const schedule = schedules.filter((candidate) => candidate.projectId === projectId && candidate.functionId === functionId && candidate.dueAt <= now)
        .sort((left, right) => right.dueAt - left.dueAt)[0];
      if (!schedule) return undefined;
      const seen = enqueued.get(projectId, functionId) ?? new Set<number>();
      if (seen.has(schedule.dueAt)) return undefined;
      seen.add(schedule.dueAt);
      enqueued.set(projectId, functionId, seen);
      return { ...schedule };
    },
    enqueueVersion(projectId: string, functionId: string, functionVersion: number, enqueuedAt: number) {
      validateProjectId(projectId);
      if (!functionId || !Number.isSafeInteger(functionVersion) || functionVersion < 1 || !Number.isSafeInteger(enqueuedAt) || enqueuedAt < 0)
        throw new Error("invalid function enqueue");
      return { projectId, functionId, functionVersion, enqueuedAt };
    },
  };
}
