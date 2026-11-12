import { randomBytes, scryptSync, timingSafeEqual } from "node:crypto";
import { request, type PluginFunction, type RpcRequest } from "../sdk-typescript/src/index.js";

export function signInRequest(projectId: string, pluginId: string, version: string, provider: string, credential: unknown) {
  return request(`auth-signin-${projectId}`, { protocol: 1, projectId, pluginId, version, operation: "auth.signin", payload: { provider, credential } });
}

export function signUpRequest(projectId: string, pluginId: string, version: string, username: string, password: string) {
  return request(`auth-signup-${projectId}-${username}`, { protocol: 1, projectId, pluginId, version, operation: "auth.signup", payload: { username, password } });
}

export function guestRequest(projectId: string, pluginId: string, version: string) {
  return request(`auth-guest-${projectId}`, { protocol: 1, projectId, pluginId, version, operation: "auth.guest", payload: {} });
}

type Account = { username: string; salt: Buffer; passwordHash: Buffer };

export class AuthService {
  private readonly accounts = new Map<string, Map<string, Account>>();
  private readonly sessions = new Map<string, { projectId: string; username: string }>();

  signup(projectId: string, username: string, password: string) {
    this.validate(projectId, username, password);
    const project = this.accounts.get(projectId) ?? new Map<string, Account>();
    if (project.has(username)) throw new Error("account already exists");
    const salt = randomBytes(16);
    project.set(username, { username, salt, passwordHash: this.hash(password, salt) });
    this.accounts.set(projectId, project);
    return { username };
  }

  signin(projectId: string, username: string, password: string) {
    const account = this.accounts.get(projectId)?.get(username);
    if (!account || !timingSafeEqual(account.passwordHash, this.hash(password, account.salt))) throw new Error("invalid credentials");
    const token = randomBytes(32).toString("hex");
    this.sessions.set(token, { projectId, username });
    return { token, username };
  }

  session(projectId: string, token: string) {
    const session = this.sessions.get(token);
    if (!session || session.projectId !== projectId) throw new Error("invalid session");
    return { username: session.username };
  }

  guest(projectId: string) {
    if (!projectId) throw new Error("project id required");
    const token = randomBytes(32).toString("hex");
    const username = `guest-${randomBytes(8).toString("hex")}`;
    this.sessions.set(token, { projectId, username });
    return { token, username, guest: true };
  }

  private hash(password: string, salt: Buffer) { return scryptSync(password, salt, 32); }
  private validate(projectId: string, username: string, password: string) {
    if (!projectId || !username || password.length < 8) throw new Error("invalid signup");
  }
}

export function functions(service = new AuthService()): Readonly<Record<string, PluginFunction>> {
  const payload = (rpc: RpcRequest) => rpc.params.payload as Record<string, string> | undefined;
  return {
    signup: (rpc) => { const p = payload(rpc); return service.signup(rpc.params.projectId, p?.username ?? "", p?.password ?? ""); },
    signin: (rpc) => { const p = payload(rpc); return service.signin(rpc.params.projectId, p?.username ?? p?.provider ?? "", p?.password ?? String(p?.credential ?? "")); },
    session: (rpc) => { const p = payload(rpc); return service.session(rpc.params.projectId, p?.token ?? ""); },
    guest: (rpc) => service.guest(rpc.params.projectId),
  };
}
