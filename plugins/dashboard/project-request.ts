import { pluginRequest, type RpcRequest } from "../sdk-typescript/src/index.js";

export function projectCreateRequest(pluginId: string, version: string, id: string): RpcRequest {
  return pluginRequest(
    "dashboard",
    "operator",
    pluginId,
    version,
    "project.create",
    { id },
    `project-${id}`,
  );
}

export function pluginPromoteRequest(
  projectId: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, "plugin.promote", pluginId, version);
}

export function pluginRollbackRequest(
  projectId: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, "plugin.rollback", pluginId, version);
}

export function pluginDisablePreviewRequest(
  projectId: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, "plugin.disable.preview", pluginId, version);
}

export function pluginDisableApplyRequest(
  projectId: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, "plugin.disable.apply", pluginId, version);
}

export function pluginLifecycleRequest(
  projectId: string,
  operation: "plugin.drain" | "plugin.resume" | "plugin.stop",
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, operation, pluginId, version);
}

export function pluginUninstallRequest(
  projectId: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return adminRequest(projectId, "plugin.uninstall", pluginId, version);
}

function adminRequest(
  projectId: string,
  operation: string,
  pluginId: string,
  version: string,
): RpcRequest {
  return pluginRequest(
    "dashboard",
    projectId,
    "dashboard",
    "0.1.0",
    operation,
    { targetPlugin: pluginId, targetVersion: version },
    `dashboard-${operation}-${projectId}-${pluginId}-${version}`,
  );
}
