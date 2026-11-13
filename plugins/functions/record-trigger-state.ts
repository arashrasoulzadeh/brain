import type { RecordChangeTrigger } from "./index.js";
export type RecordTriggerState = { triggers: RecordChangeTrigger[] };
export function createRecordTriggerState(): RecordTriggerState { return { triggers: [] }; }
