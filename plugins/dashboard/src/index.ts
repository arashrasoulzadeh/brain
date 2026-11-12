import { request, type PluginHandler, type RpcRequest } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "dashboard",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["project.manage", "plugin.install", "plugin.grant", "plugin.enable"],
  extensionPoints: ["admin.navigation", "admin.page"],
  configSchema: { title: "string" },
} as const;

export function projectCreateRequest(id: string): RpcRequest {
  return request(`dashboard-project-${id}`, {
    protocol: 1, projectId: "operator", pluginId: manifest.id, version: manifest.version,
    operation: "project.create", payload: { id },
  });
}

export function adminRequest(projectId: string, operation: string, payload: unknown): RpcRequest {
  return request(`dashboard-${operation}-${projectId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation, payload });
}

export class DashboardService {
  private readonly projects = new Set<string>();
  private readonly installations = new Set<string>();
  private readonly grants = new Set<string>();

  createProject(projectId: string) {
    if (!projectId.trim()) throw new Error("project id required");
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error("invalid project id");
    if (this.projects.has(projectId)) throw new Error("project already exists");
    this.projects.add(projectId);
    return { projectId };
  }

  hasProject(projectId: string) { return this.projects.has(projectId); }

  installPlugin(projectId: string, pluginId: string, version: string) {
    if (!this.projects.has(projectId)) throw new Error("unknown project");
    if (!pluginId.trim() || !version.trim()) throw new Error("plugin identity required");
    const key = `${projectId}:${pluginId}@${version}`;
    if (this.installations.has(key)) throw new Error("plugin already installed");
    this.installations.add(key);
    return { projectId, pluginId, version, enabled: false };
  }

  isInstalled(projectId: string, pluginId: string, version: string) { return this.installations.has(`${projectId}:${pluginId}@${version}`); }

  grantCapability(projectId: string, pluginId: string, version: string, capability: string) {
    if (!this.isInstalled(projectId, pluginId, version)) throw new Error("plugin not installed");
    if (!capability.trim()) throw new Error("capability required");
    this.grants.add(`${projectId}:${pluginId}@${version}:${capability}`);
    return { projectId, pluginId, version, capability };
  }

  hasGrant(projectId: string, pluginId: string, version: string, capability: string) { return this.grants.has(`${projectId}:${pluginId}@${version}:${capability}`); }
}

export function createHandler(service = new DashboardService()): PluginHandler {
 return (rpc) => {
  const supported = ["project.create", "plugin.install", "plugin.grant", "plugin.enable"];
  if (!supported.includes(rpc.params.operation)) throw new Error("unsupported dashboard operation");
  if (rpc.params.operation === "project.create") {
    const payload = rpc.params.payload as { id?: string } | undefined;
    return { accepted: true, operation: rpc.params.operation, ...service.createProject(payload?.id ?? "") };
  }
  if (rpc.params.operation === "plugin.install" || rpc.params.operation === "plugin.grant") {
    const payload = rpc.params.payload as { targetPlugin?: string; targetVersion?: string; capability?: string } | undefined;
    const pluginId = payload?.targetPlugin ?? "";
    const version = payload?.targetVersion ?? "";
    if (rpc.params.operation === "plugin.install") return { accepted: true, operation: rpc.params.operation, ...service.installPlugin(rpc.params.projectId, pluginId, version) };
    return { accepted: true, operation: rpc.params.operation, ...service.grantCapability(rpc.params.projectId, pluginId, version, payload?.capability ?? "") };
  }
  return { accepted: true, operation: rpc.params.operation, payload: rpc.params.payload ?? null };
 };
}

export const handle = createHandler();
