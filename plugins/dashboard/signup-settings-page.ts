import type { SignupSettings } from "./types.js";

export function signupSettingsPage(
  projectId: string,
  settings: SignupSettings,
): { projectId: string; title: string; settings: SignupSettings } {
  if (!projectId || settings.allowedDomains.some((domain) => !domain || domain.includes("/")))
    throw new Error("invalid signup settings");
  return {
    projectId,
    title: "Signup settings",
    settings: { ...settings, allowedDomains: [...settings.allowedDomains] },
  };
}
