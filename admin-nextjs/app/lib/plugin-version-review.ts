import type { PluginManifest, PluginStatus } from "./admin-types.js";

export function pluginVersionKey(value: PluginManifest): string {
  return value.ID + "@" + value.Version;
}
export function installationVersionKey(value: PluginStatus): string {
  return value.PluginID + "@" + value.Version;
}
export function pluginCatalogState(installation: PluginStatus | undefined): "uninstalled" | "draining" | "not-ready" | "disabled" | "worker-unavailable" | "enabled" {
  if (!installation) return "uninstalled";
  if (installation.Draining) return "draining";
  if (!installation.Ready) return "not-ready";
  if (!installation.Enabled) return "disabled";
  return installation.WorkerReady === false ? "worker-unavailable" : "enabled";
}
export function pluginVersionReview(projectID: string, manifest: PluginManifest, manifests: readonly PluginManifest[], installations: readonly PluginStatus[]) {
  const scoped = installations.filter(item => item.ProjectID === projectID);
  const byKey = new Map(scoped.map(item => [installationVersionKey(item), item]));
  const manifestByKey = new Map(manifests.map(item => [pluginVersionKey(item), item]));
  const versions = manifests.filter(item => item.ID === manifest.ID).map(item => ({ manifest: item, installation: byKey.get(pluginVersionKey(item)) }));
  const selected = byKey.get(pluginVersionKey(manifest));
  const promoted = versions.find(item => item.installation?.Promoted);
  const otherEnabled = versions.filter(item => item.manifest.Version !== manifest.Version && item.installation?.Enabled);
  const previous = promoted?.manifest;
  const newlyRequested = previous && previous.Version !== manifest.Version ? (manifest.Capabilities ?? []).filter(capability => !previous.Capabilities.includes(capability)) : [];
  const missingApprovals = newlyRequested.filter(capability => !selected?.ApprovedCapabilities?.includes(capability));
  const missingGrants = (manifest.Capabilities ?? []).filter(capability => !selected?.Grants?.includes(capability));
  const missingConfig = Object.keys(manifest.ConfigSchema ?? {}).filter(name => !selected?.ConfiguredFields?.includes(name));
  const incompatibleConfig = previous && previous.Version !== manifest.Version ? Object.entries(previous.ConfigSchema ?? {}).filter(([name, type]) => manifest.ConfigSchema?.[name] !== type).map(([name]) => name) : [];
  const dependencies = [...(manifest.Dependencies ?? []).map(key => ({ key, required: true })), ...(manifest.OptionalDependencies ?? []).map(key => ({ key, required: false }))].map(dep => {
    const installation = byKey.get(dep.key);
    const registered = manifestByKey.get(dep.key);
    const active = scoped.find(item => item.PluginID === installation?.PluginID && item.Promoted);
    const state = pluginCatalogState(installation);
    const available = !!installation && !!registered && installation.Enabled && installation.Ready && !installation.Draining &&
      (!active || active.Version === installation.Version) && registered.Capabilities.every(capability => installation.Grants?.includes(capability));
    return { ...dep, state, available, registered: !!registered, selectedOtherVersion: !!active && active.Version !== installation?.Version };
  });
  const requiredUnavailable = dependencies.filter(item => item.required && !item.available);
  return { versions, selected, promoted, otherEnabled, newlyRequested, missingApprovals, missingGrants, missingConfig, incompatibleConfig, dependencies, requiredUnavailable };
}
