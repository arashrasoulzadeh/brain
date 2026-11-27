const SESSION_KEY = "brain-admin-session";
const PROJECT_KEY = "brain-admin-project";
export const ADMIN_PROJECT_CHANGED = "brain:project-changed";

// The sole URL contract for the Next.js control plane and project APIs.
export function systemAdminURL(resource: string): string {
  if (!/^[a-z][a-z0-9/-]*$/.test(resource) || resource.includes("..") || resource.includes("//")) {
    throw new Error("invalid admin resource");
  }
  return `/_system/admin/${resource}`;
}

export function projectAdminURL(project: string, resource: string): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(project) || project === "_system" ||
      !/^[a-z][a-z0-9/-]*$/.test(resource) || resource.includes("..") || resource.includes("//")) {
    throw new Error("invalid project API URL");
  }
  return `/${encodeURIComponent(project)}/admin/${resource}`;
}

export function selectedProjectAdminURL(resource: string): string {
  const project = adminUIContext(window.location.pathname).namespace ?? window.sessionStorage.getItem(PROJECT_KEY);
  if (!project) throw new Error("select a project before using its API");
  return projectAdminURL(project, resource);
}

export function readAdminSession(): string {
  return window.sessionStorage.getItem(SESSION_KEY) ?? "";
}

export function saveAdminSession(token: string): void {
  window.sessionStorage.setItem(SESSION_KEY, token);
}

export function clearAdminSession(): void {
  window.sessionStorage.removeItem(SESSION_KEY);
  window.sessionStorage.removeItem(PROJECT_KEY);
}

export function adminHeaders(token: string, extra: Record<string, string> = {}): Record<string, string> {
  return { ...extra, authorization: `Bearer ${token}` };
}

export async function revokeAdminSession(token: string, request: typeof fetch = fetch): Promise<void> {
  if (!token) return;
  await request(systemAdminURL("session"), { method: "DELETE", headers: adminHeaders(token) });
}

export async function signOutAdminSession(): Promise<void> {
  const token = readAdminSession();
  clearAdminSession();
  await revokeAdminSession(token).catch(() => undefined);
}

// The selected project follows the operator session across admin routes.
export function saveAdminProject(id: string): void {
  if ((window.sessionStorage.getItem(PROJECT_KEY) ?? "") === id) return;
  if (id) window.sessionStorage.setItem(PROJECT_KEY, id);
  else window.sessionStorage.removeItem(PROJECT_KEY);
  window.dispatchEvent(new Event(ADMIN_PROJECT_CHANGED));
}

export function preferredAdminProject<T extends { ID: string; Slug?: string; Active: boolean }>(projects: readonly T[], namespace?: string): T | undefined {
  if (namespace !== undefined) return projects.find(project => project.Active && (project.ID === namespace || project.Slug === namespace));
  const preferred = window.sessionStorage.getItem(PROJECT_KEY);
  return preferred ? projects.find(project => project.Active && project.ID === preferred) : projects.find(project => project.Active);
}

export function systemUIURL(page = ""): string {
  return "/_system/ui" + (page ? "/" + page : "");
}

export function projectUIURL(project: string, ...segments: string[]): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(project) ||
      segments.some(segment => !segment || segment === "." || segment === ".." || /[/\\]/.test(segment))) {
    throw new Error("invalid project UI path");
  }
  return "/" + encodeURIComponent(project) + "/ui" + (segments.length ? "/" + segments.map(encodeURIComponent).join("/") : "");
}

export function adminUIContext(pathname: string): { namespace?: string; page: string } {
  const [, namespace, ui, page] = pathname.split("/");
  if (ui !== "ui") return { page: "projects" };
  return namespace === "_system" ? { page: page || "projects" } : { namespace, page: page || "plugins" };
}
