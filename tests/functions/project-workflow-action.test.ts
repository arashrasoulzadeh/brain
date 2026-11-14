import {
  registerProjectWorkflowAction,
  authorizeProjectWorkflowAction,
  replaceProjectWorkflowStep,
} from "../../plugins/functions/index.js";

const actions = new Map();
const grants = new Set<string>();
const action = registerProjectWorkflowAction(actions, {
  projectId: "dev",
  id: "notify-owner",
  operation: "notify.owner",
  capability: "workflow.project",
});

if (action.projectId !== "dev" || !actions.has("dev:notify-owner")) {
  throw new Error("project workflow action was not registered");
}

grants.add("dev:notify-owner:workflow.project");
authorizeProjectWorkflowAction(actions, grants, "dev", "notify-owner");
grants.delete("dev:notify-owner:workflow.project");
try {
  authorizeProjectWorkflowAction(actions, grants, "dev", "notify-owner");
  throw new Error("revoked workflow grant was accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "workflow grant revoked") throw error;
}

for (const candidate of [
  {
    projectId: "dev",
    id: "notify-owner",
    operation: "notify.owner",
    capability: "workflow.project",
  },
  { projectId: "prod/dev", id: "other", operation: "notify.owner", capability: "workflow.project" },
  { projectId: "dev", id: "deploy", operation: "deploy.release", capability: "workflow.project" },
]) {
  try {
    registerProjectWorkflowAction(actions, candidate);
    throw new Error("invalid project workflow action was registered");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "invalid project workflow action")
      throw error;
  }
}

const replaced = replaceProjectWorkflowStep(actions, "dev", "notify-owner", {
  projectId: "dev",
  id: "notify-owner",
  operation: "notify.admin",
  capability: "workflow.project",
});
if (replaced.operation !== "notify.admin") throw new Error("workflow step was not replaced");

for (const [projectId, stepId] of [
  ["prod", "notify-owner"],
  ["dev", "missing"],
] as const) {
  try {
    replaceProjectWorkflowStep(actions, projectId, stepId, replaced);
    throw new Error("invalid workflow step replacement was accepted");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "invalid workflow step replacement")
      throw error;
  }
}
