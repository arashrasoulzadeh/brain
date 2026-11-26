import { createRequestDispatcher, type PluginFunction, type PluginContext, type RpcRequest } from "../sdk-typescript/src/index.js";
import { createAuthState, type AuthState } from "./state.js";
import { AccountSignup, AccountSignin } from "./accounts.js";
import { AccountSessionReader, GuestSessionIssuer } from "./sessions.js";

export function createAuthOperations(state: AuthState = createAuthState()) {
  const signup = new AccountSignup(state);
  const signin = new AccountSignin(state);
  const session = new AccountSessionReader(state);
  const guest = new GuestSessionIssuer(state);
  return {
    signup: signup.signup.bind(signup),
    signin: signin.signin.bind(signin),
    session: session.session.bind(session),
    guest: guest.guest.bind(guest),
  };
}

export type AuthOperations = ReturnType<typeof createAuthOperations>;

// One service instance per worker; project identity comes from its validated RPC request.
export const dispatchRequest = createRequestDispatcher(functions());

export function functions(
  service: AuthOperations = createAuthOperations(),
): Readonly<Record<string, PluginFunction>> {
  return {
    signup: async (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      const result = await service.signup(projectOf(rpc, context), payload?.username ?? "", payload?.password ?? "");
      context?.plugins.observe.fire("account.created", { created: true });
      return result;
    },
    signin: async (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      const result = await service.signin(
        projectOf(rpc, context),
        payload?.username ?? payload?.provider ?? "",
        payload?.password ?? String(payload?.credential ?? ""),
      );
      context?.plugins.observe.fire("session.established", { guest: false });
      return result;
    },
    session: (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.session(projectOf(rpc, context), payload?.token ?? "");
    },
    guest: (rpc, context) => {
      const result = service.guest(projectOf(rpc, context));
      context?.plugins.observe.fire("session.established", { guest: true });
      return result;
    },
  };
}

function projectOf(rpc: RpcRequest, context?: PluginContext): string {
  if (context && context.project.id !== rpc.params.projectId) throw new Error("project context mismatch");
  return context?.project.id ?? rpc.params.projectId;
}
