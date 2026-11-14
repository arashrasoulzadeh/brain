import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { ScheduleState } from "./schedule-state.js";
export class ScheduleAdder {
  constructor(private readonly state: ScheduleState) {}
  add(projectId: string, functionId: string, expression: string, dueAt: number) {
    validateProjectId(projectId);
    if (
      !functionId ||
      expression.trim().split(/\s+/).length !== 5 ||
      !Number.isSafeInteger(dueAt) ||
      dueAt < 0
    )
      throw new Error("invalid schedule");
    const schedule = { projectId, functionId, expression: expression.trim(), dueAt };
    this.state.schedules.push(schedule);
    return { ...schedule };
  }
}
