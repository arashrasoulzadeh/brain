import { createVersionedRecordState } from "./versioned-record-state.js";
import { VersionedRecordReader } from "./versioned-record-reader.js";
import { VersionedRecordUpdater } from "./versioned-record-updater.js";
export function createVersionedRecord<T extends Record<string, unknown>>(value: T) { const state = createVersionedRecordState(value); const reader = new VersionedRecordReader(state); const updater = new VersionedRecordUpdater(state); return { get version() { return state.version; }, read: reader.read.bind(reader), update: updater.update.bind(updater) }; }
