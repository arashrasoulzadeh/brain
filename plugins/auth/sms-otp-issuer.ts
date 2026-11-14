import { createHmac, randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { OTPChallenge } from "./otp-types.js";

export class SMSOTPIssuer {
  constructor(
    private readonly challenges: Map<string, OTPChallenge>,
    private readonly secret: string,
    private readonly inbox: { deliver(destination: string, body: string): void },
  ) {}
  issue(
    projectId: string,
    destination: string,
    purpose: string,
    now: number,
    ttlSeconds: number,
  ): void {
    validateProjectId(projectId);
    if (!/^\+[1-9][0-9]{7,14}$/.test(destination) || !purpose || ttlSeconds <= 0)
      throw new Error("invalid SMS OTP");
    const code = String(100000 + (randomBytes(4).readUInt32BE(0) % 900000));
    const digest = createHmac("sha256", this.secret).update(code).digest("hex");
    this.challenges.set(`${projectId}:${destination}:${purpose}`, {
      digest,
      expiresAt: now + ttlSeconds * 1000,
      used: false,
    });
    this.inbox.deliver(destination, code);
  }
}
