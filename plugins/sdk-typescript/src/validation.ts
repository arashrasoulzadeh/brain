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
