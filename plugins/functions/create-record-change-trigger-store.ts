import { createRecordTriggerState } from "./record-trigger-state.js";
import { RecordTriggerAdder } from "./record-trigger-adder.js";
import { RecordTriggerMatcher } from "./record-trigger-matcher.js";
import { RecordTriggerDispatcher } from "./record-trigger-dispatcher.js";
export function createRecordChangeTriggerStore() { const state = createRecordTriggerState(); const adder = new RecordTriggerAdder(state); const matcher = new RecordTriggerMatcher(state); const dispatcher = new RecordTriggerDispatcher(state); return { add: adder.add.bind(adder), match: matcher.match.bind(matcher), dispatch: dispatcher.dispatch.bind(dispatcher) }; }
