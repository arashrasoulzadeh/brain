const SESSION_KEY = "brain-admin-session";

export function readAdminSession(): string {
  return window.sessionStorage.getItem(SESSION_KEY) ?? "";
}

export function saveAdminSession(token: string): void {
  window.sessionStorage.setItem(SESSION_KEY, token);
}

export function clearAdminSession(): void {
  window.sessionStorage.removeItem(SESSION_KEY);
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
