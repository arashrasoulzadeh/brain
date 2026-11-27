import type { PluginContext } from "../sdk-typescript/src/index.js";

/** Rules live in the host-owned, project-scoped model, not in worker memory. */
export function describe(context: PluginContext) {
  return { projectId: context.project.id, policy: "deny-by-default", scope: "invocation" };
}
