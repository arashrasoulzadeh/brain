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
