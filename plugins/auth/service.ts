import { createRequestDispatcher, requireOperator, requireProjectInput, validateProjectId, type PluginFunction, type PluginContext, type RpcRequest } from "../sdk-typescript/src/index.js";
import { createAuthState, type AuthState } from "./state.js";
import { AccountSignup, AccountSignin, projectUserID } from "./accounts.js";
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
    identity: session.identity.bind(session),
    guest: guest.guest.bind(guest),
    listAccounts: (projectId: string, after: string) => listAccounts(state, projectId, after),
    revokeSessions: (projectId: string, username: string) => revokeSessions(state, projectId, username),
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
    identity: (rpc, context) => {
      const payload = rpc.params.payload as Record<string, string> | undefined;
      return service.identity(projectOf(rpc, context), payload?.token ?? "");
    },
    listAccounts: (rpc, context) => {
      const operator = requireOperator(context, "auth.admin.read");
      const payload = requireProjectInput(operator, rpc.params.payload);
      const after = payload.after ?? "";
      if (typeof after !== "string" || after.length > 1000) throw new Error("invalid account cursor");
      return service.listAccounts(projectOf(rpc, operator), after);
    },
    revokeSessions: (rpc, context) => {
      if (!context) throw new Error("operator authorization required");
      const payload = requireProjectInput(context, rpc.params.payload);
      if (typeof payload.username !== "string") throw new Error("invalid username");
      const username = validateProjectId(payload.username, "invalid username");
      requireOperator(context, "auth.admin.manage", username);
      return service.revokeSessions(projectOf(rpc, context), username);
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

/** Project-scoped metadata only: no password hashes, session tokens or digests. */
function listAccounts(state: AuthState, projectId: string, after: string) {
  const users = new Map<string, { username: string; userId: string; guest: boolean; activeSessions: number }>();
  for (const account of state.accounts.get(projectId)?.values() ?? []) {
    users.set(account.username, { username: account.username, userId: account.userId, guest: false, activeSessions: 0 });
  }
  const now = Date.now();
  for (const session of state.sessions.values()) {
    if (session.projectId !== projectId || session.expiresAt <= now) continue;
    const user = users.get(session.username);
    if (user) user.activeSessions++;
    else if (session.username.startsWith("guest-")) users.set(session.username, {
      username: session.username, userId: projectUserID(projectId, session.username), guest: true, activeSessions: 1,
    });
  }
  const matches = [...users.values()].filter(user => user.username > after)
    .sort((a, b) => a.username < b.username ? -1 : a.username > b.username ? 1 : 0);
  const rows = matches.slice(0, 50);
  return { rows, next: matches.length > 50 ? rows[49].username : "" };
}

function revokeSessions(state: AuthState, projectId: string, username: string) {
  let revoked = 0;
  for (const [digest, session] of state.sessions) {
    if (session.projectId === projectId && session.username === username) {
      state.sessions.delete(digest);
      revoked++;
    }
  }
  return { username, revoked };
}
