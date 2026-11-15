import { validateProjectId } from "../sdk-typescript/src/index.js";

const DEPLOYMENT_CAPABILITY = "deployment.execute";

export function authorizeDeploymentPlugin(
  grants: ReadonlySet<string>,
  projectId: string,
  pluginId: string,
  version: string,
): void {
  if (
    !pluginId.trim() ||
    !version.trim() ||
    !grants.has(`${projectId}:${pluginId}@${version}:${DEPLOYMENT_CAPABILITY}`)
  )
    throw new Error("deployment plugin grant required");
  validateProjectId(projectId, "deployment plugin grant required");
}
