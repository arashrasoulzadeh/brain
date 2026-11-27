export const functionMethods = ["GET", "POST", "PUT", "PATCH"] as const;
export type FunctionMethod = typeof functionMethods[number];
export type FunctionTemplate = "echo" | "health" | "transform";

export function functionTemplate(template: FunctionTemplate, method: FunctionMethod = "POST"): string {
  const handlers: Record<FunctionTemplate, string> = {
    echo: "  return { input };",
    health: "  return { status: 'ok' };",
    transform: "  return { items: Array.isArray(input) ? input : [] };",
  };
  return "export const config = { method: '" + method + "' };\n\nexport default async function handler(input) {\n" + handlers[template] + "\n}\n";
}

export function functionNameError(name: string): string {
  return /^[A-Za-z][A-Za-z0-9_-]{0,63}$/.test(name.trim()) ? "" : "Use 1–64 letters, digits, underscores or hyphens, starting with a letter.";
}

export function functionVersion(value: string): number | undefined {
  const version = Number(value);
  return Number.isSafeInteger(version) && version > 0 ? version : undefined;
}

export function functionPermissions(value: string): string[] {
  return [...new Set(value.split(",").map(permission => permission.trim()).filter(Boolean))];
}

export function functionInput(value: string): unknown {
  try { return JSON.parse(value) as unknown; }
  catch { throw new Error("Input must be valid JSON."); }
}
