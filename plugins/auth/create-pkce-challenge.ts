import { createHash } from "node:crypto";
export function createPKCEChallenge(verifier: string): string {
  return createHash("sha256").update(verifier).digest("base64url");
}
