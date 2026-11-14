import { registerProjectWorkflowAction } from "../../plugins/functions/index.js";

const actions = new Map();
const action = registerProjectWorkflowAction(actions, {
  projectId: "dev",
  id: "notify-owner",
  operation: "notify.owner",
  capability: "workflow.project",
});

if (action.projectId !== "dev" || !actions.has("dev:notify-owner")) {
  throw new Error("project workflow action was not registered");
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
