import { projectUIURL } from "./admin-api.js";
import type { AdminPage, PluginStatus, SystemManifest } from "./admin-types.js";

export type AdminContribution = { projectID: string; pluginID: string; version: string; page: AdminPage; href: string; maxUploadBytes?: number };
export type AdminCatalog = { manifests: SystemManifest[] | null; installations: PluginStatus[] | null };

export function adminContributions(projectID: string, catalog: AdminCatalog): AdminContribution[] {
  if (!projectID) return [];
  const result: AdminContribution[] = [];
  for (const installation of catalog.installations ?? []) {
    if (installation.ProjectID !== projectID || !installation.Enabled || !installation.Ready || installation.Draining) continue;
    const manifest = catalog.manifests?.find(item => item.ID === installation.PluginID && item.Version === installation.Version);
    if (manifest?.Admin?.version !== 1 || !manifest.Capabilities.every(capability => installation.Grants?.includes(capability))) continue;
    for (const page of manifest.Admin.pages) {
      const actions = page.actions.filter(action => {
        const route = manifest.Routes?.find(item => item.Operation === action.operation);
        return route && installation.Grants?.includes(route.Capability);
      });
      if (!actions.length) continue;
      const href = projectUIURL(projectID, "extensions", manifest.ID, manifest.Version, page.id);
      result.push({ projectID, pluginID: manifest.ID, version: manifest.Version, page: { ...page, actions }, href, ...(manifest.ID === "files" && installation.MaxUploadBytes ? { maxUploadBytes: installation.MaxUploadBytes } : {}) });
    }
  }
  return result.sort((a, b) => a.page.title.localeCompare(b.page.title) || a.href.localeCompare(b.href));
}

export function adminPageExtensions(target: AdminContribution, contributions: readonly AdminContribution[]): AdminContribution[] {
  return contributions.filter(item => item.projectID === target.projectID && item.pluginID !== target.pluginID &&
    item.page.extends === target.pluginID + "/" + target.page.id);
}

export function adminLegacyContributions(projectID: string, catalog: AdminCatalog, route: string): AdminContribution[] {
  return adminContributions(projectID, catalog).filter(item => adminAliasMatches(item, route));
}

// Only a page that explicitly offers the tab slot can host another plugin's UI.
export function adminEmbeddedExtensions(target: AdminContribution, contributions: readonly AdminContribution[]): AdminContribution[] {
  return target.page.extension_slots?.includes("tab")
    ? adminPageExtensions(target, contributions).filter(item => item.page.extends_slot === "tab")
    : [];
}

export function adminLegacyRootPages(pages: readonly AdminContribution[]): AdminContribution[] {
  // Avoid hiding mutually extending pages or recursively nesting arbitrary plugins.
  return pages.filter(item => !pages.some(target => target.href !== item.href &&
    target.page.extends_slot !== "tab" &&
    adminEmbeddedExtensions(target, [item]).length > 0));
}

/** Keep the old combined room/matchmaking URL working without advertising it as a menu entry. */
export function adminAliasMatches(item: AdminContribution, route: string): boolean {
  return item.page.legacy_route === route || (route === "live-sessions" &&
    ((item.pluginID === "rooms" && item.page.legacy_route === "rooms") ||
     (item.pluginID === "matchmaking" && item.page.legacy_route === "matchmaking")));
}

export function adminAliasTitle(route: string): string {
  const names: Record<string, string> = { auth: "Authentication", pvp: "PvP", "access-management": "Access management", subscriptions: "Realtime subscriptions" };
  return Object.prototype.hasOwnProperty.call(names, route) ? names[route] :
    route ? route[0].toUpperCase() + route.slice(1).replaceAll("-", " ") : "Plugin administration";
}


// An enabled plugin without a contributed action page still has a dedicated
// project workspace: the host Dashboard or its exact Manage destination.
export function adminWorkspaceFallbacks(projectID: string, catalog: AdminCatalog, contributions: readonly AdminContribution[]) {
  return (catalog.installations ?? []).flatMap(item => {
    if (item.ProjectID !== projectID || !item.Enabled || !item.Ready || item.Draining) return [];
    const manifest = catalog.manifests?.find(m => m.ID === item.PluginID && m.Version === item.Version);
    if (!manifest || !manifest.Capabilities.every(capability => item.Grants?.includes(capability))) return [];
    if (item.PluginID === "functions" || contributions.some(page => page.projectID === projectID && page.pluginID === item.PluginID && page.version === item.Version)) return [];
    return [{ href: item.PluginID === "dashboard" ? projectUIURL(projectID, "dashboard") : projectUIURL(projectID, "plugins", item.PluginID + "@" + item.Version), label: manifest.Title || item.PluginID, icon: "resource" as const, detail: item.PluginID + "@" + item.Version }];
  }).filter((item, index, list) => list.findIndex(candidate => candidate.href === item.href) === index);
}
