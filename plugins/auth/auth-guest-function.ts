import type { PluginFunction } from "../sdk-typescript/src/index.js";
import type { AuthOperations } from "./auth-functions.js";
export function authGuestFunction(service: AuthOperations): PluginFunction {
  return (rpc) => service.guest(rpc.params.projectId);
}
