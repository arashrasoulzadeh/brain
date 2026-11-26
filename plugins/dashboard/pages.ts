import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { AuthSettings, SignupSettings, FunctionSummary, WorkflowSummary, LiveSessionSummary } from "./types.js";

export type AdminPanelSection =
  "projects" | "plugins" | "database" | "functions" | "workflows" | "files" | "live-sessions";

export function adminPanel(projectId: string): {
  projectId: string;
  title: string;
  sections: readonly AdminPanelSection[];
  destructiveConfirmationRequired: true;
} {
  validateProjectId(projectId, "invalid admin panel project");
  return {
    projectId,
    title: "Brain administration",
    sections: [
      "projects",
      "plugins",
      "database",
      "functions",
      "workflows",
      "files",
      "live-sessions",
    ],
    destructiveConfirmationRequired: true,
  };
}

export function authSettingsPage(
  projectId: string,
  settings: AuthSettings,
): { projectId: string; title: string; settings: AuthSettings } {
  validateProjectId(projectId, "invalid auth settings");
  if (settings.oidcProviders.some((provider) => !provider))
    throw new Error("invalid auth settings");
  return {
    projectId,
    title: "Authentication settings",
    settings: { ...settings, oidcProviders: [...settings.oidcProviders] },
  };
}

export function signupSettingsPage(
  projectId: string,
  settings: SignupSettings,
): { projectId: string; title: string; settings: SignupSettings } {
  validateProjectId(projectId, "invalid signup settings");
  if (settings.allowedDomains.some((domain) => !domain || domain.includes("/")))
    throw new Error("invalid signup settings");
  return {
    projectId,
    title: "Signup settings",
    settings: { ...settings, allowedDomains: [...settings.allowedDomains] },
  };
}

export function recordBrowserPage(
  projectId: string,
  table: string,
  rows: readonly Record<string, unknown>[],
  approvedColumns: readonly string[],
): {
  projectId: string;
  table: string;
  columns: readonly string[];
  rows: Record<string, unknown>[];
} {
  validateProjectId(projectId, "invalid record browser");
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table) || !approvedColumns.length)
    throw new Error("invalid record browser");
  return {
    projectId,
    table,
    columns: [...approvedColumns],
    rows: rows.map((row) =>
      Object.fromEntries(
        approvedColumns
          .filter((column) => Object.hasOwn(row, column))
          .map((column) => [column, row[column]]),
      ),
    ),
  };
}

export function recordEditPage(
  projectId: string,
  table: string,
  recordId: string,
  expectedVersion: number,
  patch: Record<string, unknown>,
  approvedColumns: readonly string[],
): {
  projectId: string;
  table: string;
  recordId: string;
  expectedVersion: number;
  patch: Record<string, unknown>;
} {
  if (
    !recordId ||
    !Number.isInteger(expectedVersion) ||
    expectedVersion < 0 ||
    !approvedColumns.length
  )
    throw new Error("invalid record edit");
  validateProjectId(projectId, "invalid record edit");
  const filtered = Object.fromEntries(
    Object.entries(patch).filter(([column]) => approvedColumns.includes(column)),
  );
  return { projectId, table, recordId, expectedVersion, patch: filtered };
}

export function tableManagementPage(
  projectId: string,
  tables: readonly string[],
): { projectId: string; title: string; tables: readonly string[] } {
  validateProjectId(projectId, "invalid table management");
  if (tables.some((table) => !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)))
    throw new Error("invalid table management");
  return { projectId, title: "Table management", tables: [...tables] };
}

export function fileManagementPage(
  projectId: string,
  files: readonly { objectId: string; size: number; status: string }[],
): {
  projectId: string;
  title: string;
  files: readonly { objectId: string; size: number; status: string }[];
} {
  validateProjectId(projectId, "invalid file management");
  if (files.some((file) => !file.objectId || !Number.isSafeInteger(file.size) || file.size < 0))
    throw new Error("invalid file management");
  return { projectId, title: "File management", files: files.map((file) => ({ ...file })) };
}

export function functionManagementPage(projectId: string, functions: readonly FunctionSummary[]) {
  validateProjectId(projectId, "invalid function management");
  if (functions.some((item) => !item.id || !Number.isSafeInteger(item.version) || !item.status))
    throw new Error("invalid function management");
  return { projectId, title: "Functions", functions: functions.map((item) => ({ ...item })) };
}

export function workflowManagementPage(projectId: string, workflows: readonly WorkflowSummary[]) {
  validateProjectId(projectId, "invalid workflow management");
  if (
    workflows.some(
      (item) =>
        !item.id ||
        !item.status ||
        (item.lastRunAt !== undefined && !Number.isFinite(item.lastRunAt)),
    )
  )
    throw new Error("invalid workflow management");
  return { projectId, title: "Workflows", workflows: workflows.map((item) => ({ ...item })) };
}

export function liveSessionPage(projectId: string, sessions: readonly LiveSessionSummary[]) {
  validateProjectId(projectId, "invalid live sessions");
  if (
    sessions.some(
      (item) =>
        !item.id ||
        !item.kind ||
        !item.status ||
        !Number.isSafeInteger(item.participantCount) ||
        item.participantCount < 0,
    )
  )
    throw new Error("invalid live sessions");
  return { projectId, title: "Live sessions", sessions: sessions.map((item) => ({ ...item })) };
}

export function browserQuickstart(
  projectId: string,
  apiOrigin: string,
): { projectId: string; apiOrigin: string; steps: readonly string[] } {
  const origin = new URL(apiOrigin);
  validateProjectId(projectId, "invalid browser quickstart");
  if (!["http:", "https:"].includes(origin.protocol)) throw new Error("invalid browser quickstart");
  return {
    projectId,
    apiOrigin: origin.origin,
    steps: ["create-session", "load-project", "open-dashboard"],
  };
}
