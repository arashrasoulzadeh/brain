import { adminHeaders, preferredAdminProject, projectAdminURL, systemAdminURL } from "./admin-api.js";

export type FunctionDraft = { project_id: string; name: string; source: string; version?: number; public?: boolean; permissions?: string[] | null; method?: string; created_at: string; published_at?: string | null };
export type FunctionScope = { project: string; namespace: string; token: string };
export const MAX_SOURCE_BYTES = 256 * 1024;
export const MAX_INPUT_BYTES = 64 * 1024;
export const functionTemplates = {
  echo: { title: "JSON echo", source: "export const config = { method: 'POST' };\n\nexport default async function handler(input, context) {\n  return { input, project: context.project.id };\n}\n", input: '{\n  "message": "Hello"\n}' },
  health: { title: "Health endpoint", source: "export const config = { method: 'GET' };\n\nexport default function handler(input, context) {\n  return { ok: true, project: context.project.id };\n}\n", input: "null" },
  transform: { title: "Transform input", source: "export const config = { method: 'POST' };\n\nexport default function handler(input, context) {\n  return { name: context.helpers.call('string.trim', input.name) };\n}\n", input: '{\n  "name": "  Brain  "\n}' },
} as const;
export type FunctionTemplate = keyof typeof functionTemplates;
export function byteLength(value: string) { return new TextEncoder().encode(value).length; }
export function functionFormError(name: string, source: string, version: string): string {
  if (!/^[A-Za-z][A-Za-z0-9_-]{0,63}$/.test(name)) return "Name must start with a letter and contain at most 64 letters, numbers, underscores or hyphens.";
  if (!source.trim()) return "Function source is required.";
  if (byteLength(source) > MAX_SOURCE_BYTES) return "Source exceeds 256 KiB.";
  if (!/^[1-9][0-9]*$/.test(version) || !Number.isSafeInteger(Number(version))) return "Version must be a positive safe integer.";
  return "";
}
export function normalizePermissions(value: string): string[] {
  return [...new Set(value.split(",").map(item => item.trim()).filter(Boolean))].sort();
}
export function parseTestInput(value: string): unknown {
  if (byteLength(value) > MAX_INPUT_BYTES) throw new Error("Test input exceeds 64 KiB.");
  try { return JSON.parse(value); } catch { throw new Error("Test input must be valid JSON."); }
}
export function publicationURL(namespace: string, name: string): string {
  return "/" + encodeURIComponent(namespace) + "/functions/" + encodeURIComponent(name);
}
export function catalogStats(drafts: readonly FunctionDraft[]) {
  const published = drafts.filter(draft => Boolean(draft.published_at)).length;
  return { total: drafts.length, published, unpublished: drafts.length - published, public: drafts.filter(draft => Boolean(draft.public)).length };
}
export function formatFunctionDate(value?: string | null): string {
  if (!value) return "—";
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? "—" : parsed.toLocaleString();
}
export async function resolveFunctionScope(token: string, namespace: string | undefined, signal: AbortSignal): Promise<FunctionScope> {
  if (!token || !namespace) throw new Error("Select an active project and sign in.");
  const response = await fetch(systemAdminURL("projects"), { headers: adminHeaders(token), signal });
  if (!response.ok) throw new Error("Projects unavailable.");
  const projects = (await response.json() as { ID: string; Slug?: string; Active: boolean }[] | null) ?? [];
  const selected = preferredAdminProject(projects, namespace);
  if (!selected) throw new Error("Project unavailable.");
  return { project: selected.ID, namespace: selected.Slug || selected.ID, token };
}
export async function loadFunctionDrafts(scope: FunctionScope, signal: AbortSignal): Promise<FunctionDraft[]> {
  const response = await fetch(projectAdminURL(scope.project, "functions"), { headers: adminHeaders(scope.token), signal });
  if (!response.ok) throw new Error("Functions unavailable (" + response.status + ").");
  const rows = (await response.json() as FunctionDraft[] | null) ?? [];
  return rows.filter(row => row.project_id === scope.project);
}
export type FunctionCommand = { operation: "functions.create" | "functions.update" | "functions.permissions" | "functions.test" | "functions.run" | "functions.publish"; name: string; source?: string; function_version?: number; permissions?: string[]; public?: boolean; input?: unknown };
export async function sendFunctionCommand(scope: FunctionScope, command: FunctionCommand, signal: AbortSignal): Promise<unknown> {
  const response = await fetch(projectAdminURL(scope.project, "functions"), {
    method: "POST", headers: adminHeaders(scope.token, { "content-type": "application/json" }), signal,
    body: JSON.stringify({ ...command, project_id: scope.project }),
  });
  const body = await response.text();
  if (!response.ok) throw new Error(body || "Request rejected (" + response.status + ").");
  if (response.status === 204) return null;
  if (command.operation === "functions.permissions") throw new Error("Settings update was not confirmed. Refresh before retrying.");
  try { return JSON.parse(body); } catch { throw new Error("Invalid function API response."); }
}
export function checkedFunctionDraft(value: unknown, scope: FunctionScope, name: string): FunctionDraft {
  const row = value as Partial<FunctionDraft> | null;
  if (!row || row.project_id !== scope.project || row.name !== name || typeof row.source !== "string" || typeof row.version !== "number") throw new Error("Function response does not match the requested project and function.");
  return row as FunctionDraft;
}
export function checkedTestResult(value: unknown, scope: FunctionScope, name: string, run: boolean): unknown {
  const row = value as { valid?: boolean; project_id?: string; name?: string; result?: unknown } | null;
  if (!row || row.valid !== true || row.project_id !== scope.project || row.name !== name || (run && !Object.hasOwn(row, "result"))) throw new Error("Host did not confirm the requested function " + (run ? "execution." : "validation."));
  return run ? row.result : true;
}
export function checkedPublication(value: unknown, scope: FunctionScope, name: string): { version: number; method: string; published_at: string } {
  const row = value as { published?: boolean; project_id?: string; name?: string; version?: number; method?: string; published_at?: string } | null;
  if (!row || row.published !== true || row.project_id !== scope.project || row.name !== name || typeof row.version !== "number" || !row.method || !row.published_at) throw new Error("Publication response was not confirmed. Refresh before retrying.");
  return { version: row.version, method: row.method, published_at: row.published_at };
}
