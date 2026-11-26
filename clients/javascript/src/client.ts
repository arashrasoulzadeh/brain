export type ClientTransport = <Response>(
  operation: string,
  projectId: string,
  payload: unknown,
) => Promise<Response>;

export type SignInResponse = { token: string; username: string };
export type SessionResponse = { username: string };

export class JavaScriptClient {
  constructor(
    private readonly projectId: string,
    private readonly transport: ClientTransport,
  ) {
    if (!projectId.trim()) throw new Error("client project required");
    if (!/^[a-zA-Z0-9]/.test(projectId) || /[^a-zA-Z0-9_-]/.test(projectId)) {
      throw new Error("invalid client project ID");
    }
  }

  signIn(username: string, password: string): Promise<SignInResponse> {
    return this.transport<SignInResponse>("auth.signin", this.projectId, { username, password });
  }

  session(token: string): Promise<SessionResponse> {
    return this.transport<SessionResponse>("auth.session", this.projectId, { token });
  }
}
