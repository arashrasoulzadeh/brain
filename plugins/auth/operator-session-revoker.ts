import { sessionTokenDigest } from "./session-token-digest.js";

export class OperatorSessionRevoker {
  constructor(private readonly sessions: Map<string, string>, private readonly revoked: Set<string>) {}
  revoke(token: string): void { const digest = sessionTokenDigest(token); this.sessions.delete(digest); this.revoked.add(digest); }
}
