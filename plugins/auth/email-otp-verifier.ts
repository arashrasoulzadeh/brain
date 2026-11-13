import { createHmac, timingSafeEqual } from "node:crypto";
import type { OTPChallenge } from "./otp-types.js";

export class EmailOTPVerifier {
  constructor(private readonly challenges: ReadonlyMap<string, OTPChallenge>, private readonly secret: string) {}
  verify(projectId: string, destination: string, purpose: string, code: string, now: number): boolean { const challenge = this.challenges.get(`${projectId}:${destination}:${purpose}`); const digest = createHmac("sha256", this.secret).update(code).digest("hex"); if (!challenge || challenge.used || challenge.expiresAt <= now || !timingSafeEqual(Buffer.from(challenge.digest), Buffer.from(digest))) return false; (challenge as OTPChallenge).used = true; return true; }
}
