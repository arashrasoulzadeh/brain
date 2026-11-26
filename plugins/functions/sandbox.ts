import { validateProjectId } from "../sdk-typescript/src/index.js";

const CREDENTIAL_KEY = /token|secret|password|credential|private[_-]?key|api[_-]?key/i;

export function rejectPluginCredentials(value: unknown): void {
  if (Array.isArray(value)) {
    for (const entry of value) rejectPluginCredentials(entry);
    return;
  }
  if (!value || typeof value !== "object") return;
  for (const [key, entry] of Object.entries(value)) {
    if (CREDENTIAL_KEY.test(key)) throw new Error("publish credentials must remain host-owned");
    rejectPluginCredentials(entry);
  }
}

const HOST_RESOURCE_KEY = /docker|socket|containerd|host[_-]?path/i;

export function rejectPluginHostResources(value: unknown): void {
  if (Array.isArray(value)) {
    for (const entry of value) rejectPluginHostResources(entry);
    return;
  }
  if (!value || typeof value !== "object") return;
  for (const [key, entry] of Object.entries(value)) {
    if (
      HOST_RESOURCE_KEY.test(key) ||
      (typeof entry === "string" && (entry.includes("/var/run/") || entry.includes("/run/")))
    )
      throw new Error("host resources must remain host-owned");
    rejectPluginHostResources(entry);
  }
}

export function rejectSandboxEscapeAttempt(value: string): void {
  let decoded = value;
  try {
    decoded = decodeURIComponent(value);
  } catch {
    throw new Error("sandbox escape attempt rejected");
  }
  const normalized = decoded.replaceAll("\\", "/");
  if (
    !value ||
    decoded.includes("\u0000") ||
    normalized.split("/").includes("..") ||
    normalized.startsWith("/") ||
    /^[A-Za-z]:\//.test(normalized) ||
    /^https?:\/\//.test(decoded)
  )
    throw new Error("sandbox escape attempt rejected");
}

export function validateSandboxCapabilities(capabilities: readonly string[]): void {
  const denied = new Set(["filesystem", "network", "environment-credentials"]);
  if (capabilities.some((capability) => denied.has(capability)))
    throw new Error("sandbox capability denied");
}

export function validateSourceProject(sourceProjectId: string, requestedProjectId: string): void {
  if (!sourceProjectId || sourceProjectId !== requestedProjectId)
    throw new Error("cross-project source access denied");
}

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

const REQUIRED_STEPS = ["test", "build", "preflight", "deploy", "post-deploy"] as const;

export function validateDeploymentWorkflow(steps: readonly string[]): void {
  if (
    steps.length !== REQUIRED_STEPS.length ||
    steps.some((step, index) => step !== REQUIRED_STEPS[index])
  )
    throw new Error("invalid deployment workflow");
}

export type DeploymentHook = () => void | Promise<void>;
export type DeploymentHookSet = {
  test: DeploymentHook;
  build: DeploymentHook;
  preflight: DeploymentHook;
  postDeploy: DeploymentHook;
};

export function createDeploymentHookSet(hooks: Partial<DeploymentHookSet>): DeploymentHookSet {
  if (
    typeof hooks.test !== "function" ||
    typeof hooks.build !== "function" ||
    typeof hooks.preflight !== "function" ||
    typeof hooks.postDeploy !== "function"
  )
    throw new Error("all deployment hooks required");
  return {
    test: hooks.test,
    build: hooks.build,
    preflight: hooks.preflight,
    postDeploy: hooks.postDeploy,
  };
}
