export function projectOverviewContribution(projectId: string) {
  if (!projectId.trim()) throw new Error("project identity required");
  return {
    view: "project.overview",
    projectId,
    title: "Project overview",
    replace: false,
  } as const;
}
