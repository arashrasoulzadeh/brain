import type { ScheduleState } from "./schedule-state.js";
export class ScheduleReader {
  constructor(private readonly state: ScheduleState) {}
  due(now: number) {
    return this.state.schedules
      .filter((schedule) => schedule.dueAt <= now)
      .map((schedule) => ({ ...schedule }));
  }
}
