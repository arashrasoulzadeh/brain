export function firstPartySessionCookie(token: string, maxAgeSeconds: number): string {
  if (!token || !Number.isSafeInteger(maxAgeSeconds) || maxAgeSeconds <= 0)
    throw new Error("invalid session cookie");
  if (/[;\s]/.test(token)) throw new Error("invalid session cookie");
  return `__Host-brain_session=${token}; Max-Age=${maxAgeSeconds}; Path=/; HttpOnly; Secure; SameSite=Lax`;
}
