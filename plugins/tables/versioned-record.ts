export function createVersionedRecord<T extends Record<string, unknown>>(value: T) {
  let current = { ...value };
  let version = 0;
  return {
    get version() { return version; },
    read(): { value: T; version: number } {
      return { value: { ...current }, version };
    },
    update(expectedVersion: number, next: T): void {
      if (expectedVersion !== version) throw new Error("stale record version");
      current = { ...next };
      version += 1;
    },
  };
}
