type PublicCallInput = { name?: unknown; version?: unknown; input?: unknown; permission?: unknown };

export function call(input: unknown): unknown {
  const payload = (input ?? {}) as PublicCallInput;
  const name = typeof payload.name === "string" ? payload.name.trim() : "";
  const version = typeof payload.version === "string" ? payload.version.trim() : "latest";
  const permission = typeof payload.permission === "string" ? payload.permission.trim() : "public";
  if (!name) throw new Error("function name is required");
  if (permission !== "public") throw new Error("function is not public");
  return { name, version, input: payload.input ?? null, accepted: true, route: "public" };
}
