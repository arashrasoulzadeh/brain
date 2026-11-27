import { adminHeaders, systemAdminURL } from "./admin-api.js";

export type AdminRequestIdentity = { path: string; token: string };

/** Ignore public plugin calls, foreign origins, login submissions and unrelated HTTP traffic. */
export function adminRequestIdentity(input: RequestInfo | URL, headers: Headers, origin: string): AdminRequestIdentity | undefined {
  const raw = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
  const url = new URL(raw, origin);
  if (url.origin !== origin || !/^\/(?:_system|[A-Za-z0-9][A-Za-z0-9_-]*)\/admin\/[a-z][a-z0-9/-]*$/.test(url.pathname)) return undefined;
  const authorization = headers.get("authorization") ?? "";
  if (!authorization.startsWith("Bearer ") || authorization.length <= 7) return undefined;
  return { path: url.pathname, token: authorization.slice(7) };
}

/** The legacy host uses 403 for both authentication and resource policy failures. */
export async function adminSessionRejected(identity: AdminRequestIdentity, status: number, request: typeof fetch): Promise<boolean> {
  if (status === 401) return true;
  if (status !== 403) return false;
  const probe = systemAdminURL("projects");
  if (identity.path === probe) return true;
  try {
    const response = await request(probe, { headers: adminHeaders(identity.token), cache: "no-store", signal: AbortSignal.timeout(5000) });
    return response.status === 401 || response.status === 403;
  } catch {
    // Network failures are not proof that the operator's session expired.
    return false;
  }
}
