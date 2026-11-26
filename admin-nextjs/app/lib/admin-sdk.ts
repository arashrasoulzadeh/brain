import { adminHeaders } from "./admin-api";
import type { AdminContribution } from "./admin-registry";

export async function invokeAdminAction(token: string, contribution: AdminContribution, operation: string, payload: unknown, signal: AbortSignal): Promise<unknown> {
  if (!contribution.page.actions.some(action => action.operation === operation)) throw new Error("Action unavailable");
  const response = await fetch("/api/admin/actions", {
    method: "POST",
    headers: adminHeaders(token, { "content-type": "application/json" }),
    signal,
    body: JSON.stringify({ operation: "plugin.admin.invoke", project_id: contribution.projectID,
      plugin_id: contribution.pluginID, version: contribution.version, page_id: contribution.page.id,
      target_operation: operation, payload, confirmed: true }),
  });
  if (!response.ok) throw new Error(await response.text());
  const result = await response.json() as { Success: boolean; Payload?: string; ErrorCode?: string };
  if (!result.Success) throw new Error(result.ErrorCode || "Plugin operation failed");
  if (!result.Payload) return null;
  const bytes = Uint8Array.from(atob(result.Payload), char => char.charCodeAt(0));
  return JSON.parse(new TextDecoder().decode(bytes));
}
