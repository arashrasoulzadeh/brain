import { validateProjectId } from "../sdk-typescript/src/validate-project-id.js";

export function customAdminView(projectId: string) {
  validateProjectId(projectId);
  return {
    contentType: "text/html",
    projectId,
    html: `<main data-project="${projectId}"><h1>Custom admin view</h1></main>`,
  } as const;
}
