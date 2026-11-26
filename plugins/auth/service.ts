import { createRequestDispatcher, pluginRequest, type PluginFunction } from "../sdk-typescript/src/index.js";
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

export function guestRequest(projectId: string, pluginId: string, version: string) {
  return pluginRequest("auth", projectId, pluginId, version, "auth.guest", {}, "guest");
}

export function signInRequest(
  projectId: string,
  pluginId: string,
  version: string,
  provider: string,
  credential: unknown,
) {
  return pluginRequest(
    "auth",
    projectId,
    pluginId,
    version,
    "auth.signin",
    { provider, credential },
    "signin",
  );
}

export function signUpRequest(
  projectId: string,
  pluginId: string,
  version: string,
  username: string,
  password: string,
) {
  return pluginRequest(
    "auth",
    projectId,
    pluginId,
    version,
    "auth.signup",
    { username, password },
    `signup-${username}`,
  );
}
