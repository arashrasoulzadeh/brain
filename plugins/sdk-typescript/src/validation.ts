import type { PluginContext } from "./index.js";

/** Only use with context injected by the trusted worker; payload identity is never authoritative. */
export function requireProjectInput(context: PluginContext, input: unknown): Record<string, unknown> {
  if (!context?.project?.id || context.project.active !== true)
    throw new Error("host project context required");
  validateProjectId(context.project.id);
  if (!input || typeof input !== "object" || Array.isArray(input))
    throw new Error("invalid plugin input");
  const payload = input as Record<string, unknown>;
  for (const key of ["projectId", "project_id"]) {
    if (Object.hasOwn(payload, key) && payload[key] !== context.project.id)
      throw new Error("host project context mismatch");
  }
  return payload;
}

export function validateProjectId(projectId: string, errorMessage = "invalid project id"): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error(errorMessage);
  return projectId;
}

export function validateIdentifier(
  identifier: string,
  errorMessage = "invalid identifier",
): string {
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(identifier)) throw new Error(errorMessage);
  return identifier;
}

export function trimmedString(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value.trim() : fallback;
}

export function requiredString(value: unknown, message: string): string {
  const result = trimmedString(value);
  if (!result) throw new Error(message);
  return result;
}

export function validateBoundedInteger(value: number, maximum: number, error: string, minimumMaximum = 0): void {
  if (!Number.isSafeInteger(value) || !Number.isSafeInteger(maximum) || value < 0 || maximum < minimumMaximum || value > maximum)
    throw new Error(error);
}

export function requireAllowedValues(values: readonly string[], allowed: readonly string[], error: string): void {
  if (values.some((value) => !allowed.includes(value))) throw new Error(error);
}

export function projectRoomID(projectId: string, roomId: string): string {
  validateProjectId(projectId);
  validateProjectId(roomId, "invalid project room id");
  return `${projectId}:${roomId}`;
}

/** Require the effective host-verified operator and the admitted operation scope. */
export function requireOperator(context: PluginContext | undefined, capability: string, resource = ""): PluginContext {
  if (!context?.auth || context.auth.caller.kind !== "operator" || !context.auth.caller.authenticated ||
      !context.auth.can(capability, resource)) throw new Error("operator authorization required");
  return context;
}
