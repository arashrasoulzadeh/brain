import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "auth",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["auth.signin", "auth.session"],
  extensionPoints: ["auth.provider"],
  configSchema: { registration: "string" },
} as const;

export function signInRequest(projectId: string, provider: string, credential: unknown) {
  return request(`auth-signin-${projectId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation: "auth.signin", payload: { provider, credential } });
}

export const handle: PluginHandler = (rpc) => {
  if (rpc.params.operation !== "auth.signin") throw new Error("unsupported auth operation");
  return { accepted: true, operation: rpc.params.operation };
};
