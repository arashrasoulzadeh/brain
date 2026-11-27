import { adminHeaders } from "./admin-api";
import type { AdminContribution } from "./admin-registry";
import type { AdminInventoryPage } from "./admin-types";

export async function invokeAdminAction(token: string, contribution: AdminContribution, operation: string, payload: unknown, signal: AbortSignal, query = false): Promise<unknown> {
  const action = contribution.page.actions.find(item => item.operation === operation);
  if (!action || (query && !action.read_only)) throw new Error("Action unavailable");
  const response = await fetch("/api/admin/actions", {
    method: "POST",
    headers: adminHeaders(token, { "content-type": "application/json" }),
    signal,
    body: JSON.stringify({ operation: query ? "plugin.admin.query" : "plugin.admin.invoke", project_id: contribution.projectID,
      plugin_id: contribution.pluginID, version: contribution.version, page_id: contribution.page.id,
      target_operation: operation, payload, confirmed: !query }),
  });
  if (!response.ok) throw new Error(await response.text());
  const result = await response.json() as { Success: boolean; Payload?: string; ErrorCode?: string };
  if (!result.Success) throw new Error(result.ErrorCode || "Plugin operation failed");
  if (!result.Payload) return null;
  const bytes = Uint8Array.from(atob(result.Payload), char => char.charCodeAt(0));
  return JSON.parse(new TextDecoder().decode(bytes));
}

export async function adminPageRevision(token: string, contribution: AdminContribution, signal: AbortSignal, scope?: { operation: string; payload: unknown }): Promise<number> {
  const response = await fetch("/api/admin/actions", {
    method: "POST",
    headers: adminHeaders(token, { "content-type": "application/json" }),
    signal,
    body: JSON.stringify({ operation: "plugin.admin.revision", project_id: contribution.projectID,
      plugin_id: contribution.pluginID, version: contribution.version, page_id: contribution.page.id,
      target_operation: scope?.operation, payload: scope?.payload }),
  });
  if (!response.ok) throw new Error(await response.text());
  const result = await response.json() as { revision: number };
  if (!Number.isSafeInteger(result.revision)) throw new Error("Invalid observation revision");
  return result.revision;
}

export async function fetchAdminInventory(token: string, contribution: AdminContribution, after: string, signal: AbortSignal, quiet = false): Promise<AdminInventoryPage> {
  const response = await fetch("/api/admin/actions", {
    method: "POST",
    headers: adminHeaders(token, { "content-type": "application/json", ...(quiet ? { "x-brain-quiet": "1" } : {}) }),
    signal,
    body: JSON.stringify({ operation: "plugin.admin.inventory", project_id: contribution.projectID,
      plugin_id: contribution.pluginID, version: contribution.version, page_id: contribution.page.id, after, limit: 50 }),
  });
  if (!response.ok) throw new Error(await response.text());
  const result = await response.json() as AdminInventoryPage;
  if (!Array.isArray(result.rows) || result.rows.some(row => typeof row.key !== "string") ||
      (result.next !== undefined && typeof result.next !== "string")) throw new Error("Invalid inventory response");
  return result;
}
