import type { VersionedRecordState } from "./versioned-record-state.js";
export class VersionedRecordReader<T extends Record<string, unknown>> {
  constructor(private readonly state: VersionedRecordState<T>) {}
  read(): { value: T; version: number } {
    return { value: { ...this.state.value }, version: this.state.version };
  }
}
