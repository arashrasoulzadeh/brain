import {
  createApprovalStore,
  createRecordChangeTriggerStore,
  evaluateCondition,
  authorizeDeploymentPlugin,
  registerProjectWorkflowAction,
} from "../../plugins/functions/index.js";

const triggers = createRecordChangeTriggerStore();
triggers.add("dev", "profiles", "profile.changed", { functionId: "notify-owner", version: 1 });
const matchedTriggers = triggers.match("dev", "profiles", "profile.changed");
if (matchedTriggers.length !== 1 || matchedTriggers[0]?.target.functionId !== "notify-owner")
  throw new Error("named workflow trigger was not registered");

if (
  !evaluateCondition({ field: "status", operator: "equals", value: "ready" }, { status: "ready" })
)
  throw new Error("named workflow condition did not match");

const actions = new Map();
registerProjectWorkflowAction(actions, {
  projectId: "dev",
  id: "notify-owner",
  operation: "notify.owner",
  capability: "workflow.project",
});
if (!actions.has("dev:notify-owner")) throw new Error("named workflow action was not registered");

const deploymentGrants = new Set(["dev:deploy-plugin@1.0.0:deployment.execute"]);
authorizeDeploymentPlugin(deploymentGrants, "dev", "deploy-plugin", "1.0.0");
deploymentGrants.delete("dev:deploy-plugin@1.0.0:deployment.execute");
try {
  authorizeDeploymentPlugin(deploymentGrants, "dev", "deploy-plugin", "1.0.0");
  throw new Error("deployment plugin without grant was accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "deployment plugin grant required")
    throw error;
}

const approvals = createApprovalStore();
approvals.request("run-1", "workflow");
if (approvals.isApproved("run-1") || !approvals.approve("run-1", "operator"))
  throw new Error("named workflow approval was not granted");
