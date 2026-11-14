import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { RecordChangeTrigger } from "./index.js";
import type { RecordTriggerState } from "./record-trigger-state.js";
export class RecordTriggerAdder {
  constructor(private readonly state: RecordTriggerState) {}
  add(
    projectId: string,
    collection: string,
    operation: string,
    target: { functionId: string; version: number },
  ): RecordChangeTrigger {
    validateProjectId(projectId);
    if (
      !collection ||
      !operation ||
      !target.functionId ||
      !Number.isSafeInteger(target.version) ||
      target.version < 1
    )
      throw new Error("invalid record trigger");
    const trigger = { projectId, collection, operation, target: { ...target } };
    this.state.triggers.push(trigger);
    return { ...trigger, target: { ...trigger.target } };
  }
}
