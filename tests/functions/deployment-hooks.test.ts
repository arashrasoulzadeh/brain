import { createDeploymentHookSet } from "../../plugins/functions/index.js";

const calls: string[] = [];
const hooks = createDeploymentHookSet({
  test: () => {
    calls.push("test");
  },
  build: () => {
    calls.push("build");
  },
  preflight: () => {
    calls.push("preflight");
  },
  postDeploy: () => {
    calls.push("post-deploy");
  },
});
hooks.test();
hooks.build();
hooks.preflight();
hooks.postDeploy();
if (calls.join(",") !== "test,build,preflight,post-deploy")
  throw new Error("deployment hooks did not run in order");

try {
  createDeploymentHookSet({ test: () => undefined } as never);
  throw new Error("incomplete deployment hooks were accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "all deployment hooks required") throw error;
}
