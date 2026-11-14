export type VersionedRecordState<T extends Record<string, unknown>> = { value: T; version: number };
export function createVersionedRecordState<T extends Record<string, unknown>>(
  value: T,
): VersionedRecordState<T> {
  return { value: { ...value }, version: 0 };
}
