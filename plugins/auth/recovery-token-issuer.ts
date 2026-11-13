import { createHmac, randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export type RecoveryToken = { digest: string; expiresAt: number; used: boolean };
export class RecoveryTokenIssuer {
  constructor(private readonly tokens: Map<string, RecoveryToken>, private readonly secret: string, private readonly inbox: { deliver(destination: string, body: string): void }) {}
  issue(projectId: string, destination: string, now: number, ttlSeconds: number): void { validateProjectId(projectId); if (!destination.includes("@") || ttlSeconds <= 0) throw new Error("invalid recovery request"); const token = randomBytes(24).toString("base64url"); const digest = createHmac("sha256", this.secret).update(token).digest("hex"); this.tokens.set(`${projectId}:${destination}`, { digest, expiresAt: now + ttlSeconds * 1000, used: false }); this.inbox.deliver(destination, token); }
}
