import { randomBytes } from "node:crypto";
import { sessionTokenDigest } from "./session-token-digest.js";

export class OperatorSessionIssuer {
  constructor(private readonly sessions: Map<string, string>) {}
  issue(operatorId: string): string { if (!operatorId.trim()) throw new Error("operator id required"); const token = randomBytes(32).toString("hex"); this.sessions.set(sessionTokenDigest(token), operatorId); return token; }
}
