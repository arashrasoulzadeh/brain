type PluginReadiness = { PluginID: string; Enabled: boolean; Ready: boolean; Draining?: boolean };

export function enabledPluginIDs(installations: readonly PluginReadiness[]): ReadonlySet<string> {
  return new Set(installations.filter((item) => item.Enabled && item.Ready && !item.Draining).map((item) => item.PluginID));
}

export function canViewAdminPage(page: string, enabled: ReadonlySet<string>): boolean {
  switch (page) {
    case "system": return enabled.has("dashboard");
    case "functions": return enabled.has("functions");
    case "routes":
    case "monitoring": return enabled.size > 0;
    case "workers": return true;
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
