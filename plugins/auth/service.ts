import { createRequestDispatcher, type PluginFunction } from "../sdk-typescript/src/index.js";
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
      const result = await service.signup(context?.project.id ?? rpc.params.projectId, payload?.username ?? "", payload?.password ?? "");
      context?.plugins.observe.fire("accountCreated", {});
      return result;
    },
    signin: async (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      const result = await service.signin(
        context?.project.id ?? rpc.params.projectId,
        payload?.username ?? payload?.provider ?? "",
        payload?.password ?? String(payload?.credential ?? ""),
      );
      context?.plugins.observe.fire("signedIn", {});
      return result;
    },
    session: (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.session(context?.project.id ?? rpc.params.projectId, payload?.token ?? "");
    },
    guest: (rpc, context) => {
      const result = service.guest(context?.project.id ?? rpc.params.projectId);
      context?.plugins.observe.fire("guestSessionCreated", {});
      return result;
    },
  };
}
