const SESSION_KEY = "brain-admin-session";
const PROJECT_KEY = "brain-admin-project";
export const ADMIN_PROJECT_CHANGED = "brain:project-changed";

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
  await request("/api/admin/session", { method: "DELETE", headers: adminHeaders(token) });
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

export function preferredAdminProject<T extends { ID: string; Active: boolean }>(projects: readonly T[]): T | undefined {
  const preferred = window.sessionStorage.getItem(PROJECT_KEY);
  return preferred ? projects.find(project => project.Active && project.ID === preferred) : projects.find(project => project.Active);
}
