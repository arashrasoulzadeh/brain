import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { ScheduleState } from "./schedule-state.js";
export class MissedScheduleEnqueuer {
  constructor(private readonly state: ScheduleState) {}
  enqueueLatestMissed(projectId: string, functionId: string, now: number) {
    validateProjectId(projectId);
    if (!Number.isSafeInteger(now) || now < 0) throw new Error("invalid schedule time");
    const schedule = this.state.schedules
      .filter(
        (candidate) =>
          candidate.projectId === projectId &&
          candidate.functionId === functionId &&
          candidate.dueAt <= now,
      )
      .sort((left, right) => right.dueAt - left.dueAt)[0];
    if (!schedule) return undefined;
    const key = `${projectId}:${functionId}:${schedule.dueAt}`;
    if (this.state.enqueued.has(key)) return undefined;
    this.state.enqueued.add(key);
    return { ...schedule };
  }
}
