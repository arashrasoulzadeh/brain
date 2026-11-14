export function requireRecentAuthentication(
  authenticatedAt: number,
  now: number,
  maxAgeMs: number,
): void {
  if (maxAgeMs <= 0 || now - authenticatedAt > maxAgeMs)
    throw new Error("recent authentication required");
}
