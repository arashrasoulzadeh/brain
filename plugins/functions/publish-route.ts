export function publish(input: unknown): unknown {
  const payload = (input ?? {}) as { name?: unknown; source?: unknown; version?: unknown };
  const name = typeof payload.name === "string" ? payload.name.trim() : "";
  const source = typeof payload.source === "string" ? payload.source : "";
  const version = typeof payload.version === "string" ? payload.version.trim() : "";
  if (!name || !source || !version)
    throw new Error("function name, source, and version are required");
  return { name, version, published: true };
}
