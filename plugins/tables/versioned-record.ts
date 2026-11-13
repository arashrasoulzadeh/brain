export class VersionedRecord<T extends Record<string, unknown>> {
  private value: T;
  version = 0;
  constructor(value: T) { this.value = { ...value }; }
  update(expectedVersion: number, next: T): void { if (expectedVersion !== this.version) throw new Error("stale record version"); this.value = { ...next }; this.version += 1; }
  read(): { value: T; version: number } { return { value: { ...this.value }, version: this.version }; }
}
