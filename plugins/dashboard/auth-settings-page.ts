import type { AuthSettings } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export function authSettingsPage(
  projectId: string,
  settings: AuthSettings,
): { projectId: string; title: string; settings: AuthSettings } {
  try {
    validateProjectId(projectId);
  } catch {
    throw new Error("invalid auth settings");
  }
  if (settings.oidcProviders.some((provider) => !provider))
    throw new Error("invalid auth settings");
  return {
    projectId,
    title: "Authentication settings",
    settings: { ...settings, oidcProviders: [...settings.oidcProviders] },
  };
}
