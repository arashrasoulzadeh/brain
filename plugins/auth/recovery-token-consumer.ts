import { createHmac, timingSafeEqual } from "node:crypto";
import type { RecoveryToken } from "./recovery-token-issuer.js";

export class RecoveryTokenConsumer {
  constructor(private readonly tokens: ReadonlyMap<string, RecoveryToken>, private readonly secret: string) {}
  consume(projectId: string, destination: string, token: string, now: number): boolean { const record = this.tokens.get(`${projectId}:${destination}`); if (!record || record.used || record.expiresAt <= now) return false; const digest = createHmac("sha256", this.secret).update(token).digest("hex"); if (!timingSafeEqual(Buffer.from(record.digest), Buffer.from(digest))) return false; (record as RecoveryToken).used = true; return true; }
}
