export type AdminSessionResponse = { session: string };
export type AdminAction = {
  operation: string;
  project_id?: string;
  plugin_id?: string;
  version?: string;
  capability?: string;
  resource?: string;
  key?: string;
  confirmed?: boolean;
};

export async function adminLogin(
  apiOrigin: string,
  username: string,
  secret: string,
  request: typeof fetch = fetch,
): Promise<AdminSessionResponse> {
  const response = await request(`${apiOrigin}/admin/session`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ username, secret }),
  });
  if (!response.ok) throw new Error("admin login failed");
  return (await response.json()) as AdminSessionResponse;
}

export async function adminAction(
  apiOrigin: string,
  session: string,
  action: AdminAction,
  request: typeof fetch = fetch,
): Promise<void> {
  const response = await request(`${apiOrigin}/admin/actions`, {
    method: "POST",
    headers: { authorization: `Bearer ${session}`, "content-type": "application/json" },
    body: JSON.stringify(action),
  });
  if (!response.ok) throw new Error("admin action failed");
}
