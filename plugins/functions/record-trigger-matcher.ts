import type { RecordChangeTrigger } from "./index.js";
import type { RecordTriggerState } from "./record-trigger-state.js";
export class RecordTriggerMatcher { constructor(private readonly state: RecordTriggerState) {} match(projectId: string, collection: string, operation: string): RecordChangeTrigger[] { return this.state.triggers.filter((trigger) => trigger.projectId === projectId && trigger.collection === collection && trigger.operation === operation).map((trigger) => ({ ...trigger, target: { ...trigger.target } })); } }
