export function customViewOrigin(origin: string, projectId: string) {
  const parsed = new URL(origin);
  if (parsed.protocol !== "https:" || !parsed.host || !projectId.trim())
    throw new Error("invalid custom view origin");
  return { origin: parsed.origin, projectId, isolated: true } as const;
}
