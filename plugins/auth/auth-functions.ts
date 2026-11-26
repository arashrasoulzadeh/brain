import { createRequestDispatcher, type PluginFunction } from "../sdk-typescript/src/index.js";
import { createAuthOperations } from "./create-auth-operations.js";

export type AuthOperations = ReturnType<typeof createAuthOperations>;

// One service instance per worker; project identity comes from its validated RPC request.
export const dispatchRequest = createRequestDispatcher(functions());

export function functions(
  service: AuthOperations = createAuthOperations(),
): Readonly<Record<string, PluginFunction>> {
  return {
    signup: (rpc) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.signup(rpc.params.projectId, payload?.username ?? "", payload?.password ?? "");
    },
    signin: (rpc) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.signin(
        rpc.params.projectId,
        payload?.username ?? payload?.provider ?? "",
        payload?.password ?? String(payload?.credential ?? ""),
      );
    },
    session: (rpc) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.session(rpc.params.projectId, payload?.token ?? "");
    },
    guest: (rpc) => service.guest(rpc.params.projectId),
  };
}
