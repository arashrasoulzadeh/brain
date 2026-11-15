import { validateProjectId } from "../sdk-typescript/src/index.js";

export function browserQuickstart(
  projectId: string,
  apiOrigin: string,
): { projectId: string; apiOrigin: string; steps: readonly string[] } {
  const origin = new URL(apiOrigin);
  try {
    validateProjectId(projectId);
  } catch {
    throw new Error("invalid browser quickstart");
  }
  if (!["http:", "https:"].includes(origin.protocol)) throw new Error("invalid browser quickstart");
  return {
    projectId,
    apiOrigin: origin.origin,
    steps: ["create-session", "load-project", "open-dashboard"],
  };
}
