import { timingSafeEqual } from "node:crypto";
import { createPKCEChallenge } from "./create-pkce-challenge.js";
export function verifyPKCE(verifier: string, challenge: string, method: string): boolean {
  if (method !== "S256") return false;
  const expected = Buffer.from(createPKCEChallenge(verifier));
  const received = Buffer.from(challenge);
  return expected.length === received.length && timingSafeEqual(expected, received);
}
