const REQUIRED_STEPS = ["test", "build", "preflight", "deploy", "post-deploy"] as const;

export function validateDeploymentWorkflow(steps: readonly string[]): void {
  if (
    steps.length !== REQUIRED_STEPS.length ||
    steps.some((step, index) => step !== REQUIRED_STEPS[index])
  )
    throw new Error("invalid deployment workflow");
}
