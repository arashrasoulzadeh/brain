import type { RecordChangeTrigger } from "./index.js";
import type { RecordTriggerState } from "./record-trigger-state.js";
import { RecordTriggerMatcher } from "./record-trigger-matcher.js";
export class RecordTriggerDispatcher { private readonly matcher: RecordTriggerMatcher; constructor(private readonly state: RecordTriggerState) { this.matcher = new RecordTriggerMatcher(state); } dispatch(projectId: string, collection: string, operation: string, committed: boolean, ancestry: readonly string[] = []): RecordChangeTrigger[] { return committed ? this.matcher.match(projectId, collection, operation).filter((trigger) => !ancestry.includes(trigger.target.functionId)) : []; } }
