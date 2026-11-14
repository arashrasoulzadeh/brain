import type { AuthSettings } from "./types.js";

export function authSettingsPage(
  projectId: string,
  settings: AuthSettings,
): { projectId: string; title: string; settings: AuthSettings } {
  if (!projectId || settings.oidcProviders.some((provider) => !provider))
    throw new Error("invalid auth settings");
  return {
    projectId,
    title: "Authentication settings",
    settings: { ...settings, oidcProviders: [...settings.oidcProviders] },
  };
}
