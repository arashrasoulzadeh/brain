import { validateProjectId } from "../sdk-typescript/src/index.js";

export function pluginHealthContribution(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return { view: "plugin.health", projectId, title: "Plugin health", replace: false } as const;
}
