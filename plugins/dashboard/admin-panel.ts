import { validateProjectId } from "../sdk-typescript/src/index.js";

export type AdminPanelSection =
  "projects" | "plugins" | "database" | "functions" | "workflows" | "files" | "live-sessions";

export function adminPanel(projectId: string): {
  projectId: string;
  title: string;
  sections: readonly AdminPanelSection[];
  destructiveConfirmationRequired: true;
} {
  validateProjectId(projectId, "invalid admin panel project");
  return {
    projectId,
    title: "Brain administration",
    sections: [
      "projects",
      "plugins",
      "database",
      "functions",
      "workflows",
      "files",
      "live-sessions",
    ],
    destructiveConfirmationRequired: true,
  };
}
