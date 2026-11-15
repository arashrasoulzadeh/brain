import { validateProjectId } from "../sdk-typescript/src/index.js";

export function projectOverviewContribution(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return {
    view: "project.overview",
    projectId,
    title: "Project overview",
    replace: false,
  } as const;
}
