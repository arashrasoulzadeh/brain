export function pluginHealthContribution(projectId: string) {
  if (!projectId.trim()) throw new Error("project identity required");
  return { view: "plugin.health", projectId, title: "Plugin health", replace: false } as const;
}
