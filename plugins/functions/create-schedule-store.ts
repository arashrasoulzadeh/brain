import { createScheduleState } from "./schedule-state.js";
import { ScheduleAdder } from "./schedule-adder.js";
import { ScheduleReader } from "./schedule-reader.js";
import { ScheduleClaimer } from "./schedule-claimer.js";
import { ScheduleOccurrenceRecorder } from "./schedule-occurrence-recorder.js";
import { MissedScheduleEnqueuer } from "./missed-schedule-enqueuer.js";
import { FunctionVersionEnqueuer } from "./function-version-enqueuer.js";
export function createScheduleStore() {
  const state = createScheduleState();
  const adder = new ScheduleAdder(state);
  const reader = new ScheduleReader(state);
  const claimer = new ScheduleClaimer(state);
  const occurrence = new ScheduleOccurrenceRecorder(state);
  const missed = new MissedScheduleEnqueuer(state);
  const version = new FunctionVersionEnqueuer();
  return {
    add: adder.add.bind(adder),
    due: reader.due.bind(reader),
    claim: claimer.claim.bind(claimer),
    recordOccurrence: occurrence.recordOccurrence.bind(occurrence),
    enqueueLatestMissed: missed.enqueueLatestMissed.bind(missed),
    enqueueVersion: version.enqueueVersion.bind(version),
  };
}
