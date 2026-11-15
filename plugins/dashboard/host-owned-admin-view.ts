import { validateProjectId } from "../sdk-typescript/src/index.js";

export function hostOwnedAdminView(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return {
    kind: "host-admin-view",
    projectId,
    title: "Project administration",
    trusted: true,
  } as const;
}
