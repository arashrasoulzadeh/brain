import { randomBytes, scryptSync, timingSafeEqual } from "node:crypto";
import { request, type PluginHandler } from "../../sdk-typescript/src/index.js";

export const manifest = {
  id: "auth",
  version: "0.1.0",
  hostApi: 1,
  capabilities: ["auth.signup", "auth.signin", "auth.session"],
  extensionPoints: ["auth.provider"],
  configSchema: { registration: "string" },
} as const;

export function signInRequest(projectId: string, provider: string, credential: unknown) {
  return request(`auth-signin-${projectId}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation: "auth.signin", payload: { provider, credential } });
}

export function signUpRequest(projectId: string, username: string, password: string) {
  return request(`auth-signup-${projectId}-${username}`, { protocol: 1, projectId, pluginId: manifest.id, version: manifest.version, operation: "auth.signup", payload: { username, password } });
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

  private hash(password: string, salt: Buffer) { return scryptSync(password, salt, 32); }
  private validate(projectId: string, username: string, password: string) {
    if (!projectId || !username || password.length < 8) throw new Error("invalid signup");
  }
}

export function createHandler(service = new AuthService()): PluginHandler {
  return (rpc) => {
    const payload = rpc.params.payload as Record<string, string> | undefined;
    switch (rpc.params.operation) {
      case "auth.signup": return { accepted: true, operation: rpc.params.operation, ...service.signup(rpc.params.projectId, payload?.username ?? "", payload?.password ?? "") };
      case "auth.signin": return { accepted: true, operation: rpc.params.operation, ...service.signin(rpc.params.projectId, payload?.username ?? payload?.provider ?? "", payload?.password ?? String(payload?.credential ?? "")) };
      case "auth.session": return { accepted: true, operation: rpc.params.operation, ...service.session(rpc.params.projectId, payload?.token ?? "") };
      default: throw new Error("unsupported auth operation");
    }
  };
}

export const handle = createHandler();
