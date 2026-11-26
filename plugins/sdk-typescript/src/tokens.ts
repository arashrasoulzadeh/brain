import { createHash, createHmac, randomBytes, randomInt, timingSafeEqual } from "node:crypto";

import { consumeOnce, expiresAt } from "./validation.js";

export type DigestToken = { digest: string; expiresAt: number; used: boolean };

export function randomToken(bytes = 24, encoding: "hex" | "base64url" = "base64url"): string {
  if (!Number.isSafeInteger(bytes) || bytes < 1 || bytes > 256) throw new Error("invalid token size");
  return randomBytes(bytes).toString(encoding);
}

export function numericOTP(): string {
  return String(randomInt(100000, 1000000));
}

export function tokenDigest(token: string): string {
  return createHash("sha256").update(token).digest("hex");
}

export function secretsEqual(left: string, right: string): boolean {
  const a = Buffer.from(left);
  const b = Buffer.from(right);
  return a.length === b.length && timingSafeEqual(a, b);
}

function keyedDigest(secret: string, token: string): string {
  return createHmac("sha256", secret).update(token).digest("hex");
}

export function createDigestToken(secret: string, token: string, now: number, ttlMs: number, error: string): DigestToken {
  return { digest: keyedDigest(secret, token), expiresAt: expiresAt(now, ttlMs, error), used: false };
}

export function consumeDigestToken(record: DigestToken | undefined, secret: string, token: string, now: number): boolean {
  return !!consumeOnce(record, now, (candidate) => secretsEqual(candidate.digest, keyedDigest(secret, token)));
}
