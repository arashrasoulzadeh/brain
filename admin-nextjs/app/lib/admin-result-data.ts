export function readAdminResultPath(value: unknown, path?: string): unknown {
  if (!path) return value;
  for (const key of path.split(".")) {
    if (!key || key === "__proto__" || key === "constructor" || key === "prototype" || !value || typeof value !== "object" || !Object.hasOwn(value, key)) return undefined;
    value = (value as Record<string, unknown>)[key];
  }
  return value;
}

export function recordRows(value: unknown): Record<string, unknown>[] | undefined {
  return Array.isArray(value) && value.every(row => row && typeof row === "object" && !Array.isArray(row)) ? value as Record<string, unknown>[] : undefined;
}

export function displayAdminValue(value: unknown): string {
  return typeof value === "string" ? value : JSON.stringify(value) ?? "—";
}
