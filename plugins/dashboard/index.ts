import {
  request,
  type RpcRequest,
} from "../sdk-typescript/src/index.js";

export function projectCreateRequest(pluginId: string, version: string, id: string): RpcRequest {
  return request(`dashboard-project-${id}`, {
    protocol: 1,
    projectId: "operator",
    pluginId,
    version,
    operation: "project.create",
    payload: { id },
  });
}

export function adminRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: string,
  payload: unknown,
): RpcRequest {
  return request(`dashboard-${operation}-${projectId}`, {
    protocol: 1,
    projectId,
    pluginId,
    version,
    operation,
    payload,
  });
}

export type AuthSettings = { passwordEnabled: boolean; oidcProviders: readonly string[]; emailRecoveryEnabled: boolean };
export function authSettingsPage(projectId: string, settings: AuthSettings): { projectId: string; title: string; settings: AuthSettings } {
  if (!projectId || settings.oidcProviders.some((provider) => !provider)) throw new Error("invalid auth settings");
  return { projectId, title: "Authentication settings", settings: { ...settings, oidcProviders: [...settings.oidcProviders] } };
}
export type SignupSettings = { enabled: boolean; allowedDomains: readonly string[]; requireInvite: boolean };
export function signupSettingsPage(projectId: string, settings: SignupSettings): { projectId: string; title: string; settings: SignupSettings } {
  if (!projectId || settings.allowedDomains.some((domain) => !domain || domain.includes("/"))) throw new Error("invalid signup settings");
  return { projectId, title: "Signup settings", settings: { ...settings, allowedDomains: [...settings.allowedDomains] } };
}
export function recordBrowserPage(projectId: string, table: string, rows: readonly Record<string, unknown>[], approvedColumns: readonly string[]): { projectId: string; table: string; columns: readonly string[]; rows: Record<string, unknown>[] } {
  if (!projectId || !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !approvedColumns.length) throw new Error("invalid record browser");
  return { projectId, table, columns: [...approvedColumns], rows: rows.map((row) => Object.fromEntries(approvedColumns.filter((column) => Object.hasOwn(row, column)).map((column) => [column, row[column]]))) };
}
export function recordEditPage(projectId: string, table: string, recordId: string, expectedVersion: number, patch: Record<string, unknown>, approvedColumns: readonly string[]): { projectId: string; table: string; recordId: string; expectedVersion: number; patch: Record<string, unknown> } {
  if (!projectId || !recordId || !Number.isInteger(expectedVersion) || expectedVersion < 0 || !approvedColumns.length) throw new Error("invalid record edit");
  const filtered = Object.fromEntries(Object.entries(patch).filter(([column]) => approvedColumns.includes(column)));
  return { projectId, table, recordId, expectedVersion, patch: filtered };
}
export function tableManagementPage(projectId: string, tables: readonly string[]): { projectId: string; title: string; tables: readonly string[] } {
  if (!projectId || tables.some((table) => !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table))) throw new Error("invalid table management");
  return { projectId, title: "Table management", tables: [...tables] };
}
export function fileManagementPage(projectId: string, files: readonly { objectId: string; size: number; status: string }[]): { projectId: string; title: string; files: readonly { objectId: string; size: number; status: string }[] } {
  if (!projectId || files.some((file) => !file.objectId || !Number.isSafeInteger(file.size) || file.size < 0)) throw new Error("invalid file management");
  return { projectId, title: "File management", files: files.map((file) => ({ ...file })) };
}
export type FormAudit = { projectId: string; actorId: string; action: string; occurredAt: number };
export class AuditedFormActions {
  private readonly entries: FormAudit[] = [];
  record(projectId: string, actorId: string, action: string, occurredAt: number): void { if (!projectId || !actorId || !action || !Number.isFinite(occurredAt)) throw new Error("invalid form audit"); this.entries.push({ projectId, actorId, action, occurredAt }); }
  list(projectId: string): FormAudit[] { return this.entries.filter((entry) => entry.projectId === projectId).map((entry) => ({ ...entry })); }
}

