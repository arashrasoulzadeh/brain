import type { AuthorizationCode } from "./authorization-code-types.js";

export class AuthorizationCodeConsumer {
  constructor(private readonly codes: Map<string, AuthorizationCode>) {}
  consume(code: string, projectId: string, clientId: string, redirectUri: string, now: number): { userId: string } { const record = this.codes.get(code); if (!record || record.used || record.expiresAt <= now || record.projectId !== projectId || record.clientId !== clientId || record.redirectUri !== redirectUri) throw new Error("authorization code unavailable"); record.used = true; return { userId: record.userId }; }
}
