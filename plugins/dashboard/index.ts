import {
  request,
  type RpcRequest,
} from "../sdk-typescript/src/index.js";

export function projectCreateRequest(pluginId: string, version: string, id: string): RpcRequest {
  return request(`dashboard-project-${id}`, {
    protocol: 1,
    projectId: "operator",
    pluginId,
    version,
    operation: "project.create",
    payload: { id },
  });
}

export function adminRequest(
  projectId: string,
  pluginId: string,
  version: string,
  operation: string,
  payload: unknown,
): RpcRequest {
  return request(`dashboard-${operation}-${projectId}`, {
    protocol: 1,
    projectId,
    pluginId,
    version,
    operation,
    payload,
  });
}

export type AuthSettings = { passwordEnabled: boolean; oidcProviders: readonly string[]; emailRecoveryEnabled: boolean };
export function authSettingsPage(projectId: string, settings: AuthSettings): { projectId: string; title: string; settings: AuthSettings } {
  if (!projectId || settings.oidcProviders.some((provider) => !provider)) throw new Error("invalid auth settings");
  return { projectId, title: "Authentication settings", settings: { ...settings, oidcProviders: [...settings.oidcProviders] } };
}

