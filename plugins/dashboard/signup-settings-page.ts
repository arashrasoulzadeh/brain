import type { SignupSettings } from "./types.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export function signupSettingsPage(
  projectId: string,
  settings: SignupSettings,
): { projectId: string; title: string; settings: SignupSettings } {
  try {
    validateProjectId(projectId);
  } catch {
    throw new Error("invalid signup settings");
  }
  if (settings.allowedDomains.some((domain) => !domain || domain.includes("/")))
    throw new Error("invalid signup settings");
  return {
    projectId,
    title: "Signup settings",
    settings: { ...settings, allowedDomains: [...settings.allowedDomains] },
  };
}
