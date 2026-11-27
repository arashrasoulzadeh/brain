import { createRequestDispatcher, requireOperator, requireProjectInput, validateProjectId, type PluginFunction, type PluginContext, type RpcRequest } from "../sdk-typescript/src/index.js";
import { createAuthState, type Account, type AuthSession, type AuthState } from "./state.js";
import { AccountSignup, AccountSignin, projectUserID } from "./accounts.js";
import { AccountSessionReader, GuestSessionIssuer, sessionTokenDigest } from "./sessions.js";

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

// Standalone SDK fixtures retain their local service. Production requests carry
// a host-owned, operation-scoped snapshot; the registry rejects caller snapshots.
const transientDispatch = createRequestDispatcher(functions());

type DurableAuthInput = {
  account?: Account;
  session?: AuthSession;
  page?: { rows: { username: string; userId: string; guest: boolean; activeSessions: number }[]; next: string };
};

export function dispatchRequest(name: string, rpc: RpcRequest, context?: PluginContext): unknown {
  const payload = rpc.params.payload;
  if (!payload || typeof payload !== "object" || !Object.hasOwn(payload, "_brainAuth"))
    return transientDispatch(name, rpc, context);
  return dispatchDurable(name, rpc, context);
}

async function dispatchDurable(name: string, rpc: RpcRequest, context?: PluginContext): Promise<unknown> {
  if (!context) throw new Error("host project context required");
  const payload = requireProjectInput(context, rpc.params.payload);
  const project = projectOf(rpc, context);
  const stored = payload._brainAuth as DurableAuthInput;
  if (!stored || typeof stored !== "object" || Array.isArray(stored)) throw new Error("invalid stored authentication state");
  if (name === "listAccounts") {
    requireOperator(context, "auth.admin.read");
    if (!stored.page || !Array.isArray(stored.page.rows) || typeof stored.page.next !== "string")
      throw new Error("invalid account page");
    return stored.page;
  }
  if (name === "revokeSessions") {
    const username = validateProjectId(String(payload.username ?? ""), "invalid username");
    requireOperator(context, "auth.admin.manage", username);
    // The host computes the actual affected-row count in its transaction.
    context.plugins.observe.fire("sessions.revoked", {});
    return { username, revoked: 0 };
  }
  const state = createAuthState();
  if (stored.account) {
    if (stored.account.userId !== project + ":" + stored.account.username) throw new Error("invalid account state");
    state.accounts.set(project, new Map([[stored.account.username, stored.account]]));
  }
  if (stored.session) {
    if (stored.session.projectId !== project || typeof payload.token !== "string") throw new Error("invalid session state");
    state.sessions.set(sessionTokenDigest(payload.token), stored.session);
  }
  const result = await createRequestDispatcher(functions(createAuthOperations(state)))(name, rpc, context);
  if (name === "signup") {
    const account = state.accounts.get(project)?.get(String(payload.username));
    if (!account) throw new Error("account state missing");
    return { result, passwordHash: account.passwordHash };
  }
  if (name === "signin" || name === "guest") {
    const token = (result as { token?: string }).token;
    const session = typeof token === "string" ? state.sessions.get(sessionTokenDigest(token)) : undefined;
    if (!session) throw new Error("session state missing");
    return { result, expiresAt: session.expiresAt };
  }
  return result;
}

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
      const result = service.revokeSessions(projectOf(rpc, context), username);
      context.plugins.observe.fire("sessions.revoked", {});
      return result;
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
