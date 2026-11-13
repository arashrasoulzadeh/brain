import { sessionTokenDigest } from "./session-token-digest.js";

export class OperatorSessionValidator {
  constructor(private readonly sessions: ReadonlyMap<string, string>, private readonly revoked: ReadonlySet<string>) {}
  valid(token: string, operatorId: string): boolean { const digest = sessionTokenDigest(token); return !this.revoked.has(digest) && this.sessions.get(digest) === operatorId; }
}
