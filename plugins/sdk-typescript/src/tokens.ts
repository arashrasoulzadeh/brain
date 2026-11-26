import { createHash, createHmac, randomBytes, randomInt, timingSafeEqual } from "node:crypto";

export type ExpiringToken = { expiresAt: number; used?: boolean };
export type DigestToken = { digest: string; expiresAt: number; used: boolean };

export function expiresAt(now: number, ttlMs: number, error: string): number {
  const expiry = now + ttlMs;
  if (!Number.isFinite(now) || !Number.isFinite(ttlMs) || ttlMs <= 0 || !Number.isFinite(expiry))
    throw new Error(error);
  return expiry;
}

export function isActiveToken<T extends ExpiringToken>(record: T | undefined, now: number): record is T {
  return !!record && !record.used && Number.isFinite(now) && Number.isFinite(record.expiresAt) && record.expiresAt > now;
}

export function consumeOnce<T extends ExpiringToken & { used: boolean }>(
  record: T | undefined, now: number, matches: (record: T) => boolean,
): T | undefined {
  if (!isActiveToken(record, now) || !matches(record)) return undefined;
  record.used = true;
  return record;
}

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
