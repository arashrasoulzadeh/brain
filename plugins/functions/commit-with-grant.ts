export function commitWithGrant<T>(
  hasGrant: () => boolean,
  commit: () => T,
  capability: string,
): T {
  if (!capability || !hasGrant()) throw new Error("grant revoked before commit");
  return commit();
}
