import { validateDeploymentWorkflow } from "../../plugins/functions/index.js";

validateDeploymentWorkflow(["test", "build", "preflight", "deploy", "post-deploy"]);
for (const steps of [
  ["build", "test", "preflight", "deploy", "post-deploy"],
  ["test", "build", "deploy"],
  ["test", "build", "preflight", "deploy", "deploy", "post-deploy"],
  ["test", "build", "preflight", "deploy", "shell"],
] as const) {
  try {
    validateDeploymentWorkflow(steps);
    throw new Error("invalid deployment workflow was accepted");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "invalid deployment workflow") throw error;
  }
}
