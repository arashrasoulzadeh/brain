import { adminHeaders, projectAdminURL, readAdminSession, notifyPluginCatalogChanged } from "./admin-api.js";
import type { PluginManifest } from "./admin-types.js";

export async function installAndActivatePlugin(token: string, projectID: string, manifest: PluginManifest, request: typeof fetch = fetch): Promise<void> {
  if (!token || token !== readAdminSession()) throw new Error("Sign in again before installing");
  const response = await request(projectAdminURL(projectID, "actions"), {
    method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }),
    body: JSON.stringify({ operation: "plugin.install.activate", project_id: projectID, plugin_id: manifest.ID, version: manifest.Version, capabilities: manifest.Capabilities ?? [], confirmed: true }),
  });
  if (!response.ok) throw new Error(await response.text());
  notifyPluginCatalogChanged(projectID);
}
