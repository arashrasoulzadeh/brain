import { projectUIURL } from "./admin-api.js";
import type { AdminPage, PluginStatus, SystemManifest } from "./admin-types.js";

export type AdminContribution = { projectID: string; pluginID: string; version: string; page: AdminPage; href: string };
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
      result.push({ projectID, pluginID: manifest.ID, version: manifest.Version, page: { ...page, actions }, href });
    }
  }
  return result.sort((a, b) => a.page.title.localeCompare(b.page.title) || a.href.localeCompare(b.href));
}

export function adminPageExtensions(target: AdminContribution, contributions: readonly AdminContribution[]): AdminContribution[] {
  return contributions.filter(item => item.projectID === target.projectID && item.pluginID !== target.pluginID &&
    item.page.extends === target.pluginID + "/" + target.page.id);
}

export function adminLegacyContributions(projectID: string, catalog: AdminCatalog, route: string): AdminContribution[] {
  return adminContributions(projectID, catalog).filter(item => item.page.legacy_route === route);
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
