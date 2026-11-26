type PluginReadiness = { PluginID: string; Enabled: boolean; Ready: boolean };

export function enabledPluginIDs(installations: readonly PluginReadiness[]): ReadonlySet<string> {
  return new Set(installations.filter((item) => item.Enabled && item.Ready).map((item) => item.PluginID));
}

export function canViewAdminPage(page: string, enabled: ReadonlySet<string>): boolean {
  switch (page) {
    case "system": return enabled.has("dashboard");
    case "functions": return enabled.has("functions");
    case "routes":
    case "monitoring":
    case "workers": return enabled.size > 0;
    default: return true;
  }
}

export function visibleAdminPages(manifestCount: number, enabled: ReadonlySet<string>): string[] {
  return [
    "projects", ...(manifestCount > 0 ? ["plugins"] : []), "audit",
    ...["system", "functions", "routes", "monitoring", "workers"]
      .filter((page) => canViewAdminPage(page, enabled)),
  ];
}
