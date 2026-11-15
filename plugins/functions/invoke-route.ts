export function invoke(input: unknown): unknown {
  const payload = (input ?? {}) as { name?: unknown; input?: unknown };
  const name = typeof payload.name === "string" ? payload.name.trim() : "";
  if (!name) throw new Error("function name is required");
  return { name, input: payload.input ?? null, invoked: true };
}
