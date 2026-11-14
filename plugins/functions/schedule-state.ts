import type { FunctionSchedule } from "./index.js";
export type ScheduleState = {
  schedules: FunctionSchedule[];
  claims: Map<string, { workerId: string; expiresAt: number }>;
  occurrences: Set<string>;
  enqueued: Set<string>;
};
export function createScheduleState(): ScheduleState {
  return { schedules: [], claims: new Map(), occurrences: new Set(), enqueued: new Set() };
}
