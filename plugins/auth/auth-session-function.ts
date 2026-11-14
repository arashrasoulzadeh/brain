import type { PluginFunction } from "../sdk-typescript/src/index.js";
import type { AuthOperations } from "./auth-functions.js";
export function authSessionFunction(service: AuthOperations): PluginFunction {
  return (rpc) => {
    const payload = rpc.params.payload as Record<string, string> | undefined;
    return service.session(rpc.params.projectId, payload?.token ?? "");
  };
}
