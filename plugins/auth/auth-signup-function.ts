import type { PluginFunction } from "../sdk-typescript/src/index.js";
import type { AuthOperations } from "./auth-functions.js";
export function authSignupFunction(service: AuthOperations): PluginFunction {
  return (rpc) => {
    const payload = rpc.params.payload as Record<string, string> | undefined;
    return service.signup(rpc.params.projectId, payload?.username ?? "", payload?.password ?? "");
  };
}
